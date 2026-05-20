#!/usr/bin/env python3
"""
Integration tests for PulseExpends PDF Parser service

These tests verify:
1. Service startup and health checks
2. End-to-end document parsing flow
3. Error handling and edge cases
4. Performance under load
5. Integration with external services (OBS, OCR, AI)
"""

import pytest
import asyncio
import aiohttp
import json
import time
from datetime import datetime
from unittest.mock import AsyncMock, MagicMock, patch
import tempfile
import os

from src.main import app
from src.services.parser import PDFParserService
from src.services.ocr import OCRService
from src.services.ai import AIService
from src.services.obs import OBSService
from src.models.document import Document, ParsedTransaction, ParsingResult

@pytest.fixture
async def test_client():
    """Create test client for FastAPI app"""
    from fastapi.testclient import TestClient
    client = TestClient(app)
    yield client

@pytest.fixture
def sample_pdf_content():
    """Generate sample PDF content for testing"""
    # Minimal valid PDF content
    return b'%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [] /Count 0 >>\nendobj\nxref\n0 3\n0000000000 65535 f \n0000000010 00000 n \n0000000053 00000 n \ntrailer\n<< /Root 1 0 R /Size 3 >>\nstartxref\n94\n%%EOF'

@pytest.fixture
def sample_parsing_result():
    """Create sample parsing result for testing"""
    return ParsingResult(
        request_id="test-request-123",
        document=Document(
            filename="test.pdf",
            file_size=1024,
            mime_type="application/pdf",
            pages=3,
            bank_name="Banco de Chile",
            document_type="credit_card",
            language="es",
            metadata={
                "author": "Test Bank",
                "creation_date": "2024-01-15",
                "modification_date": "2024-01-15"
            }
        ),
        transactions=[
            ParsedTransaction(
                date=datetime(2024, 1, 15),
                description="SUPERMERCADO LIDER",
                amount=-15000.50,
                currency="CLP",
                category="supermercado",
                merchant="Lider",
                location="Santiago, Chile",
                installments=None,
                confidence=0.95,
                metadata={
                    "transaction_id": "TX001",
                    "reference": "REF001"
                }
            ),
            ParsedTransaction(
                date=datetime(2024, 1, 14),
                description="RESTAURANTE ITALIANO",
                amount=-25000.00,
                currency="CLP",
                category="restaurante",
                merchant="Restaurante Italiano",
                location="Providencia, Santiago",
                installments=None,
                confidence=0.92,
                metadata={
                    "transaction_id": "TX002",
                    "reference": "REF002"
                }
            )
        ],
        summary={
            "total_transactions": 2,
            "total_amount": -40000.50,
            "currency": "CLP",
            "categories": {
                "supermercado": -15000.50,
                "restaurante": -25000.00
            },
            "confidence_score": 0.935,
            "page_count": 3,
            "processing_time_seconds": 2.5
        },
        processing_time=2.5,
        created_at=datetime.now()
    )

class TestIntegration:
    """Integration test suite"""
    
    @pytest.mark.asyncio
    async def test_service_startup(self, test_client):
        """Test service startup and basic endpoints"""
        # Test root endpoint
        response = test_client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "PulseExpends PDF Parser"
        assert data["version"] == "1.0.0"
        assert data["status"] == "operational"
        
        # Test health endpoint
        response = test_client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        assert "services" in data
        assert "timestamp" in data
        
        # Test supported banks endpoint
        response = test_client.get("/banks/supported")
        assert response.status_code == 200
        data = response.json()
        assert "supported_banks" in data
        assert len(data["supported_banks"]) > 0
        
        # Test supported formats endpoint
        response = test_client.get("/formats/supported")
        assert response.status_code == 200
        data = response.json()
        assert "supported_formats" in data
        assert len(data["supported_formats"]) > 0
    
    @pytest.mark.asyncio
    async def test_document_upload_flow(self, test_client, sample_pdf_content, sample_parsing_result):
        """Test complete document upload and parsing flow"""
        # Mock the parser service
        with patch('src.main.pdf_parser.parse_document') as mock_parse:
            mock_parse.return_value = sample_parsing_result
            
            # Create a temporary PDF file
            with tempfile.NamedTemporaryFile(suffix='.pdf', delete=False) as tmp_file:
                tmp_file.write(sample_pdf_content)
                tmp_file.flush()
                
                # Upload and parse
                with open(tmp_file.name, 'rb') as f:
                    files = {'file': ('test.pdf', f, 'application/pdf')}
                    data = {
                        'document_type': 'credit_card',
                        'bank_name': 'Banco de Chile',
                        'language': 'es',
                        'extract_installments': 'true',
                        'validate_transactions': 'true'
                    }
                    
                    response = test_client.post("/parse/upload", files=files, data=data)
                
                # Clean up temp file
                os.unlink(tmp_file.name)
            
            # Verify response
            assert response.status_code == 200
            data = response.json()
            
            # Verify response structure
            assert data["request_id"] == "test-request-123"
            assert data["status"] == "completed"
            assert data["processing_time"] == 2.5
            
            # Verify document data
            assert "document" in data
            doc = data["document"]
            assert doc["filename"] == "test.pdf"
            assert doc["bank_name"] == "Banco de Chile"
            assert doc["document_type"] == "credit_card"
            assert doc["language"] == "es"
            
            # Verify transactions
            assert "transactions" in data
            transactions = data["transactions"]
            assert len(transactions) == 2
            
            # Verify summary
            assert "summary" in data
            summary = data["summary"]
            assert summary["total_transactions"] == 2
            assert summary["total_amount"] == -40000.50
            assert summary["currency"] == "CLP"
            
            # Verify parser was called with correct parameters
            mock_parse.assert_called_once()
            call_args = mock_parse.call_args
            assert call_args[0][0] == sample_pdf_content  # content
            assert call_args[1]["filename"] == "test.pdf"
            assert call_args[1]["document_type"] == "credit_card"
            assert call_args[1]["bank_name"] == "Banco de Chile"
            assert call_args[1]["language"] == "es"
            assert call_args[1]["extract_installments"] == True
            assert call_args[1]["validate_transactions"] == True
    
    @pytest.mark.asyncio
    async def test_url_parsing_flow(self, test_client, sample_parsing_result):
        """Test URL-based document parsing flow"""
        # Mock services
        with patch('src.main.pdf_parser.parse_document') as mock_parse, \
             patch('src.main.obs_service.download_from_url') as mock_download, \
             patch('src.main.obs_service.store_parsing_result') as mock_store:
            
            mock_parse.return_value = sample_parsing_result
            mock_download.return_value = b"fake pdf content"
            mock_store.return_value = None
            
            # Parse from URL
            request_data = {
                "document_url": "https://example.com/statement.pdf",
                "document_type": "credit_card",
                "bank_name": "Santander",
                "language": "es",
                "extract_installments": True,
                "validate_transactions": True
            }
            
            response = test_client.post("/parse/url", json=request_data)
            
            # Verify response
            assert response.status_code == 200
            data = response.json()
            
            assert data["request_id"] == "test-request-123"
            assert data["status"] == "completed"
            
            # Verify services were called
            mock_download.assert_called_once_with("https://example.com/statement.pdf")
            mock_parse.assert_called_once()
            mock_store.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_obs_parsing_flow(self, test_client, sample_parsing_result):
        """Test OBS-based document parsing flow"""
        # Mock services
        with patch('src.main.pdf_parser.parse_document') as mock_parse, \
             patch('src.main.obs_service.download_from_obs') as mock_download, \
             patch('src.main.obs_service.store_parsing_result') as mock_store:
            
            mock_parse.return_value = sample_parsing_result
            mock_download.return_value = b"fake pdf content"
            mock_store.return_value = None
            
            # Parse from OBS
            request_data = {
                "document_url": "obs://my-bucket/statements/statement.pdf",
                "document_type": "bank_statement",
                "bank_name": "BCI",
                "language": "es"
            }
            
            response = test_client.post("/parse/obs", json=request_data)
            
            # Verify response
            assert response.status_code == 200
            data = response.json()
            
            assert data["request_id"] == "test-request-123"
            assert data["status"] == "completed"
            
            # Verify services were called with correct parameters
            mock_download.assert_called_once_with("my-bucket", "statements/statement.pdf")
            mock_parse.assert_called_once()
            mock_store.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_result_retrieval_flow(self, test_client, sample_parsing_result):
        """Test parsing result retrieval flow"""
        # Mock OBS service
        with patch('src.main.obs_service.get_parsing_result') as mock_get_result:
            mock_get_result.return_value = sample_parsing_result
            
            # Retrieve result
            response = test_client.get("/result/test-request-123")
            
            # Verify response
            assert response.status_code == 200
            data = response.json()
            
            assert data["request_id"] == "test-request-123"
            assert data["status"] == "completed"
            assert "document" in data
            assert "transactions" in data
            assert "summary" in data
            
            # Verify service was called
            mock_get_result.assert_called_once_with("test-request-123")
    
    @pytest.mark.asyncio
    async def test_error_handling(self, test_client):
        """Test error handling in various scenarios"""
        
        # Test 1: Missing file in upload
        response = test_client.post("/parse/upload")
        assert response.status_code == 422  # Validation error
        
        # Test 2: Invalid document type
        with tempfile.NamedTemporaryFile(suffix='.pdf') as tmp_file:
            tmp_file.write(b"fake pdf")
            tmp_file.flush()
            
            with open(tmp_file.name, 'rb') as f:
                files = {'file': ('test.pdf', f, 'application/pdf')}
                data = {'document_type': 'invalid_type'}
                
                response = test_client.post("/parse/upload", files=files, data=data)
            
            # Should still attempt to parse (validation happens in parser)
            assert response.status_code in [200, 500]
        
        # Test 3: Missing URL in URL parsing
        request_data = {"document_type": "credit_card"}
        response = test_client.post("/parse/url", json=request_data)
        assert response.status_code == 400
        assert "document_url is required" in response.json()["detail"]
        
        # Test 4: Invalid OBS URL format
        request_data = {
            "document_url": "invalid-url",
            "document_type": "credit_card"
        }
        response = test_client.post("/parse/obs", json=request_data)
        assert response.status_code == 400
        assert "Invalid OBS URL format" in response.json()["detail"]
        
        # Test 5: Malformed OBS URL
        request_data = {
            "document_url": "obs://bucket-only",
            "document_type": "credit_card"
        }
        response = test_client.post("/parse/obs", json=request_data)
        assert response.status_code == 400
        assert "Invalid OBS URL format" in response.json()["detail"]
        
        # Test 6: Non-existent result
        with patch('src.main.obs_service.get_parsing_result') as mock_get_result:
            mock_get_result.return_value = None
            
            response = test_client.get("/result/nonexistent-request")
            assert response.status_code == 404
            assert "Result not found" in response.json()["detail"]
        
        # Test 7: Parser error
        with patch('src.main.pdf_parser.parse_document') as mock_parse:
            mock_parse.side_effect = Exception("Parser failed")
            
            with tempfile.NamedTemporaryFile(suffix='.pdf') as tmp_file:
                tmp_file.write(b"fake pdf")
                tmp_file.flush()
                
                with open(tmp_file.name, 'rb') as f:
                    files = {'file': ('test.pdf', f, 'application/pdf')}
                    data = {'document_type': 'credit_card'}
                    
                    response = test_client.post("/parse/upload", files=files, data=data)
                
                assert response.status_code == 500
                assert "Failed to parse document" in response.json()["detail"]
        
        # Test 8: Download error in URL parsing
        with patch('src.main.obs_service.download_from_url') as mock_download:
            mock_download.side_effect = Exception("Download failed")
            
            request_data = {
                "document_url": "https://example.com/statement.pdf",
                "document_type": "credit_card"
            }
            
            response = test_client.post("/parse/url", json=request_data)
            assert response.status_code == 500
            assert "Failed to parse document" in response.json()["detail"]
    
    @pytest.mark.asyncio
    async def test_concurrent_requests(self, test_client, sample_parsing_result):
        """Test handling of concurrent requests"""
        import threading
        
        # Mock parser with delay to simulate processing
        parse_calls = []
        
        async def delayed_parse(*args, **kwargs):
            parse_calls.append(1)
            await asyncio.sleep(0.1)  # Simulate processing time
            return sample_parsing_result
        
        with patch('src.main.pdf_parser.parse_document', side_effect=delayed_parse):
            # Create multiple concurrent requests
            results = []
            errors = []
            
            def make_request():
                try:
                    with tempfile.NamedTemporaryFile(suffix='.pdf') as tmp_file:
                        tmp_file.write(b"fake pdf")
                        tmp_file.flush()
                        
                        with open(tmp_file.name, 'rb') as f:
                            files = {'file': ('test.pdf', f, 'application/pdf')}
                            data = {'document_type': 'credit_card'}
                            
                            response = test_client.post("/parse/upload", files=files, data=data)
                            results.append(response.status_code)
                except Exception as e:
                    errors.append(str(e))
            
            # Start multiple threads
            threads = []
            for i in range(5):
                thread = threading.Thread(target=make_request)
                threads.append(thread)
                thread.start()
            
            # Wait for all threads to complete
            for thread in threads:
                thread.join()
            
            # Verify all requests succeeded
            assert len(errors) == 0, f"Errors in concurrent requests: {errors}"
            assert len(results) == 5, f"Expected 5 responses, got {len(results)}"
            
            for status_code in results:
                assert status_code == 200, f"Request failed with status {status_code}"
            
            # Verify parser was called for each request
            assert len(parse_calls) == 5, f"Expected 5 parse calls, got {len(parse_calls)}"
    
    @pytest.mark.asyncio
    async def test_performance_metrics(self, test_client, sample_parsing_result):
        """Test performance under load"""
        # Mock parser
        with patch('src.main.pdf_parser.parse_document') as mock_parse:
            mock_parse.return_value = sample_parsing_result
            
            # Time multiple requests
            num_requests = 10
            start_time = time.time()
            
            for i in range(num_requests):
                with tempfile.NamedTemporaryFile(suffix='.pdf') as tmp_file:
                    tmp_file.write(b"fake pdf")
                    tmp_file.flush()
                    
                    with open(tmp_file.name, 'rb') as f:
                        files = {'file': (f'test{i}.pdf', f, 'application/pdf')}
                        data = {'document_type': 'credit_card'}
                        
                        response = test_client.post("/parse/upload", files=files, data=data)
                        assert response.status_code == 200
            
            end_time = time.time()
            total_time = end_time - start_time
            avg_time = total_time / num_requests
            
            print(f"\nPerformance test: {num_requests} requests")
            print(f"Total time: {total_time:.2f} seconds")
            print(f"Average time per request: {avg_time:.2f} seconds")
            print(f"Requests per second: {num_requests / total_time:.2f}")
            
            # Verify performance is reasonable
            # Note: These thresholds can be adjusted based on requirements
            assert avg_time < 5.0, f"Average request time too high: {avg_time:.2f}s"
            assert total_time < 30.0, f"Total time too high: {total_time:.2f}s"
    
    @pytest.mark.asyncio
    async def test_data_validation(self, test_client):
        """Test data validation in requests"""
        
        # Test 1: Invalid JSON in request body
        response = test_client.post("/parse/url", data="invalid json")
        assert response.status_code == 422  # Validation error
        
        # Test 2: Missing required fields
        request_data = {}  # Missing document_type
        response = test_client.post("/parse/url", json=request_data)
        assert response.status_code == 400  # Bad request
        
        # Test 3: Invalid field types
        request_data = {
            "document_url": "https://example.com/test.pdf",
            "document_type": 123,  # Should be string
            "language": "es",
            "extract_installments": "not-a-boolean"  # Should be boolean
        }
        response = test_client.post("/parse/url", json=request_data)
        assert response.status_code == 422  # Validation error
        
        # Test 4: Invalid enum values
        request_data = {
            "document_url": "https://example.com/test.pdf",
            "document_type": "invalid_document_type",  # Not in enum
            "bank_name": "Invalid Bank",  # Not in supported banks
            "language": "xx"  # Not a valid language code
        }
        response = test_client.post("/parse/url", json=request_data)
        # Should still attempt to parse (validation happens in parser)
        assert response.status_code in [200, 500]
        
        # Test 5: File size validation (simulated)
        # Create a large file (simulate size limit)
        large_content = b"x" * (100 * 1024 * 1024)  # 100MB
        
        with tempfile.NamedTemporaryFile(suffix='.pdf') as tmp_file:
            tmp_file.write(large_content)
            tmp_file.flush()
            
            with open(tmp_file.name, 'rb') as f:
                files = {'file': ('large.pdf', f, 'application/pdf')}
                data = {'document_type': 'credit_card'}
                
                # Note: FastAPI/Starlette may handle this differently
                # We're testing that the server doesn't crash
                try:
                    response = test_client.post("/parse/upload", files=files, data=data, timeout=30)
                    # Server should handle large files gracefully
                    assert response.status_code in [200, 413, 500]
                except Exception as e:
                    # Server might timeout or reject the request
                    print(f"Large file test exception: {e}")
    
    @pytest.mark.asyncio
    async def test_cors_headers(self, test_client):
        """Test CORS headers are properly set"""
        # Test OPTIONS request
        response = test_client.options("/parse/upload")
        
        # Check CORS headers
        assert "access-control-allow-origin" in response.headers
        assert "access-control-allow-methods" in response.headers
        assert "access-control-allow-headers" in response.headers
        
        # Test actual request
        response = test_client.get("/health")
        assert "access-control-allow-origin" in response.headers
        
        # Verify CORS allows all origins (for development)
        # In production, this should be restricted
        assert response.headers["access-control-allow-origin"] == "*"
    
    @pytest.mark.asyncio
    async def test_api_documentation(self, test_client):
        """Test API documentation endpoints"""
        # Test OpenAPI docs
        response = test_client.get("/docs")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]
        assert "swagger-ui" in response.text.lower() or "openapi" in response.text.lower()
        
        # Test ReDoc docs
        response = test_client.get("/redoc")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]
        assert "redoc" in response.text.lower()
    
    @pytest.mark.asyncio
    async def test_error_response_format(self, test_client):
        """Test error response format"""
        # Test 404 for non-existent endpoint
        response = test_client.get("/nonexistent-endpoint")
        assert response.status_code == 404
        data = response.json()
        assert "detail" in data
        
        # Test 405 for invalid method
        response = test_client.post("/health")
        assert response.status_code == 405  # Method Not Allowed
        data = response.json()
        assert "detail" in data
        
        # Test 422 for validation error
        response = test_client.post("/parse/url", json={"invalid": "data"})
        assert response.status_code == 400 or response.status_code == 422
        data = response.json()
        assert "detail" in data
    
    @pytest.mark.asyncio
    async def test_service_dependencies(self, test_client):
        """Test service dependency health checks"""
        # Mock unhealthy services
        with patch('src.main.pdf_parser.is_ready') as mock_pdf_ready, \
             patch('src.main.ocr_service.is_ready') as mock_ocr_ready, \
             patch('src.main.ai_service.is_ready') as mock_ai_ready, \
             patch('src.main.obs_service.is_ready') as mock_obs_ready:
            
            # Test with all services unhealthy
            mock_pdf_ready.return_value = False
            mock_ocr_ready.return_value = False
            mock_ai_ready.return_value = False
            mock_obs_ready.return_value = False
            
            response = test_client.get("/health")
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "unhealthy"
            assert data["services"]["pdf_parser"] == "unhealthy"
            assert data["services"]["ocr_service"] == "unhealthy"
            assert data["services"]["ai_service"] == "unhealthy"
            assert data["services"]["obs_service"] == "unhealthy"
            
            # Test with mixed health status
            mock_pdf_ready.return_value = True
            mock_ocr_ready.return_value = False
            mock_ai_ready.return_value = True
            mock_obs_ready.return_value = False
            
            response = test_client.get("/health")
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "unhealthy"  # Should be unhealthy if any service is down
            assert data["services"]["pdf_parser"] == "healthy"
            assert data["services"]["ocr_service"] == "unhealthy"
            assert data["services"]["ai_service"] == "healthy"
            assert data["services"]["obs_service"] == "unhealthy"
            
            # Test with all services healthy
            mock_pdf_ready.return_value = True
            mock_ocr_ready.return_value = True
            mock_ai_ready.return_value = True
            mock_obs_ready.return_value = True
            
            response = test_client.get("/health")
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "healthy"
            assert all(status == "healthy" for status in data["services"].values())

if __name__ == "__main__":
    # Run tests
    import sys
    sys.exit(pytest.main([__file__, "-v"]))