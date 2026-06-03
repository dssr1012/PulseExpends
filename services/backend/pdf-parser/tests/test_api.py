import pytest
import json
from datetime import datetime
from unittest.mock import AsyncMock, MagicMock, patch
from fastapi.testclient import TestClient
from fastapi import HTTPException

from src.main import app
from src.services.parser import PDFParserService
from src.services.ocr import OCRService
from src.services.ai import AIService
from src.services.obs import OBSService
from src.models.document import Document, ParsedTransaction, ParsingResult

client = TestClient(app)

@pytest.fixture
def mock_services():
    """Mock all services for testing"""
    with patch('src.main.pdf_parser') as mock_pdf_parser, \
         patch('src.main.ocr_service') as mock_ocr_service, \
         patch('src.main.ai_service') as mock_ai_service, \
         patch('src.main.obs_service') as mock_obs_service:
        
        # Configure mock services
        mock_pdf_parser.is_ready.return_value = True
        mock_ocr_service.is_ready.return_value = True
        mock_ai_service.is_ready.return_value = True
        mock_obs_service.is_ready.return_value = True
        
        yield {
            'pdf_parser': mock_pdf_parser,
            'ocr_service': mock_ocr_service,
            'ai_service': mock_ai_service,
            'obs_service': mock_obs_service
        }

def test_root_endpoint():
    """Test root endpoint"""
    response = client.get("/")
    assert response.status_code == 200
    data = response.json()
    assert data["service"] == "PulseExpends PDF Parser"
    assert data["version"] == "1.0.0"
    assert data["status"] == "operational"

def test_health_check(mock_services):
    """Test health check endpoint"""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert data["version"] == "1.0.0"
    assert "timestamp" in data
    assert "services" in data
    assert data["services"]["pdf_parser"] == "healthy"
    assert data["services"]["ocr_service"] == "healthy"
    assert data["services"]["ai_service"] == "healthy"
    assert data["services"]["obs_service"] == "healthy"

def test_health_check_unhealthy(mock_services):
    """Test health check when a service is unhealthy"""
    mock_services['pdf_parser'].is_ready.return_value = False
    
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "unhealthy"
    assert data["services"]["pdf_parser"] == "unhealthy"

def test_get_supported_banks():
    """Test supported banks endpoint"""
    response = client.get("/banks/supported")
    assert response.status_code == 200
    data = response.json()
    assert "supported_banks" in data
    banks = data["supported_banks"]
    
    # Check that we have expected banks
    bank_names = [bank["name"] for bank in banks]
    expected_banks = [
        "Banco de Chile", "Banco Estado", "Santander", "BCI",
        "Scotiabank", "Itaú", "Falabella", "Ripley", "Paris", "Cencosud"
    ]
    
    for expected_bank in expected_banks:
        assert expected_bank in bank_names
    
    # Check bank structure
    for bank in banks:
        assert "name" in bank
        assert "code" in bank
        assert "country" in bank
        assert "supported_documents" in bank
        assert "languages" in bank

def test_get_supported_formats():
    """Test supported formats endpoint"""
    response = client.get("/formats/supported")
    assert response.status_code == 200
    data = response.json()
    assert "supported_formats" in data
    formats = data["supported_formats"]
    
    # Check expected formats
    format_names = [fmt["format"] for fmt in formats]
    expected_formats = ["PDF", "JPEG", "PNG", "TIFF", "BMP"]
    
    for expected_format in expected_formats:
        assert expected_format in format_names
    
    # Check format structure
    for fmt in formats:
        assert "format" in fmt
        assert "extensions" in fmt
        assert "mime_types" in fmt
        assert isinstance(fmt["extensions"], list)
        assert isinstance(fmt["mime_types"], list)

def test_parse_upload_success(mock_services):
    """Test successful document upload parsing"""
    # Mock parsing result
    mock_result = ParsingResult(
        request_id="test-request-123",
        document=Document(
            filename="test.pdf",
            file_size=1024,
            mime_type="application/pdf",
            pages=5,
            bank_name="Banco de Chile",
            document_type="credit_card",
            language="es"
        ),
        transactions=[
            ParsedTransaction(
                date=datetime.now(),
                description="Supermercado",
                amount=-15000.50,
                currency="CLP",
                category="supermercado",
                merchant="Supermercado Lider",
                location="Santiago, Chile",
                installments=None,
                confidence=0.95
            )
        ],
        summary={
            "total_transactions": 1,
            "total_amount": -15000.50,
            "currency": "CLP",
            "categories": {"supermercado": -15000.50},
            "confidence_score": 0.95
        },
        processing_time=2.5,
        created_at=datetime.now()
    )
    
    mock_services['pdf_parser'].parse_document = AsyncMock(return_value=mock_result)
    
    # Create test file
    test_file = ("test.pdf", b"fake pdf content", "application/pdf")
    
    # Make request
    response = client.post(
        "/parse/upload",
        files={"file": test_file},
        data={
            "document_type": "credit_card",
            "bank_name": "Banco de Chile",
            "language": "es",
            "extract_installments": "true",
            "validate_transactions": "true"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    
    assert data["request_id"] == "test-request-123"
    assert data["status"] == "completed"
    assert data["processing_time"] == 2.5
    assert "document" in data
    assert "transactions" in data
    assert "summary" in data
    
    # Verify document data
    doc = data["document"]
    assert doc["filename"] == "test.pdf"
    assert doc["bank_name"] == "Banco de Chile"
    assert doc["document_type"] == "credit_card"
    
    # Verify transactions
    transactions = data["transactions"]
    assert len(transactions) == 1
    assert transactions[0]["description"] == "Supermercado"
    assert transactions[0]["amount"] == -15000.50
    assert transactions[0]["currency"] == "CLP"
    assert transactions[0]["category"] == "supermercado"

def test_parse_upload_missing_file():
    """Test document upload without file"""
    response = client.post("/parse/upload")
    assert response.status_code == 422  # Validation error

def test_parse_upload_parser_error(mock_services):
    """Test document upload with parser error"""
    mock_services['pdf_parser'].parse_document = AsyncMock(
        side_effect=Exception("Failed to parse PDF")
    )
    
    test_file = ("test.pdf", b"fake pdf content", "application/pdf")
    
    response = client.post(
        "/parse/upload",
        files={"file": test_file},
        data={"document_type": "credit_card"}
    )
    
    assert response.status_code == 500
    data = response.json()
    assert "detail" in data
    assert "Failed to parse document" in data["detail"]

def test_parse_url_success(mock_services):
    """Test successful URL parsing"""
    # Mock parsing result
    mock_result = ParsingResult(
        request_id="test-request-456",
        document=Document(
            filename="statement.pdf",
            file_size=2048,
            mime_type="application/pdf",
            pages=3,
            bank_name="Santander",
            document_type="credit_card",
            language="es"
        ),
        transactions=[
            ParsedTransaction(
                date=datetime.now(),
                description="Restaurante",
                amount=-25000.00,
                currency="CLP",
                category="restaurante",
                merchant="Restaurante Italiano",
                location="Providencia, Santiago",
                installments=None,
                confidence=0.92
            )
        ],
        summary={
            "total_transactions": 1,
            "total_amount": -25000.00,
            "currency": "CLP",
            "categories": {"restaurante": -25000.00},
            "confidence_score": 0.92
        },
        processing_time=1.8,
        created_at=datetime.now()
    )
    
    mock_services['pdf_parser'].parse_document = AsyncMock(return_value=mock_result)
    mock_services['obs_service'].download_from_url = AsyncMock(return_value=b"fake pdf content")
    mock_services['obs_service'].store_parsing_result = AsyncMock()
    
    # Make request
    request_data = {
        "document_url": "https://example.com/statement.pdf",
        "document_type": "credit_card",
        "bank_name": "Santander",
        "language": "es",
        "extract_installments": True,
        "validate_transactions": True
    }
    
    response = client.post("/parse/url", json=request_data)
    
    assert response.status_code == 200
    data = response.json()
    
    assert data["request_id"] == "test-request-456"
    assert data["status"] == "completed"
    assert "document" in data
    assert "transactions" in data
    
    # Verify OBS service was called
    mock_services['obs_service'].download_from_url.assert_called_once_with(
        "https://example.com/statement.pdf"
    )
    mock_services['obs_service'].store_parsing_result.assert_called_once()

def test_parse_url_missing_url():
    """Test URL parsing without URL"""
    request_data = {
        "document_type": "credit_card"
    }
    
    response = client.post("/parse/url", json=request_data)
    assert response.status_code == 400
    data = response.json()
    assert "document_url is required" in data["detail"]

def test_parse_obs_success(mock_services):
    """Test successful OBS parsing"""
    # Mock parsing result
    mock_result = ParsingResult(
        request_id="test-request-789",
        document=Document(
            filename="bank_statement.pdf",
            file_size=3072,
            mime_type="application/pdf",
            pages=4,
            bank_name="BCI",
            document_type="bank_statement",
            language="es"
        ),
        transactions=[
            ParsedTransaction(
                date=datetime.now(),
                description="Transferencia",
                amount=100000.00,
                currency="CLP",
                category="ingreso",
                merchant=None,
                location=None,
                installments=None,
                confidence=0.98
            )
        ],
        summary={
            "total_transactions": 1,
            "total_amount": 100000.00,
            "currency": "CLP",
            "categories": {"ingreso": 100000.00},
            "confidence_score": 0.98
        },
        processing_time=3.2,
        created_at=datetime.now()
    )
    
    mock_services['pdf_parser'].parse_document = AsyncMock(return_value=mock_result)
    mock_services['obs_service'].download_from_obs = AsyncMock(return_value=b"fake pdf content")
    mock_services['obs_service'].store_parsing_result = AsyncMock()
    
    # Make request
    request_data = {
        "document_url": "obs://my-bucket/statements/bank_statement.pdf",
        "document_type": "bank_statement",
        "bank_name": "BCI",
        "language": "es"
    }
    
    response = client.post("/parse/obs", json=request_data)
    
    assert response.status_code == 200
    data = response.json()
    
    assert data["request_id"] == "test-request-789"
    assert data["status"] == "completed"
    
    # Verify OBS service was called with correct parameters
    mock_services['obs_service'].download_from_obs.assert_called_once_with(
        "my-bucket", "statements/bank_statement.pdf"
    )
    mock_services['obs_service'].store_parsing_result.assert_called_once()

def test_parse_obs_invalid_url():
    """Test OBS parsing with invalid URL"""
    request_data = {
        "document_url": "invalid-url",
        "document_type": "credit_card"
    }
    
    response = client.post("/parse/obs", json=request_data)
    assert response.status_code == 400
    data = response.json()
    assert "Invalid OBS URL format" in data["detail"]

def test_parse_obs_malformed_url():
    """Test OBS parsing with malformed URL"""
    request_data = {
        "document_url": "obs://bucket-only",
        "document_type": "credit_card"
    }
    
    response = client.post("/parse/obs", json=request_data)
    assert response.status_code == 400
    data = response.json()
    assert "Invalid OBS URL format" in data["detail"]

def test_get_result_success(mock_services):
    """Test successful result retrieval"""
    # Mock parsing result
    mock_result = ParsingResult(
        request_id="test-request-999",
        document=Document(
            filename="test.pdf",
            file_size=1024,
            mime_type="application/pdf",
            pages=5,
            bank_name="Banco de Chile",
            document_type="credit_card",
            language="es"
        ),
        transactions=[
            ParsedTransaction(
                date=datetime.now(),
                description="Test transaction",
                amount=-10000.00,
                currency="CLP",
                category="test",
                merchant=None,
                location=None,
                installments=None,
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
        processing_time=2.0,
        created_at=datetime.now()
    )
    
    mock_services['obs_service'].get_parsing_result = AsyncMock(return_value=mock_result)
    
    # Make request
    response = client.get("/result/test-request-999")
    
    assert response.status_code == 200
    data = response.json()
    
    assert data["request_id"] == "test-request-999"
    assert data["status"] == "completed"
    assert data["processing_time"] == 2.0
    
    # Verify OBS service was called
    mock_services['obs_service'].get_parsing_result.assert_called_once_with("test-request-999")

def test_get_result_not_found(mock_services):
    """Test result retrieval for non-existent request"""
    mock_services['obs_service'].get_parsing_result = AsyncMock(return_value=None)
    
    response = client.get("/result/nonexistent-request")
    
    assert response.status_code == 404
    data = response.json()
    assert "Result not found" in data["detail"]

def test_get_result_error(mock_services):
    """Test result retrieval with error"""
    mock_services['obs_service'].get_parsing_result = AsyncMock(
        side_effect=Exception("Database error")
    )
    
    response = client.get("/result/error-request")
    
    assert response.status_code == 500
    data = response.json()
    assert "Failed to retrieve result" in data["detail"]

def test_parse_upload_invalid_document_type():
    """Test document upload with invalid document type"""
    test_file = ("test.pdf", b"fake pdf content", "application/pdf")
    
    response = client.post(
        "/parse/upload",
        files={"file": test_file},
        data={"document_type": "invalid_type"}
    )
    
    # Should still succeed as validation happens in parser
    assert response.status_code == 200 or response.status_code == 500

def test_parse_upload_large_file(mock_services):
    """Test document upload with large file handling"""
    # Mock parsing result
    mock_result = ParsingResult(
        request_id="test-request-large",
        document=Document(
            filename="large.pdf",
            file_size=50 * 1024 * 1024,  # 50MB
            mime_type="application/pdf",
            pages=100,
            bank_name="Banco Estado",
            document_type="bank_statement",
            language="es"
        ),
        transactions=[],
        summary={
            "total_transactions": 0,
            "total_amount": 0,
            "currency": "CLP",
            "categories": {},
            "confidence_score": 0.0
        },
        processing_time=10.5,
        created_at=datetime.now()
    )
    
    mock_services['pdf_parser'].parse_document = AsyncMock(return_value=mock_result)
    
    # Create large file (in memory for test)
    large_content = b"x" * (10 * 1024 * 1024)  # 10MB for test
    test_file = ("large.pdf", large_content, "application/pdf")
    
    response = client.post(
        "/parse/upload",
        files={"file": test_file},
        data={"document_type": "bank_statement"}
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["request_id"] == "test-request-large"
    assert data["document"]["file_size"] == 50 * 1024 * 1024

def test_parse_url_download_error(mock_services):
    """Test URL parsing with download error"""
    mock_services['obs_service'].download_from_url = AsyncMock(
        side_effect=Exception("Failed to download from URL")
    )
    
    request_data = {
        "document_url": "https://example.com/statement.pdf",
        "document_type": "credit_card"
    }
    
    response = client.post("/parse/url", json=request_data)
    
    assert response.status_code == 500
    data = response.json()
    assert "Failed to parse document" in data["detail"]

def test_parse_obs_download_error(mock_services):
    """Test OBS parsing with download error"""
    mock_services['obs_service'].download_from_obs = AsyncMock(
        side_effect=Exception("OBS connection failed")
    )
    
    request_data = {
        "document_url": "obs://my-bucket/statements/bank_statement.pdf",
        "document_type": "bank_statement"
    }
    
    response = client.post("/parse/obs", json=request_data)
    
    assert response.status_code == 500
    data = response.json()
    assert "Failed to parse document" in data["detail"]

def test_parse_upload_unsupported_format(mock_services):
    """Test document upload with unsupported format"""
    mock_services['pdf_parser'].parse_document = AsyncMock(
        side_effect=ValueError("Unsupported file format")
    )
    
    test_file = ("test.txt", b"plain text content", "text/plain")
    
    response = client.post(
        "/parse/upload",
        files={"file": test_file},
        data={"document_type": "credit_card"}
    )
    
    assert response.status_code == 500
    data = response.json()
    assert "Failed to parse document" in data["detail"]

def test_concurrent_requests(mock_services):
    """Test handling of concurrent requests"""
    import asyncio
    
    # Mock parsing with delay to simulate concurrent processing
    async def delayed_parse(*args, **kwargs):
        await asyncio.sleep(0.1)
        return ParsingResult(
            request_id="concurrent-test",
            document=Document(
                filename="test.pdf",
                file_size=1024,
                mime_type="application/pdf",
                pages=1,
                bank_name="generic",
                document_type="credit_card",
                language="es"
            ),
            transactions=[],
            summary={},
            processing_time=0.1,
            created_at=datetime.now()
        )
    
    mock_services['pdf_parser'].parse_document = delayed_parse
    
    # Create multiple concurrent requests
    import threading
    results = []
    
    def make_request():
        test_file = ("test.pdf", b"fake pdf content", "application/pdf")
        response = client.post(
            "/parse/upload",
            files={"file": test_file},
            data={"document_type": "credit_card"}
        )
        results.append(response.status_code)
    
    # Start multiple threads
    threads = []
    for _ in range(5):
        thread = threading.Thread(target=make_request)
        threads.append(thread)
        thread.start()
    
    # Wait for all threads to complete
    for thread in threads:
        thread.join()
    
    # All requests should succeed
    for status_code in results:
        assert status_code == 200

def test_cors_headers():
    """Test CORS headers are properly set"""
    response = client.options("/parse/upload")
    
    # Check CORS headers
    assert "access-control-allow-origin" in response.headers
    assert "access-control-allow-methods" in response.headers
    assert "access-control-allow-headers" in response.headers
    
    # Test actual request
    response = client.get("/health")
    assert "access-control-allow-origin" in response.headers

def test_api_documentation():
    """Test api documentation endpoints"""
    # Test OpenAPI docs
    response = client.get("/docs")
    assert response.status_code == 200
    assert "text/html" in response.headers["content-type"]
    
    # Test ReDoc docs
    response = client.get("/redoc")
    assert response.status_code == 200
    assert "text/html" in response.headers["content-type"]

def test_error_response_format():
    """Test error response format"""
    # Make request to non-existent endpoint
    response = client.get("/nonexistent")
    assert response.status_code == 404
    
    data = response.json()
    assert "detail" in data
    
    # Make invalid request
    response = client.post("/parse/url", data="invalid json")
    assert response.status_code == 422  # Validation error
    
    data = response.json()
    assert "detail" in data

if __name__ == "__main__":
    pytest.main([__file__, "-v"])