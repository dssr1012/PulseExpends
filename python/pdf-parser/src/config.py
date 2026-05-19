from pydantic_settings import BaseSettings
from typing import Optional
import os

class Settings(BaseSettings):
    # Server settings
    HOST: str = "0.0.0.0"
    PORT: int = 8000
    DEBUG: bool = False
    
    # OBS settings
    OBS_ENDPOINT: str = "https://obs.la-south-2.myhuaweicloud.com"
    OBS_ACCESS_KEY_ID: str = ""
    OBS_SECRET_ACCESS_KEY: str = ""
    OBS_BUCKET_NAME: str = "pulse-expends-documents"
    OBS_REGION: str = "la-south-2"
    
    # AI/ML settings
    OPENAI_API_KEY: Optional[str] = None
    OPENAI_MODEL: str = "gpt-4"
    OPENAI_MAX_TOKENS: int = 2000
    OPENAI_TEMPERATURE: float = 0.1
    
    # OCR settings
    TESSERACT_PATH: Optional[str] = None
    DEFAULT_LANGUAGE: str = "spa+eng"
    
    # Redis settings (for async processing)
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379
    REDIS_PASSWORD: Optional[str] = None
    REDIS_DB: int = 0
    
    # Celery settings
    CELERY_BROKER_URL: str = "redis://localhost:6379/0"
    CELERY_RESULT_BACKEND: str = "redis://localhost:6379/0"
    
    # Security
    API_KEY: Optional[str] = None
    JWT_SECRET: str = "change-this-in-production"
    JWT_ALGORITHM: str = "HS256"
    JWT_EXPIRATION_MINUTES: int = 30
    
    # Rate limiting
    RATE_LIMIT_PER_MINUTE: int = 60
    RATE_LIMIT_PER_HOUR: int = 1000
    
    # Logging
    LOG_LEVEL: str = "INFO"
    LOG_FORMAT: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    
    # Storage
    MAX_FILE_SIZE_MB: int = 50
    ALLOWED_EXTENSIONS: list = [".pdf", ".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"]
    
    # Processing
    MAX_PAGES_PER_DOCUMENT: int = 50
    MAX_CONCURRENT_PROCESSES: int = 4
    PROCESSING_TIMEOUT_SECONDS: int = 300
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False

# Create settings instance
settings = Settings()

# Validate settings
def validate_settings():
    """Validate critical settings"""
    if not settings.OBS_ACCESS_KEY_ID or not settings.OBS_SECRET_ACCESS_KEY:
        raise ValueError("OBS_ACCESS_KEY_ID and OBS_SECRET_ACCESS_KEY must be set")
    
    if settings.OPENAI_API_KEY and len(settings.OPENAI_API_KEY) < 20:
        raise ValueError("OPENAI_API_KEY appears to be invalid")
    
    # Set TESSERACT_PATH if not provided
    if not settings.TESSERACT_PATH:
        settings.TESSERACT_PATH = "/usr/bin/tesseract"
    
    # Create necessary directories
    os.makedirs("temp", exist_ok=True)
    os.makedirs("logs", exist_ok=True)
    os.makedirs("cache", exist_ok=True)
    
    return True

# Validate on import
try:
    validate_settings()
except Exception as e:
    print(f"Warning: Settings validation failed: {e}")
    print("Some features may not work correctly.")