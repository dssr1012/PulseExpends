import logging
import asyncio
import tempfile
import os
from typing import Optional, List
import subprocess
import pytesseract
from PIL import Image
import pdfplumber
import cv2
import numpy as np

from ..config import settings

logger = logging.getLogger(__name__)

class OCRService:
    """Service for OCR text extraction from PDFs and images"""
    
    def __init__(self):
        self.tesseract_path = settings.TESSERACT_PATH
        self.default_language = settings.DEFAULT_LANGUAGE
        self._check_tesseract()
    
    def _check_tesseract(self):
        """Check if Tesseract OCR is available"""
        try:
            subprocess.run([self.tesseract_path, "--version"], 
                          capture_output=True, check=True)
            logger.info(f"Tesseract found at {self.tesseract_path}")
        except (subprocess.CalledProcessError, FileNotFoundError):
            logger.warning(f"Tesseract not found at {self.tesseract_path}. OCR may not work properly.")
    
    async def extract_from_pdf(self, pdf_path: str, language: str = "spa+eng") -> str:
        """Extract text from PDF file"""
        logger.info(f"Extracting text from PDF: {pdf_path}")
        
        try:
            text_parts = []
            
            # Open PDF with pdfplumber
            with pdfplumber.open(pdf_path) as pdf:
                total_pages = len(pdf.pages)
                logger.info(f"PDF has {total_pages} pages")
                
                # Limit pages to prevent excessive processing
                max_pages = min(total_pages, settings.MAX_PAGES_PER_DOCUMENT)
                
                for page_num, page in enumerate(pdf.pages[:max_pages], 1):
                    logger.debug(f"Processing page {page_num}/{max_pages}")
                    
                    # Try to extract text directly first (for text-based PDFs)
                    page_text = page.extract_text()
                    
                    if page_text and len(page_text.strip()) > 50:
                        # Text extraction successful
                        text_parts.append(page_text)
                    else:
                        # Fall back to OCR for scanned PDFs
                        logger.debug(f"Page {page_num} appears to be scanned, using OCR")
                        
                        # Convert page to image
                        image = page.to_image(resolution=300)
                        temp_image_path = tempfile.mktemp(suffix='.png')
                        image.save(temp_image_path, format='PNG')
                        
                        try:
                            # Extract text from image using OCR
                            ocr_text = await self._extract_from_image_file(temp_image_path, language)
                            text_parts.append(ocr_text)
                        finally:
                            # Clean up temporary image file
                            if os.path.exists(temp_image_path):
                                os.unlink(temp_image_path)
            
            # Combine all text
            full_text = "\n\n".join(text_parts)
            
            # Clean up text
            full_text = self._clean_text(full_text)
            
            logger.info(f"Successfully extracted {len(full_text)} characters from PDF")
            return full_text
            
        except Exception as e:
            logger.error(f"Error extracting text from PDF {pdf_path}: {str(e)}", exc_info=True)
            raise
    
    async def extract_from_image(self, image_path: str, language: str = "spa+eng") -> str:
        """Extract text from image file"""
        logger.info(f"Extracting text from image: {image_path}")
        
        try:
            # Preprocess image for better OCR
            processed_image_path = await self._preprocess_image(image_path)
            
            try:
                # Extract text using OCR
                text = await self._extract_from_image_file(processed_image_path, language)
                
                # Clean up text
                text = self._clean_text(text)
                
                logger.info(f"Successfully extracted {len(text)} characters from image")
                return text
                
            finally:
                # Clean up processed image file
                if os.path.exists(processed_image_path) and processed_image_path != image_path:
                    os.unlink(processed_image_path)
                    
        except Exception as e:
            logger.error(f"Error extracting text from image {image_path}: {str(e)}", exc_info=True)
            raise
    
    async def _extract_from_image_file(self, image_path: str, language: str) -> str:
        """Extract text from image file using Tesseract OCR"""
        try:
            # Configure Tesseract
            pytesseract.pytesseract.tesseract_cmd = self.tesseract_path
            
            # Set language
            lang = language if language else self.default_language
            
            # Extract text
            text = pytesseract.image_to_string(
                Image.open(image_path),
                lang=lang,
                config='--psm 3 --oem 3'  # Page segmentation mode 3, OCR Engine mode 3
            )
            
            return text
            
        except Exception as e:
            logger.error(f"OCR failed for {image_path}: {str(e)}")
            raise
    
    async def _preprocess_image(self, image_path: str) -> str:
        """Preprocess image to improve OCR accuracy"""
        try:
            # Read image
            img = cv2.imread(image_path)
            if img is None:
                logger.warning(f"Could not read image {image_path}, skipping preprocessing")
                return image_path
            
            # Convert to grayscale
            gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
            
            # Apply adaptive thresholding
            thresh = cv2.adaptiveThreshold(
                gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
                cv2.THRESH_BINARY, 11, 2
            )
            
            # Denoise
            denoised = cv2.fastNlMeansDenoising(thresh, None, 10, 7, 21)
            
            # Save processed image
            processed_path = tempfile.mktemp(suffix='.png')
            cv2.imwrite(processed_path, denoised)
            
            return processed_path
            
        except Exception as e:
            logger.warning(f"Image preprocessing failed: {str(e)}")
            # Return original path if preprocessing fails
            return image_path
    
    def _clean_text(self, text: str) -> str:
        """Clean and normalize extracted text"""
        if not text:
            return ""
        
        # Remove extra whitespace
        lines = text.split('\n')
        cleaned_lines = []
        
        for line in lines:
            line = line.strip()
            if line:  # Skip empty lines
                # Remove multiple spaces
                line = ' '.join(line.split())
                cleaned_lines.append(line)
        
        # Join lines with proper spacing
        cleaned_text = '\n'.join(cleaned_lines)
        
        # Fix common OCR errors
        replacements = {
            '|': 'I',
            '1': 'I',  # Context-dependent, but common in OCR
            '0': 'O',  # Context-dependent
            '5': 'S',
            '8': 'B',
        }
        
        for wrong, correct in replacements.items():
            cleaned_text = cleaned_text.replace(wrong, correct)
        
        # Remove common noise patterns
        noise_patterns = [
            r'\s+\.\s+',
            r'\s+-\s+-\s+',
            r'\s+\|\s+',
            r'\s+_\s+',
        ]
        
        import re
        for pattern in noise_patterns:
            cleaned_text = re.sub(pattern, ' ', cleaned_text)
        
        return cleaned_text
    
    def detect_language(self, text: str) -> str:
        """Detect language of text (simple detection)"""
        if not text:
            return "unknown"
        
        # Simple language detection based on common words
        spanish_words = ['el', 'la', 'los', 'las', 'de', 'que', 'y', 'en', 'un', 'una', 'por', 'con']
        english_words = ['the', 'and', 'for', 'with', 'from', 'this', 'that', 'have', 'are', 'you']
        
        text_lower = text.lower()
        spanish_count = sum(1 for word in spanish_words if word in text_lower)
        english_count = sum(1 for word in english_words if word in text_lower)
        
        if spanish_count > english_count:
            return "es"
        elif english_count > spanish_count:
            return "en"
        else:
            return "unknown"
    
    def extract_tables(self, pdf_path: str, page_num: int = 0) -> List[List[List[str]]]:
        """Extract tables from PDF"""
        try:
            with pdfplumber.open(pdf_path) as pdf:
                if page_num >= len(pdf.pages):
                    logger.warning(f"Page {page_num} not found in PDF")
                    return []
                
                page = pdf.pages[page_num]
                tables = page.extract_tables()
                
                # Clean table data
                cleaned_tables = []
                for table in tables:
                    cleaned_table = []
                    for row in table:
                        cleaned_row = [cell.strip() if cell else "" for cell in row]
                        cleaned_table.append(cleaned_row)
                    cleaned_tables.append(cleaned_table)
                
                return cleaned_tables
                
        except Exception as e:
            logger.error(f"Error extracting tables from PDF: {str(e)}")
            return []
    
    def extract_form_fields(self, pdf_path: str) -> dict:
        """Extract form fields from PDF"""
        try:
            with pdfplumber.open(pdf_path) as pdf:
                form_fields = {}
                
                for page_num, page in enumerate(pdf.pages):
                    # Extract annotations (form fields are often annotations)
                    if page.annots:
                        for annot in page.annots:
                            if annot.get('T'):  # Field name
                                field_name = annot['T']
                                field_value = annot.get('V', '')
                                form_fields[field_name] = field_value
                
                return form_fields
                
        except Exception as e:
            logger.error(f"Error extracting form fields from PDF: {str(e)}")
            return {}
    
    def is_ready(self) -> bool:
        """Check if OCR service is ready"""
        try:
            # Check if Tesseract is available
            result = subprocess.run(
                [self.tesseract_path, "--version"],
                capture_output=True,
                text=True,
                timeout=5
            )
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError, subprocess.CalledProcessError):
            return False