#!/usr/bin/env python3
from flask import Flask, jsonify, request
import json

app = Flask(__name__)

@app.route('/')
def index():
    return jsonify({
        "status": "ok",
        "service": "PDF Parser",
        "version": "1.0.0",
        "message": "PDF Parser is running. Full implementation coming soon."
    })

@app.route('/health')
def health():
    return jsonify({"status": "healthy"})

@app.route('/api/parse', methods=['POST'])
def parse_pdf():
    # Simulate PDF parsing
    return jsonify({
        "status": "success",
        "message": "PDF parsed successfully (placeholder)",
        "data": {
            "text": "This is a placeholder for PDF parsing functionality.",
            "metadata": {
                "pages": 1,
                "language": "en",
                "size": "A4"
            }
        }
    })

@app.route('/api/ocr', methods=['POST'])
def ocr():
    # Simulate OCR
    return jsonify({
        "status": "success",
        "message": "OCR completed (placeholder)",
        "text": "This is placeholder OCR text."
    })

if __name__ == '__main__':
    print("PDF Parser placeholder starting on :8000")
    app.run(host='0.0.0.0', port=8000, debug=False)
