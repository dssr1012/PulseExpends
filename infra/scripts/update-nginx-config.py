#!/usr/bin/env python3
import subprocess
import sys

nginx_config = '''server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}'''

# Write config to file
with open('/tmp/nginx-dashboard.conf', 'w') as f:
    f.write(nginx_config)

# Copy to ECS
cmd = f"sshpass -p 'dRZUO9i8NXnAhV' scp -o StrictHostKeyChecking=no /tmp/nginx-dashboard.conf root@182.160.24.205:/etc/nginx/sites-available/dashboard"
result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
if result.returncode != 0:
    print(f"Error copying config: {result.stderr}")
    sys.exit(1)

# Test and reload nginx
cmd = "sshpass -p 'dRZUO9i8NXnAhV' ssh -o StrictHostKeyChecking=no root@182.160.24.205 'nginx -t && systemctl reload nginx'"
result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
if result.returncode != 0:
    print(f"Error reloading nginx: {result.stderr}")
    sys.exit(1)

print("Nginx configuration updated successfully!")