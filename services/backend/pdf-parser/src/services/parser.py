import logging
import asyncio
from typing import List, Optional, Dict, Any
from datetime import datetime
import uuid
import tempfile
import os

from ..models.document import Document, ParsedTransaction, ParsingResult
from .ocr import OCRService
from .ai import AIService

logger = logging.getLogger(__name__)

class PDFParserService:
    """Main service for parsing PDF and image documents"""
    
    def __init__(self):
        self.ocr_service = OCRService()
        self.ai_service = AIService()
        self.supported_banks = self._load_bank_templates()
        
    def _load_bank_templates(self) -> Dict[str, Dict[str, Any]]:
        """Load bank-specific parsing templates"""
        return {
            "bchile": {
                "name": "Banco de Chile",
                "country": "CL",
                "credit_card_patterns": [
                    r"BANCO DE CHILE",
                    r"TARJETA DE CRÉDITO",
                    r"Estado de Cuenta"
                ],
                "fields": {
                    "card_number": r"Número de Tarjeta:\s*(\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4})",
                    "cardholder": r"Titular:\s*(.+)",
                    "period": r"Periodo:\s*(\d{2}/\d{2}/\d{4})\s*al\s*(\d{2}/\d{2}/\d{4})",
                    "due_date": r"Fecha de Vencimiento:\s*(\d{2}/\d{2}/\d{4})",
                    "total_amount": r"Total a Pagar:\s*\$?\s*([\d.,]+)",
                    "minimum_payment": r"Pago Mínimo:\s*\$?\s*([\d.,]+)"
                }
            },
            "santander": {
                "name": "Santander",
                "country": "CL",
                "credit_card_patterns": [
                    r"SANTANDER",
                    r"Estado de Cuenta",
                    r"Resumen de Tarjeta de Crédito"
                ],
                "fields": {
                    "card_number": r"Tarjeta N°:\s*(\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4})",
                    "cardholder": r"Señor\(a\):\s*(.+)",
                    "period": r"Periodo del\s*(\d{2}/\d{2}/\d{4})\s*al\s*(\d{2}/\d{2}/\d{4})",
                    "due_date": r"Fecha de Vencimiento:\s*(\d{2}/\d{2}/\d{4})",
                    "total_amount": r"Total a Pagar:\s*\$?\s*([\d.,]+)",
                    "minimum_payment": r"Pago Mínimo:\s*\$?\s*([\d.,]+)"
                }
            },
            # Add more bank templates as needed
        }
    
    async def parse_document(
        self,
        content: bytes,
        filename: str,
        document_type: str = "credit_card",
        bank_name: Optional[str] = None,
        language: str = "es",
        extract_installments: bool = True,
        validate_transactions: bool = True
    ) -> ParsingResult:
        """Parse a document and extract transaction data"""
        request_id = str(uuid.uuid4())
        logger.info(f"Starting document parsing: {filename}, request_id: {request_id}")
        
        start_time = datetime.utcnow()
        
        try:
            # Create temporary file
            with tempfile.NamedTemporaryFile(suffix=os.path.splitext(filename)[1], delete=False) as tmp_file:
                tmp_file.write(content)
                tmp_path = tmp_file.name
            
            try:
                # Extract text from document
                logger.info(f"Extracting text from document: {filename}")
                extracted_text = await self._extract_text(tmp_path, language)
                
                # Detect bank if not specified
                if not bank_name:
                    bank_name = self._detect_bank(extracted_text)
                    logger.info(f"Detected bank: {bank_name}")
                
                # Parse document with AI
                logger.info(f"Parsing document with AI: {document_type}")
                parsed_data = await self.ai_service.parse_document(
                    text=extracted_text,
                    document_type=document_type,
                    bank_name=bank_name,
                    language=language
                )
                
                # Extract transactions
                transactions = await self._extract_transactions(
                    parsed_data=parsed_data,
                    document_type=document_type,
                    extract_installments=extract_installments
                )
                
                # Validate transactions if requested
                if validate_transactions:
                    transactions = await self._validate_transactions(transactions)
                
                # Create document object
                document = Document(
                    id=request_id,
                    filename=filename,
                    document_type=document_type,
                    bank_name=bank_name,
                    language=language,
                    extracted_text=extracted_text,
                    metadata=parsed_data.get("metadata", {}),
                    processing_time=(datetime.utcnow() - start_time).total_seconds()
                )
                
                # Calculate summary
                summary = self._calculate_summary(transactions)
                
                processing_time = (datetime.utcnow() - start_time).total_seconds()
                
                result = ParsingResult(
                    request_id=request_id,
                    status="completed",
                    document=document,
                    transactions=transactions,
                    summary=summary,
                    processing_time=processing_time,
                    created_at=datetime.utcnow()
                )
                
                logger.info(f"Document parsing completed: {filename}, transactions: {len(transactions)}")
                return result
                
            finally:
                # Clean up temporary file
                os.unlink(tmp_path)
                
        except Exception as e:
            logger.error(f"error parsing document {filename}: {str(e)}", exc_info=True)
            processing_time = (datetime.utcnow() - start_time).total_seconds()
            
            return ParsingResult(
                request_id=request_id,
                status="failed",
                error_message=str(e),
                processing_time=processing_time,
                created_at=datetime.utcnow()
            )
    
    async def _extract_text(self, file_path: str, language: str) -> str:
        """Extract text from document using OCR"""
        file_ext = os.path.splitext(file_path)[1].lower()
        
        if file_ext == '.pdf':
            # Use PDF-specific text extraction
            text = await self.ocr_service.extract_from_pdf(file_path, language)
        elif file_ext in ['.jpg', '.jpeg', '.png', '.tiff', '.tif', '.bmp']:
            # Use image OCR
            text = await self.ocr_service.extract_from_image(file_path, language)
        else:
            raise ValueError(f"Unsupported file format: {file_ext}")
        
        return text
    
    def _detect_bank(self, text: str) -> str:
        """Detect bank from extracted text"""
        text_lower = text.lower()
        
        for bank_code, bank_info in self.supported_banks.items():
            for pattern in bank_info.get("credit_card_patterns", []):
                if pattern.lower() in text_lower:
                    return bank_code
        
        # Default to generic if no bank detected
        return "generic"
    
    async def _extract_transactions(
        self,
        parsed_data: Dict[str, Any],
        document_type: str,
        extract_installments: bool = True
    ) -> List[ParsedTransaction]:
        """Extract transactions from parsed data"""
        transactions = []
        
        # Get transactions from parsed data
        raw_transactions = parsed_data.get("transactions", [])
        
        for idx, raw_tx in enumerate(raw_transactions):
            try:
                transaction = ParsedTransaction(
                    id=str(uuid.uuid4()),
                    transaction_number=idx + 1,
                    date=self._parse_date(raw_tx.get("date")),
                    description=raw_tx.get("description", ""),
                    amount=self._parse_amount(raw_tx.get("amount")),
                    currency=raw_tx.get("currency", "CLP"),
                    category=self._categorize_transaction(
                        raw_tx.get("description", ""),
                        raw_tx.get("amount", 0)
                    ),
                    payment_method="credit_card",  # Default for credit card statements
                    merchant=raw_tx.get("merchant", ""),
                    location=raw_tx.get("location", ""),
                    reference_number=raw_tx.get("reference", ""),
                    is_installment=raw_tx.get("is_installment", False),
                    installment_details=raw_tx.get("installment_details"),
                    metadata={
                        "source_document": document_type,
                        "parsed_fields": list(raw_tx.keys())
                    }
                )
                
                # Extract installment information if requested
                if extract_installments and transaction.is_installment:
                    transaction = await self._extract_installment_details(transaction)
                
                transactions.append(transaction)
                
            except Exception as e:
                logger.warning(f"Failed to parse transaction {idx}: {str(e)}")
                continue
        
        return transactions
    
    async def _extract_installment_details(self, transaction: ParsedTransaction) -> ParsedTransaction:
        """Extract installment plan details from transaction"""
        if not transaction.is_installment:
            return transaction
        
        try:
            # Use AI to extract installment details from description
            installment_info = await self.ai_service.extract_installment_info(
                transaction.description
            )
            
            if installment_info:
                transaction.installment_details = {
                    "total_installments": installment_info.get("total_installments"),
                    "current_installment": installment_info.get("current_installment"),
                    "installment_amount": installment_info.get("installment_amount"),
                    "total_amount": installment_info.get("total_amount"),
                    "interest_rate": installment_info.get("interest_rate"),
                    "remaining_installments": installment_info.get("remaining_installments"),
                }
        
        except Exception as e:
            logger.warning(f"Failed to extract installment details: {str(e)}")
        
        return transaction
    
    async def _validate_transactions(self, transactions: List[ParsedTransaction]) -> List[ParsedTransaction]:
        """Validate and clean transaction data"""
        validated_transactions = []
        
        for transaction in transactions:
            try:
                # Validate required fields
                if not transaction.date:
                    logger.warning(f"Transaction {transaction.id} missing date")
                    continue
                
                if not transaction.description or len(transaction.description.strip()) == 0:
                    logger.warning(f"Transaction {transaction.id} missing description")
                    continue
                
                if transaction.amount == 0:
                    logger.warning(f"Transaction {transaction.id} has zero amount")
                    continue
                
                # Clean description
                transaction.description = self._clean_description(transaction.description)
                
                # Standardize currency
                transaction.currency = transaction.currency.upper()
                
                # Add validation metadata
                transaction.metadata["validated"] = True
                transaction.metadata["validation_timestamp"] = datetime.utcnow().isoformat()
                
                validated_transactions.append(transaction)
                
            except Exception as e:
                logger.warning(f"Failed to validate transaction {transaction.id}: {str(e)}")
                continue
        
        return validated_transactions
    
    def _calculate_summary(self, transactions: List[ParsedTransaction]) -> Dict[str, Any]:
        """Calculate summary statistics from transactions"""
        if not transactions:
            return {
                "total_transactions": 0,
                "total_amount": 0,
                "average_amount": 0,
                "by_category": {},
                "by_merchant": {},
                "installment_count": 0,
                "installment_total": 0
            }
        
        total_amount = sum(tx.amount for tx in transactions)
        installment_transactions = [tx for tx in transactions if tx.is_installment]
        
        # Group by category
        by_category = {}
        for tx in transactions:
            category = tx.category or "uncategorized"
            by_category[category] = by_category.get(category, 0) + tx.amount
        
        # Group by merchant
        by_merchant = {}
        for tx in transactions:
            if tx.merchant:
                by_merchant[tx.merchant] = by_merchant.get(tx.merchant, 0) + tx.amount
        
        return {
            "total_transactions": len(transactions),
            "total_amount": total_amount,
            "average_amount": total_amount / len(transactions) if transactions else 0,
            "by_category": by_category,
            "by_merchant": by_merchant,
            "installment_count": len(installment_transactions),
            "installment_total": sum(tx.amount for tx in installment_transactions),
            "currency": transactions[0].currency if transactions else "CLP"
        }
    
    def _parse_date(self, date_str: str) -> Optional[datetime]:
        """Parse date string to datetime"""
        if not date_str:
            return None
        
        # Try common date formats
        date_formats = [
            "%d/%m/%Y",
            "%d-%m-%Y",
            "%Y-%m-%d",
            "%d/%m/%y",
            "%d-%m-%y",
            "%Y/%m/%d",
            "%d %b %Y",
            "%d %B %Y",
        ]
        
        for fmt in date_formats:
            try:
                return datetime.strptime(date_str.strip(), fmt)
            except ValueError:
                continue
        
        logger.warning(f"Could not parse date: {date_str}")
        return None
    
    def _parse_amount(self, amount_str: Any) -> float:
        """Parse amount string to float"""
        if isinstance(amount_str, (int, float)):
            return float(amount_str)
        
        if not isinstance(amount_str, str):
            return 0.0
        
        try:
            # Remove currency symbols, thousand separators, etc.
            cleaned = amount_str.replace("$", "").replace("€", "").replace("£", "").replace("¥", "")
            cleaned = cleaned.replace(".", "").replace(",", ".")
            cleaned = ''.join(c for c in cleaned if c.isdigit() or c in '.-')
            
            # Handle negative amounts
            is_negative = cleaned.startswith('-') or '(cr)' in amount_str.lower()
            amount = abs(float(cleaned))
            
            return -amount if is_negative else amount
            
        except (ValueError, AttributeError):
            logger.warning(f"Could not parse amount: {amount_str}")
            return 0.0
    
    def _categorize_transaction(self, description: str, amount: float) -> str:
        """Categorize transaction based on description and amount"""
        description_lower = description.lower()
        
        # Common categories for Chilean expenses
        categories = {
            "supermercado": ["lider", "jumbo", "santa isabel", "unimarc", "tottus", "super", "mercado"],
            "restaurante": ["restaurant", "restaurante", "comida", "café", "cafe", "bar", "pub"],
            "transporte": ["uber", "didi", "cabify", "taxi", "metro", "transantiago", "bip", "combustible", "estacionamiento"],
            "servicios": ["agua", "luz", "electricidad", "gas", "internet", "telefonía", "claro", "entel", "movistar", "wom"],
            "salud": ["farmacia", "hospital", "clínica", "doctor", "médico", "isapre", "fonasa"],
            "educación": ["colegio", "universidad", "curso", "libro", "material"],
            "entretenimiento": ["cine", "netflix", "spotify", "youtube", "disney", "hbo", "parque", "concierto"],
            "ropa": ["ropa", "zapatos", "tienda", "mall", "falabella", "ripley", "paris"],
            "tecnología": ["computador", "celular", "tablet", "televisor", "electrónica", "samsung", "apple"],
            "hogar": ["mueble", "decoración", "cocina", "baño", "herramienta", "ferretería"],
        }
        
        for category, keywords in categories.items():
            for keyword in keywords:
                if keyword in description_lower:
                    return category
        
        # Default categories based on amount
        if amount < 0:
            return "gasto"
        else:
            return "ingreso"
    
    def _clean_description(self, description: str) -> str:
        """Clean transaction description"""
        if not description:
            return ""
        
        # Remove extra whitespace
        cleaned = ' '.join(description.split())
        
        # Remove common prefixes/suffixes
        prefixes = ["COMPRA ", "PAGO ", "RETIRO ", "DEPOSITO ", "TRANSFERENCIA "]
        for prefix in prefixes:
            if cleaned.startswith(prefix):
                cleaned = cleaned[len(prefix):]
        
        # Capitalize first letter of each word
        cleaned = cleaned.title()
        
        return cleaned.strip()
    
    def is_ready(self) -> bool:
        """Check if service is ready"""
        return self.ocr_service.is_ready() and self.ai_service.is_ready()