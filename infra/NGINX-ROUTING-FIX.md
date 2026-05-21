# Nginx Routing Fix — PulseExpends

## Problem
Nginx was misconfigured on ECS (`182.160.24.205`):
- Routes `/mcp/*`, `/pdf/*`, `/status/*` returned the frontend instead of proxying to APIs
- Frontend used paths `/mcp` and `/pdf` but Nginx expected `/api/mcp/` and `/api/pdf/`
- `default_server` was not set correctly
- `try_files` was wrong: `try_files / /index.html;` instead of `try_files $uri $uri/ /index.html;`

## Fix Applied
Updated `/etc/nginx/sites-available/pulseexpends`:
- Added `listen 80 default_server;` and `listen [::]:80 default_server;`
- Added both `/mcp/` and `/api/mcp/` routes for compatibility
- Added both `/pdf/` and `/api/pdf/` routes for compatibility
- Added `/api/auth/` and `/api/circles/` proxying to auth backend (port 8082)
- Fixed `try_files $uri $uri/ /index.html;`
- Removed `/etc/nginx/sites-enabled/default` symlink

## Verified Working
- Frontend: `http://182.160.24.205/` (HTTP 200)
- MCP API: `http://182.160.24.205/mcp/health` → `{"status":"healthy"}`
- PDF API: `http://182.160.24.205/pdf/health` → `{"status":"healthy"}`
- Auth API: `http://182.160.24.205/api/auth/login` → JWT returned

## Lessons Learned
1. Always verify that frontend routes match Nginx proxy locations
2. `default_server` is crucial — without it, Nginx falls back to the `default` site
3. `try_files` must use `$uri $uri/ /index.html` for SPA support
4. Keep both `/mcp/` and `/api/mcp/` for backward compatibility
5. SSH access uses `root@` not `ubuntu@` on this ECS instance
