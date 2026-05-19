"""
Pytest configuration for PDF Parser tests
"""
import pytest
import asyncio
import sys
import os

# Add src directory to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

@pytest.fixture(scope="session")
def event_loop():
    """Create event loop for async tests"""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()

@pytest.fixture
def sample_transaction_data():
    """Sample transaction data for testing"""
    return {
        "date": "2024-01-15T10:30:00Z",
        "description": "SUPERMERCADO LIDER",
        "amount": -15000.50,
        "currency": "CLP",
        "category": "supermercado",
        "merchant": "Lider",
        "location": "Santiago, Chile",
        "installments": None,
        "confidence": 0.95,
        "metadata": {
            "transaction_id": "TX001",
            "reference": "REF001"
        }
    }

@pytest.fixture
def sample_document_data():
    """Sample document data for testing"""
    return {
        "filename": "test.pdf",
        "file_size": 1024,
        "mime_type": "application/pdf",
        "pages": 3,
        "bank_name": "Banco de Chile",
        "document_type": "credit_card",
        "language": "es",
        "metadata": {
            "author": "Test Bank",
            "creation_date": "2024-01-15",
            "modification_date": "2024-01-15"
        }
    }

@pytest.fixture
def mock_obs_service():
    """Mock OBS service for testing"""
    class MockOBSService:
        def __init__(self):
            self.download_calls = []
            self.store_calls = []
            self.get_calls = []
        
        async def download_from_url(self, url):
            self.download_calls.append(('url', url))
            return b"fake pdf content"
        
        async def download_from_obs(self, bucket, key):
            self.download_calls.append(('obs', bucket, key))
            return b"fake pdf content"
        
        async def store_parsing_result(self, request_id, result):
            self.store_calls.append((request_id, result))
            return True
        
        async def get_parsing_result(self, request_id):
            self.get_calls.append(request_id)
            # Return None to simulate not found
            return None
        
        def is_ready(self):
            return True
    
    return MockOBSService()

@pytest.fixture
def mock_ocr_service():
    """Mock OCR service for testing"""
    class MockOCRService:
        def __init__(self):
            self.extract_calls = []
        
        async def extract_text(self, image_data, language="es"):
            self.extract_calls.append((image_data, language))
            return "Extracted text from image"
        
        def is_ready(self):
            return True
    
    return MockOCRService()

@pytest.fixture
def mock_ai_service():
    """Mock AI service for testing"""
    class MockAIService:
        def __init__(self):
            self.parse_calls = []
            self.classify_calls = []
        
        async def parse_transactions(self, text, bank_name="generic", language="es"):
            self.parse_calls.append((text, bank_name, language))
            return [
                {
                    "date": "2024-01-15",
                    "description": "SUPERMERCADO LIDER",
                    "amount": -15000.50,
                    "currency": "CLP",
                    "category": "supermercado",
                    "confidence": 0.95
                }
            ]
        
        async def classify_transaction(self, description, amount):
            self.classify_calls.append((description, amount))
            return "supermercado", 0.95
        
        def is_ready(self):
            return True
    
    return MockAIService()

@pytest.fixture
def mock_pdf_parser(mock_ocr_service, mock_ai_service, mock_obs_service):
    """Mock PDF parser service for testing"""
    from src.services.parser import PDFParserService
    
    class MockPDFParserService(PDFParserService):
        def __init__(self):
            self.parse_calls = []
            self.mock_ocr = mock_ocr_service
            self.mock_ai = mock_ai_service
            self.mock_obs = mock_obs_service
        
        async def parse_document(self, content, filename, document_type="credit_card", 
                               bank_name="generic", language="es", extract_installments=True,
                               validate_transactions=True):
            self.parse_calls.append((filename, document_type, bank_name, language))
            
            # Return a mock result
            from src.models.document import ParsingResult, Document, ParsedTransaction
            from datetime import datetime
            
            return ParsingResult(
                request_id="test-request-123",
                document=Document(
                    filename=filename,
                    file_size=len(content),
                    mime_type="application/pdf",
                    pages=3,
                    bank_name=bank_name,
                    document_type=document_type,
                    language=language
                ),
                transactions=[
                    ParsedTransaction(
                        date=datetime.now(),
                        description="Test Transaction",
                        amount=-10000.00,
                        currency="CLP",
                        category="test",
                        confidence=0.90
                    )
                ],
                summary={
                    "total_transactions": 1,
                    "total_amount": -10000.00,
                    "currency": "CLP",
                    "categories": {"test": -10000.00},
                    "confidence_score": 0.90
                },
                processing_time=1.0,
                created_at=datetime.now()
            )
        
        def is_ready(self):
            return True
    
    return MockPDFParserService()