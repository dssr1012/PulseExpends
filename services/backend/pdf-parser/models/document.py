from typing import List, Optional, Dict, Any
from datetime import datetime
from pydantic import BaseModel, Field, validator
from enum import Enum

class DocumentType(str, Enum):
    CREDIT_CARD = "credit_card"
    BANK_STATEMENT = "bank_statement"
    INVOICE = "invoice"
    RECEIPT = "receipt"
    OTHER = "other"

class DocumentStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"

class ParsedTransaction(BaseModel):
    """Represents a parsed transaction from a document"""
    id: str = Field(..., description="Unique transaction identifier")
    transaction_number: int = Field(..., description="Transaction number in the document")
    date: datetime = Field(..., description="Transaction date")
    description: str = Field(..., description="Transaction description")
    amount: float = Field(..., description="Transaction amount (negative for expenses, positive for income)")
    currency: str = Field(default="CLP", description="Currency code")
    category: Optional[str] = Field(None, description="Transaction category")
    subcategory: Optional[str] = Field(None, description="Transaction subcategory")
    payment_method: str = Field(default="credit_card", description="Payment method")
    merchant: Optional[str] = Field(None, description="Merchant name")
    location: Optional[str] = Field(None, description="Transaction location")
    reference_number: Optional[str] = Field(None, description="Reference number")
    is_installment: bool = Field(default=False, description="Whether this is an installment payment")
    installment_details: Optional[Dict[str, Any]] = Field(None, description="Installment details if applicable")
    tags: List[str] = Field(default_factory=list, description="Transaction tags")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")
    confidence: float = Field(default=1.0, ge=0.0, le=1.0, description="Confidence score for parsing accuracy")
    
    @validator('amount')
    def validate_amount(cls, v):
        """Validate amount is reasonable"""
        if abs(v) > 1000000000:  # 1 billion limit
            raise ValueError('Amount too large')
        return v
    
    @validator('currency')
    def validate_currency(cls, v):
        """Validate currency code"""
        if len(v) != 3:
            raise ValueError('Currency code must be 3 characters')
        return v.upper()

class Document(BaseModel):
    """Represents a parsed document"""
    id: str = Field(..., description="Unique document identifier")
    filename: str = Field(..., description="Original filename")
    document_type: DocumentType = Field(..., description="Type of document")
    bank_name: Optional[str] = Field(None, description="Bank name if applicable")
    language: str = Field(default="es", description="Document language")
    extracted_text: str = Field(..., description="Extracted text from document")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Document metadata")
    processing_time: float = Field(..., description="Processing time in seconds")
    created_at: datetime = Field(default_factory=datetime.utcnow, description="Creation timestamp")
    
    @validator('document_type')
    def validate_document_type(cls, v):
        """Validate document type"""
        if isinstance(v, str):
            try:
                return DocumentType(v)
            except ValueError:
                raise ValueError(f'Invalid document type: {v}')
        return v

class ParsingResult(BaseModel):
    """Represents the result of document parsing"""
    request_id: str = Field(..., description="Unique request identifier")
    status: DocumentStatus = Field(..., description="Parsing status")
    document: Optional[Document] = Field(None, description="Parsed document")
    transactions: List[ParsedTransaction] = Field(default_factory=list, description="Parsed transactions")
    summary: Dict[str, Any] = Field(default_factory=dict, description="Parsing summary")
    error_message: Optional[str] = Field(None, description="Error message if failed")
    processing_time: float = Field(..., description="Total processing time in seconds")
    created_at: datetime = Field(default_factory=datetime.utcnow, description="Creation timestamp")
    
    @validator('status')
    def validate_status(cls, v):
        """Validate document status"""
        if isinstance(v, str):
            try:
                return DocumentStatus(v)
            except ValueError:
                raise ValueError(f'Invalid document status: {v}')
        return v

class BankTemplate(BaseModel):
    """Template for bank-specific document parsing"""
    bank_code: str = Field(..., description="Bank identifier code")
    bank_name: str = Field(..., description="Bank name")
    country: str = Field(..., description="Country code")
    document_types: List[DocumentType] = Field(..., description="Supported document types")
    languages: List[str] = Field(..., description="Supported languages")
    
    # Parsing patterns
    header_patterns: List[str] = Field(default_factory=list, description="Patterns to identify bank in header")
    footer_patterns: List[str] = Field(default_factory=list, description="Patterns to identify bank in footer")
    
    # Field extraction patterns
    field_patterns: Dict[str, str] = Field(default_factory=dict, description="Regex patterns for field extraction")
    
    # Table detection
    table_headers: List[List[str]] = Field(default_factory=list, description="Expected table headers")
    table_formats: List[Dict[str, Any]] = Field(default_factory=list, description="Table format specifications")
    
    # Validation rules
    validation_rules: Dict[str, Any] = Field(default_factory=dict, description="Validation rules for extracted data")

class ProcessingJob(BaseModel):
    """Represents a document processing job"""
    job_id: str = Field(..., description="Unique job identifier")
    request_id: str = Field(..., description="Associated request identifier")
    status: str = Field(default="pending", description="Job status")
    document_key: str = Field(..., description="OBS document key")
    document_type: DocumentType = Field(..., description="Type of document")
    bank_name: Optional[str] = Field(None, description="Bank name if specified")
    language: str = Field(default="es", description="Document language")
    
    # Processing parameters
    extract_installments: bool = Field(default=True, description="Whether to extract installment information")
    validate_transactions: bool = Field(default=True, description="Whether to validate transactions")
    
    # Timing
    submitted_at: datetime = Field(default_factory=datetime.utcnow, description="Submission timestamp")
    started_at: Optional[datetime] = Field(None, description="Processing start timestamp")
    completed_at: Optional[datetime] = Field(None, description="Completion timestamp")
    
    # Results
    result_key: Optional[str] = Field(None, description="OBS result key")
    error_message: Optional[str] = Field(None, description="Error message if failed")
    
    # Metadata
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Job metadata")

class ValidationResult(BaseModel):
    """Represents the result of transaction validation"""
    transaction_id: str = Field(..., description="Transaction identifier")
    is_valid: bool = Field(..., description="Whether transaction is valid")
    validation_errors: List[str] = Field(default_factory=list, description="Validation errors")
    warnings: List[str] = Field(default_factory=list, description="Validation warnings")
    suggested_corrections: List[Dict[str, Any]] = Field(default_factory=list, description="Suggested corrections")
    confidence_score: float = Field(default=1.0, ge=0.0, le=1.0, description="Validation confidence")

class InstallmentPlan(BaseModel):
    """Represents an installment payment plan"""
    total_amount: float = Field(..., description="Total purchase amount")
    installment_amount: float = Field(..., description="Amount per installment")
    total_installments: int = Field(..., description="Total number of installments")
    current_installment: int = Field(..., description="Current installment number")
    remaining_installments: int = Field(..., description="Remaining installments")
    interest_rate: Optional[float] = Field(None, description="Interest rate if applicable")
    start_date: datetime = Field(..., description="Installment plan start date")
    end_date: Optional[datetime] = Field(None, description="Installment plan end date")
    merchant: Optional[str] = Field(None, description="Merchant name")
    description: Optional[str] = Field(None, description="Purchase description")

class ParsingStatistics(BaseModel):
    """Statistics about document parsing"""
    total_documents: int = Field(default=0, description="Total documents processed")
    successful_parses: int = Field(default=0, description="Successful parses")
    failed_parses: int = Field(default=0, description="Failed parses")
    average_processing_time: float = Field(default=0.0, description="Average processing time in seconds")
    
    # By document type
    by_document_type: Dict[str, int] = Field(default_factory=dict, description="Count by document type")
    
    # By bank
    by_bank: Dict[str, int] = Field(default_factory=dict, description="Count by bank")
    
    # By language
    by_language: Dict[str, int] = Field(default_factory=dict, description="Count by language")
    
    # Transaction statistics
    total_transactions: int = Field(default=0, description="Total transactions extracted")
    average_transactions_per_document: float = Field(default=0.0, description="Average transactions per document")
    
    # Error statistics
    common_errors: Dict[str, int] = Field(default_factory=dict, description="Common error types and counts")

class HealthStatus(BaseModel):
    """Health status of the parsing service"""
    status: str = Field(..., description="Overall status")
    version: str = Field(..., description="Service version")
    timestamp: datetime = Field(..., description="Check timestamp")
    
    # Component status
    ocr_service: str = Field(..., description="OCR service status")
    ai_service: str = Field(..., description="AI service status")
    obs_service: str = Field(..., description="OBS service status")
    
    # System metrics
    uptime: float = Field(..., description="Service uptime in seconds")
    memory_usage: float = Field(..., description="Memory usage percentage")
    cpu_usage: float = Field(..., description="CPU usage percentage")
    
    # Performance metrics
    requests_processed: int = Field(default=0, description="Total requests processed")
    average_response_time: float = Field(default=0.0, description="Average response time in seconds")
    
    # Last check
    last_successful_parse: Optional[datetime] = Field(None, description="Last successful parse timestamp")
    last_error: Optional[str] = Field(None, description="Last error message")