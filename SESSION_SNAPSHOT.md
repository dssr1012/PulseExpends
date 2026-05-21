# Session Snapshot — 2026-05-21

## Current Architecture Status

| Component | Status | Detail |
|-----------|--------|--------|
| **ECS Instance** | ✅ Running | `ac8.large.2`, Ubuntu 22.04, IP `182.160.24.205`, region `la-south-2a` |
| **RDS PostgreSQL** | ✅ Running | PostgreSQL 15.14, flavor `rds.pg.n1.large.2.ha` (single instance), internal IP `10.0.1.137:5432`, storage ESSD 40GB |
| **Go Auth Backend** | ✅ Running | `pulseexpends-auth` on port 8082, systemd service enabled |
| **Nginx** | ✅ Running | Reverse proxy on port 80, serves frontend + proxies API routes |
| **Frontend** | ✅ Served | Static HTML at `/opt/PulseExpends/services/frontend/` |
| **MCP Server** | Unknown | Port 8080 (Go REST API for transactions) |
| **PDF Parser** | Unknown | Port 8000 (Python/Flask) |
| **Status Dashboard** | Unknown | Port 8081 (Python HTTP server) |

### Database Schema (12 tables)
`users`, `user_auth`, `user_sessions`, `circles`, `circle_members`, `circle_invites`, `circle_activities`, `transactions`, `split_transactions`, `attachments`, `comments`, `recurring_transactions`

### Working API Endpoints (via `http://182.160.24.205`)
- `POST /api/auth/register` — Create account (returns JWT + user)
- `POST /api/auth/login` — Login (returns JWT + user)
- `GET  /api/health` — Health check (on port 8082 directly)
- `POST /api/circles/*` — Circle management (proxied to 8082)

### Nginx Proxy Routes
| Route | Backend |
|-------|---------|
| `/api/auth/` | `http://127.0.0.1:8082/api/auth/` |
| `/api/circles/` | `http://127.0.0.1:8082/api/circles/` |
| `/mcp/`, `/api/mcp/` | `http://localhost:8080/` |
| `/pdf/`, `/api/pdf/` | `http://localhost:8000/` |
| `/status/`, `/api/status/` | `http://localhost:8081/` |
| `/` | Frontend static files (`try_files`) |

---

## Exact File Modifications (This Session)

### Go Backend — `/root/PulseExpends/services/backend/auth/`
1. **`models/user.go`** — Added `TableName()` methods:
   - `UserAuth.TableName() → "user_auth"` (GORM was pluralizing to `user_auths`)
   - `UserSession.TableName() → "user_sessions"`
2. **`handlers/auth.go`** — Rewrote to fix model mismatches:
   - `UserAuth.PasswordHash` instead of `User.Password`
   - `*time.Time` for `LastLoginAt`
   - `UserSession` without `IsActive`/`LastActivityAt` (uses `ExpiresAt`/`LastUsedAt`)
   - Removed `models.JSONB` references
3. **`handlers/circles.go`** — Multiple fixes:
   - Renamed `generateRandomString` → `generateCircleCode` (duplicate symbol)
   - Replaced `models.JSONB` with `models.CircleSettings`
   - Fixed string→int conversion for limit/offset using `fmt.Sscanf`
   - Removed nil check on struct type `req.Settings`
   - Added `fmt` import
4. **`middleware/auth.go`** — Fixed `session.LastActivityAt` → `session.LastUsedAt`
5. **`main.go`** — Fixed:
   - Removed unused `context` import
   - `models.TransactionSplit` → `models.SplitTransaction`
   - `middleware.AuthMiddleware` → `middleware.AuthMiddleware(db)`

### ECS Instance Configuration
6. **`/etc/nginx/sites-available/pulseexpends`** — Added `/api/auth/` and `/api/circles/` proxy locations
7. **`/etc/systemd/system/pulseexpends-auth.service`** — Created systemd service for auth backend
8. **`/opt/pulseexpends-auth/.env`** — Environment file with DATABASE_URL, JWT_SECRET, etc.

### RDS Database
9. **Dropped views** `user_transaction_summaries` and `circle_summaries` (blocked GORM AutoMigrate ALTER TABLE)
10. **Granted schema privileges** to `pulseexpends_admin` on `public` schema

---

## Huawei Cloud Resource IDs

| Resource | ID |
|----------|-----|
| ECS Instance | `pulseexpends-dev-pulse-301eb37f-ecs` (name) |
| RDS Instance | `pulseexpends-rds-postgresql` (name) — need to verify exact ID |
| EIP | `94e1e7d5-0384-43eb-ab4b-f842255166be` → `182.160.24.205` |
| VPC | `14432977-f91e-4b8a-b7cc-0da95b0dfc45` |
| Security Group | `6e43ebe9-d7ba-46cc-9fec-c511655f98f3` |
| Enterprise Project | `9d731c5b-e130-430b-b88d-66f312596926` |
| OBS Bucket | `pulseexpends-data-dev-pulse-301eb37f` |

### Credentials
- **RDS Admin**: `pulseexpends_admin` / `pptKH9g8dWXDYnPKdCRTJzY46COSvLI`
- **RDS Internal Endpoint**: `10.0.1.137:5432`
- **SSH Key**: `/root/PulseExpends-Infra/pulse-expends-key.pem`
- **Project ID**: `1c42334636a749199423adad7a2d6ea3`

---

## Next Session Roadmap

### 1. Frontend Initialization (Priority: High)
- Initialize Vite + React + TypeScript project in `services/frontend/`
- Replace current static HTML with SPA
- Set up React Router for multi-page navigation
- Configure API client to connect to `/api/auth/` and `/api/circles/` endpoints
- Implement login/register UI components
- Set up state management (Zustand or React Context)

### 2. SSL/HTTPS Configuration (Priority: High)
- Apply for free SSL certificate via Huawei Cloud CCM (Certificate Manager)
- Domain: `pulseexpends.duckdns.org`
- Configure Nginx with SSL on port 443
- Set up HTTP→HTTPS redirect
- Consider wildcard cert for `*.pulseexpends.duckdns.org` if subdomains needed

### 3. DNS & Subdomains (Priority: Medium)
- DuckDNS does not support subdomains — need alternative:
  - Option A: Use Huawei Cloud DNS service
  - Option B: Use Cloudflare free tier
  - Option C: Purchase a proper domain
- Required subdomains: `api.`, `auth.`, `pdf.`, `status.`

### 4. Backend Hardening (Priority: Medium)
- Add rate limiting to auth endpoints
- Implement proper CORS origin validation
- Add request logging middleware
- Set up health check monitoring
- Configure log rotation for journalctl

### 5. CI/CD Pipeline (Priority: Low)
- GitHub Actions for automated builds
- Automated deployment to ECS on push to main
- Database migration automation
- Staging environment setup

### 6. Additional Features (Priority: Low)
- Google OAuth integration (endpoints exist, needs client credentials)
- MFA/TOTP implementation
- Password reset flow
- Email verification
- Circle invitation system testing

---

## Cost Estimates (When Running)

| Resource | Monthly Cost (USD) |
|----------|-------------------|
| ECS `ac8.large.2` | ~$70-100 |
| RDS `rds.pg.n1.large.2.ha` | ~$80-120 |
| EIP (traffic mode) | ~$10-20 + usage |
| OBS Storage | ~$5-10 |
| **Total (running)** | **~$165-250** |
| **Total (stopped)** | **~$5-15** (EVS disk + OBS only) |

---

## Resume Procedure

```bash
# 1. Start ECS instance (Huawei Cloud Console or API)
# 2. Start RDS instance (Huawei Cloud Console or API)
# 3. Wait for both to become ACTIVE (~2-3 minutes)
# 4. SSH into ECS
ssh -i /root/PulseExpends-Infra/pulse-expends-key.pem root@182.160.24.205
# 5. Start services
sudo systemctl start nginx
sudo systemctl start pulseexpends-auth
# 6. Verify
curl http://localhost:8082/api/health
curl -X POST http://localhost:8082/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"donnie@pulseexpends.com","password":"TestPass123!"}'
```
