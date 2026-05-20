from fastapi import FastAPI, HTTPException, UploadFile, File, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
import uvicorn
import logging
from datetime import datetime

from .config import settings
from .services.parser import PDFParserService
from .services.ocr import OCRService
from .services.ai import AIService
from .services.obs import OBSService
from .models.document import Document, ParsedTransaction, ParsingResult

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="PulseExpends PDF Parser API",
    description="AI-powered PDF/Image parsing service for credit card statements",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, restrict to specific origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize services
pdf_parser = PDFParserService()
ocr_service = OCRService()
ai_service = AIService()
obs_service = OBSService()

# Pydantic models for API
class ParseRequest(BaseModel):
    document_url: Optional[str] = None
    document_type: str = Field(..., description="Type of document: credit_card, bank_statement, invoice, receipt")
    bank_name: Optional[str] = None
    language: str = "es"
    extract_installments: bool = True
    validate_transactions: bool = True

class ParseResponse(BaseModel):
    request_id: str
    status: str
    document: Optional[Document] = None
    transactions: List[ParsedTransaction] = []
    summary: Dict[str, Any] = {}
    processing_time: float
    created_at: datetime

class HealthResponse(BaseModel):
    status: str
    version: str
    timestamp: datetime
    services: Dict[str, str]

@app.get("/")
async def root():
    return {
        "service": "PulseExpends PDF Parser",
        "version": "1.0.0",
        "status": "operational"
    }

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    services_status = {
        "pdf_parser": "healthy" if pdf_parser.is_ready() else "unhealthy",
        "ocr_service": "healthy" if ocr_service.is_ready() else "unhealthy",
        "ai_service": "healthy" if ai_service.is_ready() else "unhealthy",
        "obs_service": "healthy" if obs_service.is_ready() else "unhealthy",
    }
    
    overall_status = "healthy" if all(status == "healthy" for status in services_status.values()) else "unhealthy"
    
    return HealthResponse(
        status=overall_status,
        version="1.0.0",
        timestamp=datetime.utcnow(),
        services=services_status
    )

@app.post("/parse/upload", response_model=ParseResponse)
async def parse_upload(
    file: UploadFile = File(...),
    request: ParseRequest = None
):
    """Parse a PDF/Image file uploaded directly"""
    try:
        if request is None:
            request = ParseRequest(document_type="credit_card")
        
        # Read file content
        content = await file.read()
        filename = file.filename
        
        # Parse document
        start_time = datetime.utcnow()
        result = await pdf_parser.parse_document(
            content=content,
            filename=filename,
            document_type=request.document_type,
            bank_name=request.bank_name,
            language=request.language,
            extract_installments=request.extract_installments,
            validate_transactions=request.validate_transactions
        )
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        return ParseResponse(
            request_id=result.request_id,
            status="completed",
            document=result.document,
            transactions=result.transactions,
            summary=result.summary,
            processing_time=processing_time,
            created_at=datetime.utcnow()
        )
        
    except Exception as e:
        logger.error(f"Error parsing uploaded file: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Failed to parse document: {str(e)}")

@app.post("/parse/url", response_model=ParseResponse)
async def parse_url(
    request: ParseRequest,
    background_tasks: BackgroundTasks
):
    """Parse a document from a URL (async processing)"""
    try:
        if not request.document_url:
            raise HTTPException(status_code=400, detail="document_url is required")
        
        # Download document from URL
        document_content = await obs_service.download_from_url(request.document_url)
        
        # Parse document
        start_time = datetime.utcnow()
        result = await pdf_parser.parse_document(
            content=document_content,
            filename=request.document_url.split("/")[-1],
            document_type=request.document_type,
            bank_name=request.bank_name,
            language=request.language,
            extract_installments=request.extract_installments,
            validate_transactions=request.validate_transactions
        )
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        # Store result in OBS for later retrieval
        background_tasks.add_task(
            obs_service.store_parsing_result,
            result.request_id,
            result
        )
        
        return ParseResponse(
            request_id=result.request_id,
            status="completed",
            document=result.document,
            transactions=result.transactions,
            summary=result.summary,
            processing_time=processing_time,
            created_at=datetime.utcnow()
        )
        
    except Exception as e:
        logger.error(f"Error parsing document from URL: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Failed to parse document: {str(e)}")

@app.post("/parse/obs", response_model=ParseResponse)
async def parse_obs(
    request: ParseRequest,
    background_tasks: BackgroundTasks
):
    """Parse a document from OBS storage"""
    try:
        if not request.document_url:
            raise HTTPException(status_code=400, detail="document_url is required")
        
        # Extract OBS key from URL
        # Assuming URL format: obs://bucket-name/key/path
        if not request.document_url.startswith("obs://"):
            raise HTTPException(status_code=400, detail="Invalid OBS URL format. Use obs://bucket-name/key")
        
        parts = request.document_url[6:].split("/", 1)
        if len(parts) != 2:
            raise HTTPException(status_code=400, detail="Invalid OBS URL format")
        
        bucket_name, object_key = parts
        
        # Download document from OBS
        document_content = await obs_service.download_from_obs(bucket_name, object_key)
        
        # Parse document
        start_time = datetime.utcnow()
        result = await pdf_parser.parse_document(
            content=document_content,
            filename=object_key.split("/")[-1],
            document_type=request.document_type,
            bank_name=request.bank_name,
            language=request.language,
            extract_installments=request.extract_installments,
            validate_transactions=request.validate_transactions
        )
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        # Store result in OBS for later retrieval
        background_tasks.add_task(
            obs_service.store_parsing_result,
            result.request_id,
            result
        )
        
        return ParseResponse(
            request_id=result.request_id,
            status="completed",
            document=result.document,
            transactions=result.transactions,
            summary=result.summary,
            processing_time=processing_time,
            created_at=datetime.utcnow()
        )
        
    except Exception as e:
        logger.error(f"Error parsing document from OBS: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Failed to parse document: {str(e)}")

@app.get("/result/{request_id}")
async def get_result(request_id: str):
    """Get parsing result by request ID"""
    try:
        result = await obs_service.get_parsing_result(request_id)
        if not result:
            raise HTTPException(status_code=404, detail="Result not found")
        
        return ParseResponse(
            request_id=result.request_id,
            status="completed",
            document=result.document,
            transactions=result.transactions,
            summary=result.summary,
            processing_time=result.processing_time,
            created_at=result.created_at
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error retrieving result: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Failed to retrieve result: {str(e)}")

@app.get("/banks/supported")
async def get_supported_banks():
    """Get list of supported banks and their parsing templates"""
    return {
        "supported_banks": [
            {
                "name": "Banco de Chile",
                "code": "bchile",
                "country": "CL",
                "supported_documents": ["credit_card", "bank_statement"],
                "languages": ["es"]
            },
            {
                "name": "Banco Estado",
                "code": "bestado",
                "country": "CL",
                "supported_documents": ["credit_card", "bank_statement"],
                "languages": ["es"]
            },
            {
                "name": "Santander",
                "code": "santander",
                "country": "CL",
                "supported_documents": ["credit_card", "bank_statement"],
                "languages": ["es"]
            },
            {
                "name": "BCI",
                "code": "bci",
                "country": "CL",
                "supported_documents": ["credit_card", "bank_statement"],
                "languages": ["es"]
            },
            {
                "name": "Scotiabank",
                "code": "scotiabank",
                "country": "CL",
                "supported_documents": ["credit_card", "bank_statement"],
                "languages": ["es"]
            },
            {
                "name": "Itaú",
                "code": "itau",
                "country": "CL",
                "supported_documents": ["credit_card", "bank_statement"],
                "languages": ["es"]
            },
            {
                "name": "Falabella",
                "code": "falabella",
                "country": "CL",
                "supported_documents": ["credit_card"],
                "languages": ["es"]
            },
            {
                "name": "Ripley",
                "code": "ripley",
                "country": "CL",
                "supported_documents": ["credit_card"],
                "languages": ["es"]
            },
            {
                "name": "Paris",
                "code": "paris",
                "country": "CL",
                "supported_documents": ["credit_card"],
                "languages": ["es"]
            },
            {
                "name": "Cencosud",
                "code": "cencosud",
                "country": "CL",
                "supported_documents": ["credit_card"],
                "languages": ["es"]
            }
        ]
    }

@app.get("/formats/supported")
async def get_supported_formats():
    """Get list of supported document formats"""
    return {
        "supported_formats": [
            {"format": "PDF", "extensions": [".pdf"], "mime_types": ["application/pdf"]},
            {"format": "JPEG", "extensions": [".jpg", ".jpeg"], "mime_types": ["image/jpeg"]},
            {"format": "PNG", "extensions": [".png"], "mime_types": ["image/png"]},
            {"format": "TIFF", "extensions": [".tiff", ".tif"], "mime_types": ["image/tiff"]},
            {"format": "BMP", "extensions": [".bmp"], "mime_types": ["image/bmp"]},
        ]
    }

if __name__ == "__main__":
    uvicorn.run(
        "src.main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
        log_level="info"
    )