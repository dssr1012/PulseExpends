#!/usr/bin/env python3
"""
Simple HTTP server to serve the PulseExpends dashboard
"""

import http.server
import socketserver
import os
import json
from datetime import datetime
import threading
import time

PORT = 8081
DASHBOARD_FILE = "dashboard.html"
STATUS_FILE = "infrastructure-status.json"

class DashboardHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        # Serve dashboard.html for root path
        if self.path == '/' or self.path == '/dashboard':
            self.path = '/' + DASHBOARD_FILE
        
        # Serve status.json for /status endpoint
        elif self.path == '/status':
            self.send_status()
            return
        
        # Serve API endpoint for checking services
        elif self.path.startswith('/api/check/'):
            self.check_service(self.path.replace('/api/check/', ''))
            return
        
        # Default to file serving
        return http.server.SimpleHTTPRequestHandler.do_GET(self)
    
    def send_status(self):
        """Send current infrastructure status as JSON"""
        try:
            with open(STATUS_FILE, 'r') as f:
                status_data = json.load(f)
        except:
            status_data = {
                "error": "Status file not found",
                "timestamp": datetime.now().isoformat()
            }
        
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(status_data, indent=2).encode())
    
    def check_service(self, service):
        """Check a specific service and return status"""
        import socket
        
        services = {
            'ssh': ('182.160.24.205', 22),
            'http': ('182.160.24.205', 80),
            'mcp': ('182.160.24.205', 8080),
            'pdf': ('182.160.24.205', 8000)
        }
        
        if service not in services:
            self.send_error(404, f"Service {service} not found")
            return
        
        host, port = services[service]
        is_open = self.is_port_open(host, port)
        
        status = {
            "service": service,
            "host": host,
            "port": port,
            "status": "open" if is_open else "closed",
            "timestamp": datetime.now().isoformat()
        }
        
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(status, indent=2).encode())
    
    def is_port_open(self, host, port, timeout=2):
        """Check if a port is open"""
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(timeout)
            result = sock.connect_ex((host, port))
            sock.close()
            return result == 0
        except:
            return False
    
    def log_message(self, format, *args):
        """Override to reduce log noise"""
        pass

def update_status_periodically():
    """Update status file periodically in background"""
    while True:
        try:
            # Run the status checker
            os.system(f"cd {os.path.dirname(os.path.abspath(__file__))} && python3 check-status.py > /dev/null 2>&1")
        except:
            pass
        time.sleep(60)  # Update every minute

def main():
    # Change to the script directory
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    
    # Start background thread to update status
    updater = threading.Thread(target=update_status_periodically, daemon=True)
    updater.start()
    
    # Start HTTP server
    with socketserver.TCPServer(("", PORT), DashboardHandler) as httpd:
        print(f"🚀 Dashboard server started at http://localhost:{PORT}")
        print(f"📊 Dashboard: http://localhost:{PORT}/dashboard")
        print(f"📈 Status API: http://localhost:{PORT}/status")
        print(f"🔍 Service checks:")
        print(f"   - SSH: http://localhost:{PORT}/api/check/ssh")
        print(f"   - HTTP: http://localhost:{PORT}/api/check/http")
        print(f"   - MCP: http://localhost:{PORT}/api/check/mcp")
        print(f"   - PDF: http://localhost:{PORT}/api/check/pdf")
        print(f"\n📱 Access from other devices on the same network:")
        print(f"   Use your machine's IP address instead of localhost")
        print(f"\n🛑 Press Ctrl+C to stop the server")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n🛑 Server stopped")

if __name__ == "__main__":
    main()