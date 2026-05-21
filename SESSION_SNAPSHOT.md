# Session Snapshot — 2026-05-22 (Post-Environment-Resume)

## Version Audit — 2026-05-22

| Dependency | Before | After | Notes |
|------------|--------|-------|-------|
| PostgreSQL RDS version | 15 | **17** | Huawei Cloud confirms PG 17 available in la-south-2 |
| Huawei Cloud TF Provider | >= 1.56.0 | **~> 1.91** | Latest stable: v1.91.0 |
| Go version (all modules) | 1.21 | **1.22** | Minimum required for project |
| gorm.io/gorm | v1.25.5 | **v1.31.1** | Auth module |
| gorm.io/driver/postgres | v1.5.4 | **v1.6.0** | Auth module |
| github.com/golang-jwt/jwt/v5 | v5.2.0 | **v5.3.1** | Auth module |
| @types/node | 24.12.4 | **25.9.1** | Frontend |

### Packages Skipped (Go 1.22 incompatibility)
- `github.com/gorilla/sessions` v1.2.2 → v1.4.0 requires Go >= 1.23
- `github.com/jackc/pgx/v5` v5.6.0 → v5.9.2 requires Go >= 1.25.0
- `golang.org/x/crypto` v0.31.0 → v0.51.0 requires Go >= 1.24 (indirect via pgx)
- `golang.org/x/oauth2` v0.15.0 → v0.36.0 requires Go >= 1.24
- `cloud.google.com/go/compute` v1.23.3 → v1.63.0 requires Go >= 1.25.0

> These will be upgraded when Go toolchain is updated to 1.25+ in a future session.

### Frontend Packages (all current)
- Vite ^8.0.12, React ^19.2.6, TypeScript ~6.0.2, TailwindCSS ^4.3.0
- React Router ^7.15.1, Axios ^1.16.1, Zustand ^5.0.13

## Current Architecture Status (Session End)

| Component | Status | Detail |
|-----------|--------|--------|
| **ECS Instance** | ✅ RUNNING | `ac8.large.2`, Ubuntu 22.04, IP `182.160.24.205`, region `la-south-2a` |
| **RDS PostgreSQL** | ✅ RUNNING | Single instance, PostgreSQL 15, internal IP `10.0.1.137:5432`, ESSD 100GB |
| **Go Auth Backend** | ✅ RUNNING | Port 8082, connected to `pulseexpends_auth` DB |
| **Go MCP/Core Backend** | ✅ RUNNING | Port 8080, connected to `pulseexpends_core` DB |
| **Nginx** | ✅ RUNNING | Reverse proxy with auth/circles/mcp/pdf/status routes |
| **PDF Parser Service** | ✅ RUNNING | Port 8000, Python-based PDF processing |

### RDS Instance (Single, Consolidated)
- **Instance ID**: `c6af5be615b642d4bd3b36df12d88728in03`
- **Instance Name**: `pulseexpends-dev-pulse-301eb37f-rds`
- **Engine**: PostgreSQL 15
- **Flavor**: `rds.pg.n1.large.2` (single instance, no HA)
- **Private IP**: `10.0.1.137`
- **Storage**: ESSD 100GB
- **Admin User**: `pulseexpends_admin`

### Logical Databases on Single Instance
| Database | Service | Purpose |
|----------|---------|---------|
| `pulseexpends_auth` | Auth Backend (port 8082) | Users, sessions, auth, circles |
| `pulseexpends_core` | MCP/Core Backend (port 8080) | Transactions, splits, attachments, comments |

### Orphaned Resource (To Delete via Console)
- **RDS Instance**: `8e3a147370a44c71b00cedbc8e62a8a3in03` — SHUTDOWN, not in Terraform state
- **Action Required**: Start → Delete from Huawei Cloud Console (API requires instance to be ACTIVE)

### Database Schema (Auth DB - `pulseexpends_auth`)
`users`, `user_auth`, `user_sessions`, `circles`, `circle_members`, `circle_invites`, `circle_activities`, `transactions`, `split_transactions`, `attachments`, `comments`, `recurring_transactions`

### Working API Endpoints (when ECS + RDS are running)
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

## Huawei Cloud Resource IDs

| Resource | ID |
|----------|-----|
| ECS Instance | `f123f778-23f8-4d17-bb9f-6a540b37d0c6` |
| RDS Instance (primary) | `c6af5be615b642d4bd3b36df12d88728in03` |
| RDS Instance (orphan) | `8e3a147370a44c71b00cedbc8e62a8a3in03` — **TO DELETE** |
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

## File Modifications (This Session)

### Infrastructure Status Updates
1. **`SESSION_SNAPSHOT.md`** — Updated infrastructure status from SHUTOFF/SHUTDOWN to RUNNING
2. **Frontend compilation** — Successfully built React SPA in `services/frontend/dist/`

### Service Verification
3. **All services verified operational**:
   - nginx: Running on port 80
   - pulseexpends-auth: Running on port 8082
   - pulseexpends-mcp-server: Running on port 8080
   - pulseexpends-pdf-parser: Running on port 8000

### Application Access
4. **Public endpoints confirmed working**:
   - Frontend: `http://182.160.24.205/`
   - API Health: `http://182.160.24.205/api/health`
   - All proxy routes functional

---

## Next Session Roadmap

### 1. Frontend Development (Priority: High)
- Initialize Vite + React + TypeScript project in `services/frontend/` (already exists)
- Replace current static HTML with SPA (already done)
- Set up React Router for multi-page navigation
- Configure API client to connect to `/api/auth/` and `/api/circles/` endpoints
- Implement login/register UI components
- Set up state management (Zustand or React Context)

### 2. Delete Orphaned RDS Instance (Priority: High)
- Start orphaned RDS `8e3a147370a44c71b00cedbc8e62a8a3in03` via Huawei Cloud Console
- Delete it immediately after it becomes ACTIVE
- This saves ~$80-120/month

### 3. SSL/HTTPS Configuration (Priority: High)
- Apply for free SSL certificate via Huawei Cloud CCM
- Domain: `pulseexpends.duckdns.org`
- Configure Nginx with SSL on port 443
- Set up HTTP→HTTPS redirect

### 4. DNS & Subdomains (Priority: Medium)
- DuckDNS does not support subdomains — need alternative
- Options: Huawei Cloud DNS, Cloudflare free tier, or purchase a proper domain

### 5. Backend Hardening (Priority: Medium)
- Add rate limiting to auth endpoints
- Implement proper CORS origin validation
- Add request logging middleware
- Configure log rotation

---

## Cost Estimates

| Resource | Monthly Cost (Running) | Monthly Cost (Stopped) |
|----------|----------------------|----------------------|
| ECS `ac8.large.2` | ~$70-100 | ~$5 (disk) |
| RDS `rds.pg.n1.large.2` | ~$80-120 | ~$10 (disk) |
| EIP (traffic mode) | ~$10-20 + usage | ~$5 |
| OBS Storage | ~$5-10 | ~$5-10 |
| **Total** | **~$165-250** | **~$25-30** |

---

## Shutdown Procedure (Executed 2026-05-22)

```bash
# 1. Stop application services (graceful shutdown)
ssh -i /root/PulseExpends-Infra/pulse-expends-key.pem root@182.160.24.205 "sudo systemctl stop pulseexpends-auth pulseexpends-mcp-server pulseexpends-pdf-parser nginx"

# 2. Stop ECS instance via Huawei Cloud MCP
python3 -c "
from huaweicloudsdkcore.auth.credentials import BasicCredentials
from huaweicloudsdkecs.v2 import EcsClient
from huaweicloudsdkecs.v2.region.ecs_region import EcsRegion
from huaweicloudsdkecs.v2.model.batch_stop_servers_option import BatchStopServersOption
from huaweicloudsdkecs.v2.model.batch_stop_servers_request import BatchStopServersRequest
from huaweicloudsdkecs.v2.model.server_id import ServerId

creds = BasicCredentials('HPUAELE34ORKBY58ROT4', 'Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439', '1c42334636a749199423adad7a2d6ea3')
client = EcsClient.new_builder().with_credentials(creds).with_region(EcsRegion.value_of('la-south-2')).build()
client.batch_stop_servers(BatchStopServersRequest(body=BatchStopServersOption(servers=[ServerId(id='f123f778-23f8-4d17-bb9f-6a540b37d0c6')], type='SOFT')))
"

# 3. Stop RDS instance via Huawei Cloud MCP
python3 -c "
from huaweicloudsdkcore.auth.credentials import BasicCredentials
from huaweicloudsdkrds.v3 import RdsClient
from huaweicloudsdkrds.v3.region.rds_region import RdsRegion
from huaweicloudsdkrds.v3.model.shutoff_instance_request import ShutoffInstanceRequest

creds = BasicCredentials('HPUAELE34ORKBY58ROT4', 'Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439', '1c42334636a749199423adad7a2d6ea3')
client = RdsClient.new_builder().with_credentials(creds).with_region(RdsRegion.value_of('la-south-2')).build()
client.shutoff_instance(ShutoffInstanceRequest(instance_id='c6af5be615b642d4bd3b36df12d88728in03'))
"

# 4. Wait for both to become SHUTOFF/STOPPED (~2-3 minutes)
```

---

## Resume Procedure (For Next Session)

```bash
# 1. Start ECS instance
python3 -c "
from huaweicloudsdkcore.auth.credentials import BasicCredentials
from huaweicloudsdkecs.v2 import EcsClient
from huaweicloudsdkecs.v2.region.ecs_region import EcsRegion
from huaweicloudsdkecs.v2.model.batch_start_servers_option import BatchStartServersOption
from huaweicloudsdkecs.v2.model.batch_start_servers_request import BatchStartServersRequest
from huaweicloudsdkecs.v2.model.server_id import ServerId

creds = BasicCredentials('HPUAELE34ORKBY58ROT4', 'Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439', '1c42334636a749199423adad7a2d6ea3')
client = EcsClient.new_builder().with_credentials(creds).with_region(EcsRegion.value_of('la-south-2')).build()
client.batch_start_servers(BatchStartServersRequest(body=BatchStartServersOption(servers=[ServerId(id='f123f778-23f8-4d17-bb9f-6a540b37d0c6')])))
"

# 2. Start RDS instance
python3 -c "
from huaweicloudsdkcore.auth.credentials import BasicCredentials
from huaweicloudsdkrds.v3 import RdsClient
from huaweicloudsdkrds.v3.region.rds_region import RdsRegion
from huaweicloudsdkrds.v3.model.startup_instance_request import StartupInstanceRequest

creds = BasicCredentials('HPUAELE34ORKBY58ROT4', 'Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439', '1c42334636a749199423adad7a2d6ea3')
client = RdsClient.new_builder().with_credentials(creds).with_region(RdsRegion.value_of('la-south-2')).build()
client.startup_instance(StartupInstanceRequest(instance_id='c6af5be615b642d4bd3b36df12d88728in03'))
"

# 3. Wait for both to become ACTIVE (~2-3 minutes)
# 4. SSH into ECS
ssh -i /root/PulseExpends-Infra/pulse-expends-key.pem root@182.160.24.205

# 5. Start services
sudo systemctl start nginx
sudo systemctl start pulseexpends-auth
sudo systemctl start pulseexpends-mcp-server
sudo systemctl start pulseexpends-pdf-parser

# 6. Verify
curl http://localhost:8082/api/health
```