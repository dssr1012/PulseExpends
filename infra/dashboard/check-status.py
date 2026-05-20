#!/usr/bin/env python3
"""
PulseExpends Infrastructure Status Checker
Checks the status of Huawei Cloud resources and services
"""

import subprocess
import json
import sys
import os
from datetime import datetime

def run_terraform_output():
    """Get Terraform output values"""
    try:
        result = subprocess.run(
            ['./terraform', 'output', '-json'],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.abspath(__file__))
        )
        if result.returncode == 0:
            return json.loads(result.stdout)
        else:
            return {}
    except Exception as e:
        print(f"Error running terraform output: {e}")
        return {}

def check_port(host, port, timeout=5):
    """Check if a port is open"""
    try:
        import socket
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((host, port))
        sock.close()
        return result == 0
    except:
        return False

def check_http_service(host, port, path="/", timeout=5):
    """Check if HTTP service is responding"""
    try:
        import urllib.request
        url = f"http://{host}:{port}{path}"
        req = urllib.request.Request(url, method='HEAD')
        response = urllib.request.urlopen(req, timeout=timeout)
        return response.status < 500
    except:
        return False

def main():
    print("🔍 PulseExpends Infrastructure Status Check")
    print("=" * 50)
    
    # Get Terraform outputs
    print("\n📋 Getting Terraform outputs...")
    outputs = run_terraform_output()
    
    if not outputs:
        print("❌ Could not get Terraform outputs")
        return
    
    # Extract values
    public_ip = outputs.get('ecs_public_ip', {}).get('value', '182.160.24.205')
    domain = outputs.get('application_url', {}).get('value', 'http://pulseexpends.duckdns.org').replace('http://', '')
    
    print(f"\n🌐 Public IP: {public_ip}")
    print(f"🌍 Domain: {domain}")
    print(f"🕐 Check time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    # Check ports
    print("\n🔌 Checking open ports...")
    ports_to_check = [
        (22, 'SSH'),
        (80, 'HTTP'),
        (443, 'HTTPS'),
        (8080, 'MCP Server'),
        (8000, 'PDF Parser')
    ]
    
    for port, service in ports_to_check:
        is_open = check_port(public_ip, port)
        status = "✅ OPEN" if is_open else "❌ CLOSED"
        print(f"  Port {port:4d} ({service:15}): {status}")
    
    # Check HTTP services
    print("\n🌐 Checking HTTP services...")
    http_services = [
        (80, 'Web Server'),
        (8080, 'MCP Server'),
        (8000, 'PDF Parser')
    ]
    
    for port, service in http_services:
        is_responding = check_http_service(public_ip, port)
        status = "✅ RESPONDING" if is_responding else "❌ NOT RESPONDING"
        print(f"  {service:15} (:{port}): {status}")
    
    # Summary
    print("\n" + "=" * 50)
    print("📊 SUMMARY")
    print("=" * 50)
    
    ssh_open = check_port(public_ip, 22)
    http_open = check_port(public_ip, 80)
    mcp_open = check_port(public_ip, 8080)
    
    if not ssh_open:
        print("⚠️  SSH (port 22) is closed. Cannot access instance via SSH.")
        print("   Possible issues:")
        print("   1. Security group not allowing SSH")
        print("   2. Instance not running")
        print("   3. SSH key not properly configured")
    
    if not http_open:
        print("⚠️  HTTP (port 80) is closed. Web services not accessible.")
    
    if ssh_open and not http_open:
        print("\n💡 SUGGESTIONS:")
        print("   1. Check if Docker containers are running on the instance")
        print("   2. Check application logs: docker-compose logs")
        print("   3. Verify security group rules in Huawei Cloud Console")
    
    print("\n🔧 NEXT STEPS:")
    print("   1. Check Huawei Cloud Console for instance status")
    print("   2. Verify security group rules allow ports 22, 80, 443, 8080, 8000")
    print("   3. Check instance system logs for boot/startup errors")
    print("   4. Restart instance if needed: terraform taint huaweicloud_compute_instance.main")
    
    # Generate status JSON for dashboard
    status_data = {
        "timestamp": datetime.now().isoformat(),
        "public_ip": public_ip,
        "domain": domain,
        "ports": {service: check_port(public_ip, port) for port, service in ports_to_check},
        "http_services": {service: check_http_service(public_ip, port) for port, service in http_services},
        "terraform_outputs": outputs
    }
    
    # Save status to file
    with open('infrastructure-status.json', 'w') as f:
        json.dump(status_data, f, indent=2)
    
    print(f"\n💾 Status saved to: infrastructure-status.json")

if __name__ == "__main__":
    main()