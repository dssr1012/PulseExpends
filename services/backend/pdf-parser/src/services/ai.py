import logging
import json
import asyncio
from typing import Dict, List, Any, Optional
from datetime import datetime
import openai
from openai import AsyncOpenAI

from ..config import settings

logger = logging.getLogger(__name__)

class AIService:
    """Service for AI-powered document parsing and analysis"""
    
    def __init__(self):
        self.client = None
        self._initialize_client()
    
    def _initialize_client(self):
        """Initialize OpenAI client"""
        if settings.OPENAI_API_KEY:
            self.client = AsyncOpenAI(api_key=settings.OPENAI_API_KEY)
            logger.info("OpenAI client initialized")
        else:
            logger.warning("OPENAI_API_KEY not set, AI features will be limited")
            self.client = None
    
    async def parse_document(
        self, 
        text: str, 
        document_type: str = "credit_card",
        bank_name: Optional[str] = None,
        language: str = "es"
    ) -> Dict[str, Any]:
        """Parse document text using AI to extract structured data"""
        logger.info(f"Parsing {document_type} document with AI")
        
        if not self.client:
            logger.warning("OpenAI client not available, using fallback parsing")
            return self._fallback_parse(text, document_type, bank_name)
        
        try:
            # Prepare system prompt based on document type
            system_prompt = self._get_system_prompt(document_type, bank_name, language)
            
            # Prepare user prompt with document text
            user_prompt = self._get_user_prompt(text, document_type, language)
            
            # Call OpenAI API
            response = await self.client.chat.completions.create(
                model=settings.OPENAI_MODEL,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ],
                temperature=settings.OPENAI_TEMPERATURE,
                max_tokens=settings.OPENAI_MAX_TOKENS,
                response_format={"type": "json_object"}
            )
            
            # Parse response
            result_text = response.choices[0].message.content
            result = json.loads(result_text)
            
            logger.info(f"AI parsing completed successfully")
            return result
            
        except Exception as e:
            logger.error(f"AI parsing failed: {str(e)}", exc_info=True)
            # Fall back to basic parsing
            return self._fallback_parse(text, document_type, bank_name)
    
    async def extract_installment_info(self, description: str) -> Optional[Dict[str, Any]]:
        """Extract installment plan information from transaction description"""
        if not self.client:
            return None
        
        try:
            system_prompt = """You are a financial data extraction assistant. Extract installment plan information from transaction descriptions.
            Return a JSON object with the following fields if found:
            - total_installments: total number of installments (integer)
            - current_installment: current installment number (integer)
            - installment_amount: amount of this installment (float)
            - total_amount: total amount of the purchase (float)
            - interest_rate: interest rate if mentioned (float, optional)
            - remaining_installments: remaining installments (integer, optional)
            
            If no installment information is found, return an empty JSON object {}."""
            
            user_prompt = f"""Extract installment information from this transaction description:
            
            Description: {description}
            
            Return only the JSON object."""
            
            response = await self.client.chat.completions.create(
                model=settings.OPENAI_MODEL,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ],
                temperature=0.1,
                max_tokens=500,
                response_format={"type": "json_object"}
            )
            
            result_text = response.choices[0].message.content
            result = json.loads(result_text)
            
            # Return None if empty
            if not result:
                return None
            
            return result
            
        except Exception as e:
            logger.warning(f"Failed to extract installment info: {str(e)}")
            return None
    
    async def categorize_transaction(
        self, 
        description: str, 
        amount: float, 
        merchant: str = ""
    ) -> Dict[str, Any]:
        """Categorize transaction using AI"""
        if not self.client:
            return self._basic_categorization(description, amount, merchant)
        
        try:
            system_prompt = """You are a financial categorization assistant. Categorize transactions based on description, amount, and merchant.
            Return a JSON object with:
            - category: primary category (string)
            - subcategory: optional subcategory (string)
            - confidence: confidence score 0-1 (float)
            - tags: array of relevant tags (array of strings)
            - is_business_expense: boolean
            - is_personal_expense: boolean
            - notes: any additional notes (string, optional)
            
            Categories should be in Spanish for Chilean transactions. Common categories:
            - supermercado
            - restaurante
            - transporte
            - servicios (luz, agua, internet, etc.)
            - salud
            - educación
            - entretenimiento
            - ropa
            - tecnología
            - hogar
            - viajes
            - seguros
            - impuestos
            - otros"""
            
            user_prompt = f"""Categorize this transaction:
            
            Description: {description}
            Amount: {amount}
            Merchant: {merchant}
            
            Return only the JSON object."""
            
            response = await self.client.chat.completions.create(
                model=settings.OPENAI_MODEL,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ],
                temperature=0.2,
                max_tokens=300,
                response_format={"type": "json_object"}
            )
            
            result_text = response.choices[0].message.content
            result = json.loads(result_text)
            
            return result
            
        except Exception as e:
            logger.warning(f"AI categorization failed: {str(e)}")
            return self._basic_categorization(description, amount, merchant)
    
    async def detect_anomalies(
        self, 
        transactions: List[Dict[str, Any]], 
        historical_data: Optional[List[Dict[str, Any]]] = None
    ) -> List[Dict[str, Any]]:
        """Detect anomalies in transactions using AI"""
        if not self.client:
            return []
        
        try:
            system_prompt = """You are a financial fraud and anomaly detection assistant. Analyze transactions for anomalies.
            Return a JSON array of anomalies, each with:
            - transaction_id: reference to the transaction
            - anomaly_type: type of anomaly (unusual_amount, unusual_merchant, unusual_time, duplicate, etc.)
            - severity: low, medium, high
            - description: explanation of the anomaly
            - confidence: confidence score 0-1 (float)
            - suggested_action: what to do about it
            
            Only return anomalies you're confident about (confidence > 0.7)."""
            
            # Prepare transaction data
            tx_data = []
            for tx in transactions:
                tx_data.append({
                    "id": tx.get("id", ""),
                    "date": tx.get("date", ""),
                    "description": tx.get("description", ""),
                    "amount": tx.get("amount", 0),
                    "merchant": tx.get("merchant", ""),
                    "category": tx.get("category", ""),
                })
            
            user_prompt = f"""Analyze these transactions for anomalies:
            
            Current Transactions: {json.dumps(tx_data, ensure_ascii=False)}
            
            Historical Data: {json.dumps(historical_data or [], ensure_ascii=False)}
            
            Return only the JSON array of anomalies."""
            
            response = await self.client.chat.completions.create(
                model=settings.OPENAI_MODEL,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ],
                temperature=0.1,
                max_tokens=1000,
                response_format={"type": "json_object"}
            )
            
            result_text = response.choices[0].message.content
            result = json.loads(result_text)
            
            return result.get("anomalies", []) if isinstance(result, dict) else result
            
        except Exception as e:
            logger.warning(f"Anomaly detection failed: {str(e)}")
            return []
    
    async def summarize_statement(
        self, 
        transactions: List[Dict[str, Any]], 
        statement_period: str
    ) -> Dict[str, Any]:
        """Generate a summary of a credit card statement using AI"""
        if not self.client:
            return self._basic_summary(transactions, statement_period)
        
        try:
            system_prompt = """You are a financial summary assistant. Create a comprehensive summary of a credit card statement.
            Return a JSON object with:
            - period: statement period
            - total_charges: total amount charged
            - total_payments: total payments made
            - net_balance: net balance
            - top_categories: array of top 5 spending categories with amounts
            - largest_transaction: details of largest transaction
            - recurring_charges: array of recurring charges
            - unusual_spending: any unusual spending patterns
            - savings_opportunities: potential savings opportunities
            - summary_text: human-readable summary text in Spanish"""
            
            user_prompt = f"""Summarize this credit card statement:
            
            Statement Period: {statement_period}
            Transactions: {json.dumps(transactions, ensure_ascii=False)}
            
            Return only the JSON object."""
            
            response = await self.client.chat.completions.create(
                model=settings.OPENAI_MODEL,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ],
                temperature=0.3,
                max_tokens=800,
                response_format={"type": "json_object"}
            )
            
            result_text = response.choices[0].message.content
            result = json.loads(result_text)
            
            return result
            
        except Exception as e:
            logger.warning(f"Statement summarization failed: {str(e)}")
            return self._basic_summary(transactions, statement_period)
    
    def _get_system_prompt(self, document_type: str, bank_name: Optional[str], language: str) -> str:
        """Get system prompt for document parsing"""
        prompts = {
            "credit_card": f"""You are a financial document parsing assistant specializing in credit card statements.
            Extract all transactions and relevant information from the credit card statement.
            
            Return a JSON object with the following structure:
            {{
                "metadata": {{
                    "bank": "bank name",
                    "cardholder": "cardholder name",
                    "card_number": "card number (masked)",
                    "statement_period": "statement period",
                    "due_date": "payment due date",
                    "total_amount": "total amount due",
                    "minimum_payment": "minimum payment",
                    "currency": "currency code"
                }},
                "transactions": [
                    {{
                        "date": "transaction date",
                        "description": "transaction description",
                        "amount": "transaction amount (negative for charges)",
                        "currency": "currency",
                        "merchant": "merchant name if identifiable",
                        "location": "location if available",
                        "reference": "reference number if available",
                        "is_installment": "boolean indicating if this is an installment",
                        "installment_details": {{
                            "current": "current installment number",
                            "total": "total installments",
                            "amount": "installment amount"
                        }} if applicable
                    }}
                ],
                "summary": {{
                    "total_charges": "total charges amount",
                    "total_payments": "total payments amount",
                    "categories": {{
                        "category_name": "amount"
                    }}
                }}
            }}
            
            Parse the document carefully. For Chilean banks like {bank_name or 'the bank'}, pay attention to:
            - Dates in DD/MM/YYYY format
            - Chilean pesos (CLP) as currency
            - Common Chilean merchants and locations
            - Installment indicators like "CUOTA", "SIN INTERES", etc.
            
            Return only the JSON object, no additional text.""",
            
            "bank_statement": """You are a financial document parsing assistant specializing in bank statements.
            Extract all transactions and relevant information from the bank statement.
            
            Return a JSON object with the following structure:
            {
                "metadata": {
                    "bank": "bank name",
                    "account_holder": "account holder name",
                    "account_number": "account number (masked)",
                    "statement_period": "statement period",
                    "opening_balance": "opening balance",
                    "closing_balance": "closing balance",
                    "currency": "currency code"
                },
                "transactions": [
                    {
                        "date": "transaction date",
                        "description": "transaction description",
                        "amount": "transaction amount (positive for deposits, negative for withdrawals)",
                        "currency": "currency",
                        "type": "transaction type (deposit, withdrawal, transfer, fee, etc.)",
                        "balance_after": "balance after transaction if available",
                        "reference": "reference number if available"
                    }
                ],
                "summary": {
                    "total_deposits": "total deposits amount",
                    "total_withdrawals": "total withdrawals amount",
                    "fees": "total fees amount"
                }
            }
            
            Return only the JSON object, no additional text.""",
            
            "invoice": """You are a document parsing assistant specializing in invoices.
            Extract all relevant information from the invoice.
            
            Return a JSON object with the following structure:
            {
                "metadata": {
                    "invoice_number": "invoice number",
                    "issue_date": "invoice issue date",
                    "due_date": "payment due date",
                    "supplier": "supplier name",
                    "customer": "customer name",
                    "total_amount": "total invoice amount",
                    "currency": "currency code",
                    "tax_amount": "tax amount",
                    "subtotal": "subtotal amount"
                },
                "line_items": [
                    {
                        "description": "item description",
                        "quantity": "quantity",
                        "unit_price": "unit price",
                        "total": "line total",
                        "tax_rate": "tax rate if applicable"
                    }
                ],
                "payment_terms": "payment terms if specified",
                "notes": "any additional notes"
            }
            
            Return only the JSON object, no additional text.""",
            
            "receipt": """You are a document parsing assistant specializing in receipts.
            Extract all relevant information from the receipt.
            
            Return a JSON object with the following structure:
            {
                "metadata": {
                    "merchant": "merchant name",
                    "transaction_date": "transaction date and time",
                    "receipt_number": "receipt number",
                    "total_amount": "total amount",
                    "currency": "currency code",
                    "payment_method": "payment method",
                    "tax_amount": "tax amount",
                    "subtotal": "subtotal amount"
                },
                "items": [
                    {
                        "description": "item description",
                        "quantity": "quantity",
                        "unit_price": "unit price",
                        "total": "line total"
                    }
                ],
                "tax_breakdown": {
                    "tax_name": "tax amount"
                }
            }
            
            Return only the JSON object, no additional text."""
        }
        
        return prompts.get(document_type, prompts["credit_card"])
    
    def _get_user_prompt(self, text: str, document_type: str, language: str) -> str:
        """Get user prompt for document parsing"""
        # Truncate text if too long (respect token limits)
        max_length = 12000  # Leave room for prompts and response
        if len(text) > max_length:
            text = text[:max_length] + "\n\n[Document truncated due to length]"
        
        return f"""Parse this {document_type} document (language: {language}):

{document_text}

Return only the JSON object as specified in the system prompt.""".replace("{document_text}", text)
    
    def _fallback_parse(self, text: str, document_type: str, bank_name: Optional[str]) -> Dict[str, Any]:
        """Fallback parsing when AI is not available"""
        logger.info("Using fallback parsing")
        
        # Simple regex-based parsing for common patterns
        import re
        
        result = {
            "metadata": {
                "bank": bank_name or "unknown",
                "cardholder": "",
                "card_number": "",
                "statement_period": "",
                "due_date": "",
                "total_amount": 0,
                "minimum_payment": 0,
                "currency": "CLP"
            },
            "transactions": [],
            "summary": {
                "total_charges": 0,
                "total_payments": 0,
                "categories": {}
            }
        }
        
        # Extract common patterns
        lines = text.split('\n')
        
        for line in lines:
            # Look for date patterns followed by description and amount
            date_pattern = r'(\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4})'
            amount_pattern = r'([\$\€\£]?\s*\d{1,3}(?:\.\d{3})*(?:,\d{2})?)'
            
            date_match = re.search(date_pattern, line)
            amount_matches = re.findall(amount_pattern, line)
            
            if date_match and len(amount_matches) >= 1:
                date = date_match.group(1)
                amount_str = amount_matches[-1]  # Usually last amount in line
                
                # Extract description (text between date and amount)
                date_end = date_match.end()
                amount_start = line.rfind(amount_str)
                
                if amount_start > date_end:
                    description = line[date_end:amount_start].strip()
                    
                    # Clean amount
                    amount = self._parse_amount_fallback(amount_str)
                    
                    # Simple categorization
                    category = self._categorize_fallback(description)
                    
                    transaction = {
                        "date": date,
                        "description": description,
                        "amount": amount,
                        "currency": "CLP",
                        "merchant": self._extract_merchant_fallback(description),
                        "location": "",
                        "reference": "",
                        "is_installment": "CUOTA" in description.upper() or "INSTALMENT" in description.upper(),
                        "installment_details": {}
                    }
                    
                    result["transactions"].append(transaction)
                    
                    # Update summary
                    if amount < 0:
                        result["summary"]["total_charges"] += abs(amount)
                    else:
                        result["summary"]["total_payments"] += amount
                    
                    result["summary"]["categories"][category] = result["summary"]["categories"].get(category, 0) + abs(amount)
        
        return result
    
    def _parse_amount_fallback(self, amount_str: str) -> float:
        """Parse amount string in fallback mode"""
        try:
            # Remove currency symbols and thousand separators
            cleaned = amount_str.replace('$', '').replace('€', '').replace('£', '').replace('.', '').replace(',', '.')
            # Remove any non-numeric characters except minus and dot
            cleaned = ''.join(c for c in cleaned if c.isdigit() or c in '-.')
            return float(cleaned)
        except:
            return 0.0
    
    def _categorize_fallback(self, description: str) -> str:
        """Categorize transaction in fallback mode"""
        desc_lower = description.lower()
        
        categories = {
            "supermercado": ["lider", "jumbo", "santa isabel", "unimarc", "tottus", "super", "mercado"],
            "restaurante": ["restaurant", "restaurante", "comida", "café", "cafe", "bar", "pub", "pizza", "hamburguesa"],
            "transporte": ["uber", "didi", "cabify", "taxi", "metro", "transantiago", "bip", "combustible", "estacionamiento", "shell", "copec"],
            "servicios": ["agua", "luz", "electricidad", "gas", "internet", "telefonía", "claro", "entel", "movistar", "wom", "vtr"],
            "salud": ["farmacia", "hospital", "clínica", "doctor", "médico", "isapre", "fonasa", "cruz verde", "ahumada"],
            "educación": ["colegio", "universidad", "curso", "libro", "material", "utiles"],
            "entretenimiento": ["cine", "netflix", "spotify", "youtube", "disney", "hbo", "parque", "concierto", "teatro"],
            "ropa": ["ropa", "zapatos", "tienda", "mall", "falabella", "ripley", "paris", "h&m", "zara"],
            "tecnología": ["computador", "celular", "tablet", "televisor", "electrónica", "samsung", "apple", "iphone"],
            "hogar": ["mueble", "decoración", "cocina", "baño", "herramienta", "ferretería", "homecenter", "sodimac"],
        }
        
        for category, keywords in categories.items():
            for keyword in keywords:
                if keyword in desc_lower:
                    return category
        
        return "otros"
    
    def _extract_merchant_fallback(self, description: str) -> str:
        """Extract merchant name in fallback mode"""
        # Simple extraction - take first few words that look like a merchant name
        words = description.split()
        if len(words) >= 2:
            # Skip common prefixes
            skip_words = ["COMPRA", "PAGO", "RETIRO", "DEPOSITO", "TRANSFERENCIA", "EN", "DE", "LA", "EL"]
            merchant_words = [w for w in words[:3] if w.upper() not in skip_words]
            if merchant_words:
                return ' '.join(merchant_words)
        
        return description[:30]  # Return first 30 chars if no better extraction
    
    def _basic_categorization(self, description: str, amount: float, merchant: str) -> Dict[str, Any]:
        """Basic transaction categorization without AI"""
        category = self._categorize_fallback(description)
        
        return {
            "category": category,
            "subcategory": "",
            "confidence": 0.6,
            "tags": [],
            "is_business_expense": False,
            "is_personal_expense": True,
            "notes": "Categorized using fallback method"
        }
    
    def _basic_summary(self, transactions: List[Dict[str, Any]], statement_period: str) -> Dict[str, Any]:
        """Basic statement summary without AI"""
        total_charges = sum(tx.get("amount", 0) for tx in transactions if tx.get("amount", 0) < 0)
        total_payments = sum(tx.get("amount", 0) for tx in transactions if tx.get("amount", 0) > 0)
        
        # Group by category
        categories = {}
        for tx in transactions:
            category = tx.get("category", "otros")
            amount = abs(tx.get("amount", 0))
            categories[category] = categories.get(category, 0) + amount
        
        # Find largest transaction
        largest_tx = max(transactions, key=lambda x: abs(x.get("amount", 0))) if transactions else {}
        
        return {
            "period": statement_period,
            "total_charges": abs(total_charges),
            "total_payments": total_payments,
            "net_balance": total_charges + total_payments,
            "top_categories": sorted(categories.items(), key=lambda x: x[1], reverse=True)[:5],
            "largest_transaction": largest_tx,
            "recurring_charges": [],
            "unusual_spending": [],
            "savings_opportunities": [],
            "summary_text": f"Resumen básico para el período {statement_period}"
        }
    
    def is_ready(self) -> bool:
        """Check if AI service is ready"""
        return self.client is not None and settings.OPENAI_API_KEY is not None