import logging
import asyncio
import aiohttp
import tempfile
import os
from typing import Optional, Dict, Any, BinaryIO
from datetime import datetime, timedelta
import json
import hashlib

from ..config import settings

logger = logging.getLogger(__name__)

class OBSService:
    """Service for interacting with Huawei Cloud OBS (Object Storage Service)"""
    
    def __init__(self):
        self.endpoint = settings.OBS_ENDPOINT
        self.access_key_id = settings.OBS_ACCESS_KEY_ID
        self.secret_access_key = settings.OBS_SECRET_ACCESS_KEY
        self.bucket_name = settings.OBS_BUCKET_NAME
        self.region = settings.OBS_REGION
        
        # Initialize OBS client (using boto3-compatible interface)
        self.client = None
        self._initialize_client()
    
    def _initialize_client(self):
        """Initialize OBS client"""
        try:
            import boto3
            from botocore.client import Config
            
            # Huawei Cloud OBS uses S3-compatible API
            self.client = boto3.client(
                's3',
                endpoint_url=self.endpoint,
                aws_access_key_id=self.access_key_id,
                aws_secret_access_key=self.secret_access_key,
                region_name=self.region,
                config=Config(signature_version='s3v4')
            )
            
            # Test connection
            self._test_connection()
            logger.info(f"OBS client initialized for bucket: {self.bucket_name}")
            
        except ImportError:
            logger.error("boto3 not installed. Install with: pip install boto3")
            self.client = None
        except Exception as e:
            logger.error(f"Failed to initialize OBS client: {str(e)}")
            self.client = None
    
    def _test_connection(self):
        """Test OBS connection"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            # Check if bucket exists
            self.client.head_bucket(Bucket=self.bucket_name)
            logger.info(f"Bucket {self.bucket_name} exists and is accessible")
        except self.client.exceptions.NoSuchBucket:
            # Create bucket if it doesn't exist
            self._create_bucket()
        except Exception as e:
            logger.error(f"Failed to connect to OBS: {str(e)}")
            raise
    
    def _create_bucket(self):
        """Create OBS bucket if it doesn't exist"""
        try:
            self.client.create_bucket(
                Bucket=self.bucket_name,
                CreateBucketConfiguration={
                    'LocationConstraint': self.region
                }
            )
            logger.info(f"Created bucket: {self.bucket_name}")
            
            # Set bucket policies for security
            self._set_bucket_policies()
            
        except Exception as e:
            logger.error(f"Failed to create bucket {self.bucket_name}: {str(e)}")
            raise
    
    def _set_bucket_policies(self):
        """Set bucket security policies"""
        try:
            # Enable server-side encryption
            self.client.put_bucket_encryption(
                Bucket=self.bucket_name,
                ServerSideEncryptionConfiguration={
                    'Rules': [
                        {
                            'ApplyServerSideEncryptionByDefault': {
                                'SSEAlgorithm': 'AES256'
                            }
                        }
                    ]
                }
            )
            
            # Set lifecycle policies
            self.client.put_bucket_lifecycle_configuration(
                Bucket=self.bucket_name,
                LifecycleConfiguration={
                    'Rules': [
                        {
                            'ID': 'TempFilesExpiration',
                            'Filter': {
                                'Prefix': 'temp/'
                            },
                            'Status': 'Enabled',
                            'Expiration': {
                                'Days': 7
                            }
                        },
                        {
                            'ID': 'OldParsingResults',
                            'Filter': {
                                'Prefix': 'results/'
                            },
                            'Status': 'Enabled',
                            'Expiration': {
                                'Days': 30
                            }
                        }
                    ]
                }
            )
            
            logger.info(f"Set security policies for bucket: {self.bucket_name}")
            
        except Exception as e:
            logger.warning(f"Failed to set bucket policies: {str(e)}")
    
    async def upload_document(
        self, 
        file_content: bytes, 
        filename: str, 
        metadata: Optional[Dict[str, str]] = None,
        content_type: Optional[str] = None
    ) -> str:
        """Upload a document to OBS"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            # Generate unique key
            file_hash = hashlib.md5(file_content).hexdigest()
            timestamp = datetime.utcnow().strftime("%Y/%m/%d")
            key = f"documents/{timestamp}/{file_hash}_{filename}"
            
            # Determine content type
            if not content_type:
                content_type = self._get_content_type(filename)
            
            # Prepare metadata
            extra_args = {
                'ContentType': content_type,
                'Metadata': metadata or {}
            }
            
            # Upload file
            self.client.put_object(
                Bucket=self.bucket_name,
                Key=key,
                Body=file_content,
                **extra_args
            )
            
            # Generate presigned URL for temporary access
            url = self.client.generate_presigned_url(
                'get_object',
                Params={'Bucket': self.bucket_name, 'Key': key},
                ExpiresIn=3600  # 1 hour
            )
            
            logger.info(f"Uploaded document to OBS: {key}")
            return key
            
        except Exception as e:
            logger.error(f"Failed to upload document to OBS: {str(e)}")
            raise
    
    async def download_document(self, key: str) -> bytes:
        """Download a document from OBS"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            response = self.client.get_object(
                Bucket=self.bucket_name,
                Key=key
            )
            
            file_content = response['Body'].read()
            logger.info(f"Downloaded document from OBS: {key} ({len(file_content)} bytes)")
            return file_content
            
        except self.client.exceptions.NoSuchKey:
            logger.error(f"Document not found in OBS: {key}")
            raise FileNotFoundError(f"Document not found: {key}")
        except Exception as e:
            logger.error(f"Failed to download document from OBS: {str(e)}")
            raise
    
    async def download_from_url(self, url: str) -> bytes:
        """Download document from external URL"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(url) as response:
                    if response.status != 200:
                        raise ValueError(f"Failed to download from URL: {response.status}")
                    
                    content = await response.read()
                    logger.info(f"Downloaded document from URL: {url} ({len(content)} bytes)")
                    return content
                    
        except Exception as e:
            logger.error(f"Failed to download from URL {url}: {str(e)}")
            raise
    
    async def download_from_obs(self, bucket_name: str, object_key: str) -> bytes:
        """Download document from OBS using bucket and key"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            response = self.client.get_object(
                Bucket=bucket_name,
                Key=object_key
            )
            
            file_content = response['Body'].read()
            logger.info(f"Downloaded document from OBS: {bucket_name}/{object_key} ({len(file_content)} bytes)")
            return file_content
            
        except self.client.exceptions.NoSuchKey:
            logger.error(f"Document not found in OBS: {bucket_name}/{object_key}")
            raise FileNotFoundError(f"Document not found: {bucket_name}/{object_key}")
        except Exception as e:
            logger.error(f"Failed to download document from OBS: {str(e)}")
            raise
    
    async def store_parsing_result(self, request_id: str, result: Dict[str, Any]) -> str:
        """Store parsing result in OBS"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            # Convert result to JSON
            result_json = json.dumps(result, default=self._json_serializer, ensure_ascii=False)
            
            # Generate key
            timestamp = datetime.utcnow().strftime("%Y/%m/%d")
            key = f"results/{timestamp}/{request_id}.json"
            
            # Upload to OBS
            self.client.put_object(
                Bucket=self.bucket_name,
                Key=key,
                Body=result_json.encode('utf-8'),
                ContentType='application/json',
                Metadata={
                    'request_id': request_id,
                    'stored_at': datetime.utcnow().isoformat()
                }
            )
            
            logger.info(f"Stored parsing result in OBS: {key}")
            return key
            
        except Exception as e:
            logger.error(f"Failed to store parsing result in OBS: {str(e)}")
            raise
    
    async def get_parsing_result(self, request_id: str) -> Optional[Dict[str, Any]]:
        """Get parsing result from OBS by request ID"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            # Search for result file
            timestamp = datetime.utcnow().strftime("%Y/%m/%d")
            key = f"results/{timestamp}/{request_id}.json"
            
            try:
                response = self.client.get_object(
                    Bucket=self.bucket_name,
                    Key=key
                )
                
                result_json = response['Body'].read().decode('utf-8')
                result = json.loads(result_json)
                logger.info(f"Retrieved parsing result from OBS: {key}")
                return result
                
            except self.client.exceptions.NoSuchKey:
                # Try to find in previous days (up to 7 days)
                for days_ago in range(1, 8):
                    date = (datetime.utcnow() - timedelta(days=days_ago)).strftime("%Y/%m/%d")
                    key = f"results/{date}/{request_id}.json"
                    
                    try:
                        response = self.client.get_object(
                            Bucket=self.bucket_name,
                            Key=key
                        )
                        
                        result_json = response['Body'].read().decode('utf-8')
                        result = json.loads(result_json)
                        logger.info(f"Retrieved parsing result from OBS (archived): {key}")
                        return result
                        
                    except self.client.exceptions.NoSuchKey:
                        continue
                
                logger.warning(f"Parsing result not found for request_id: {request_id}")
                return None
                
        except Exception as e:
            logger.error(f"Failed to get parsing result from OBS: {str(e)}")
            return None
    
    async def list_documents(self, prefix: str = "", max_keys: int = 100) -> list:
        """List documents in OBS bucket"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            response = self.client.list_objects_v2(
                Bucket=self.bucket_name,
                Prefix=prefix,
                MaxKeys=max_keys
            )
            
            documents = []
            for obj in response.get('Contents', []):
                documents.append({
                    'key': obj['Key'],
                    'size': obj['Size'],
                    'last_modified': obj['LastModified'],
                    'etag': obj['ETag']
                })
            
            logger.info(f"Listed {len(documents)} documents with prefix: {prefix}")
            return documents
            
        except Exception as e:
            logger.error(f"Failed to list documents in OBS: {str(e)}")
            return []
    
    async def delete_document(self, key: str) -> bool:
        """Delete document from OBS"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            self.client.delete_object(
                Bucket=self.bucket_name,
                Key=key
            )
            
            logger.info(f"Deleted document from OBS: {key}")
            return True
            
        except Exception as e:
            logger.error(f"Failed to delete document from OBS: {str(e)}")
            return False
    
    async def generate_presigned_url(self, key: str, expires_in: int = 3600) -> str:
        """Generate presigned URL for temporary access"""
        if not self.client:
            raise ValueError("OBS client not initialized")
        
        try:
            url = self.client.generate_presigned_url(
                'get_object',
                Params={'Bucket': self.bucket_name, 'Key': key},
                ExpiresIn=expires_in
            )
            
            logger.info(f"Generated presigned URL for {key}, expires in {expires_in} seconds")
            return url
            
        except Exception as e:
            logger.error(f"Failed to generate presigned URL: {str(e)}")
            raise
    
    def _get_content_type(self, filename: str) -> str:
        """Get content type from filename"""
        extension = os.path.splitext(filename)[1].lower()
        
        content_types = {
            '.pdf': 'application/pdf',
            '.jpg': 'image/jpeg',
            '.jpeg': 'image/jpeg',
            '.png': 'image/png',
            '.tiff': 'image/tiff',
            '.tif': 'image/tiff',
            '.bmp': 'image/bmp',
            '.json': 'application/json',
            '.txt': 'text/plain',
            '.csv': 'text/csv',
        }
        
        return content_types.get(extension, 'application/octet-stream')
    
    def _json_serializer(self, obj):
        """JSON serializer for objects not serializable by default json code"""
        if isinstance(obj, datetime):
            return obj.isoformat()
        raise TypeError(f"Type {type(obj)} not serializable")
    
    def is_ready(self) -> bool:
        """Check if OBS service is ready"""
        return self.client is not None and self.access_key_id and self.secret_access_key
    
    async def health_check(self) -> Dict[str, Any]:
        """Perform health check on OBS service"""
        if not self.is_ready():
            return {
                "status": "unhealthy",
                "error": "OBS client not initialized or credentials missing"
            }
        
        try:
            # Check bucket accessibility
            self.client.head_bucket(Bucket=self.bucket_name)
            
            # Test write/read
            test_key = f"healthcheck/test_{datetime.utcnow().isoformat()}.txt"
            test_content = b"Health check test"
            
            # Write test file
            self.client.put_object(
                Bucket=self.bucket_name,
                Key=test_key,
                Body=test_content,
                ContentType='text/plain'
            )
            
            # Read test file
            response = self.client.get_object(
                Bucket=self.bucket_name,
                Key=test_key
            )
            read_content = response['Body'].read()
            
            # Delete test file
            self.client.delete_object(
                Bucket=self.bucket_name,
                Key=test_key
            )
            
            if read_content == test_content:
                return {
                    "status": "healthy",
                    "bucket": self.bucket_name,
                    "region": self.region,
                    "endpoint": self.endpoint
                }
            else:
                return {
                    "status": "unhealthy",
                    "error": "Read/write test failed"
                }
                
        except Exception as e:
            return {
                "status": "unhealthy",
                "error": str(e)
            }