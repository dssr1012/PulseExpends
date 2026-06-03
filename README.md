# PulseExpends Monorepo

A unified monorepo containing both infrastructure and application code for the PulseExpends financial management platform.

## 📁 Repository Structure

```
/
├── .github/                    # CI/CD workflows and GitHub Actions
├── infra/                      # Infrastructure as Code (Terraform, scripts)
│   ├── main.tf                 # Main Terraform configuration
│   ├── variables.tf            # Terraform variables
│   ├── rds.tf                  # RDS PostgreSQL module
│   ├── scripts/                # Deployment and management scripts
│   ├── dashboard/              # Monitoring dashboard
│   └── backup/                 # Backup configurations
├── services/                   # Application services
│   ├── backend/                # Backend services
│   │   ├── auth/               # Authentication service (Go)
│   │   ├── mcp-server/         # MCP server (Go)
│   │   ├── pdf-parser/         # PDF parsing service (Python)
│   │   └── config/             # Shared configuration
│   └── frontend/               # Frontend application
│       ├── src/                # Source code
│       ├── public/             # Static assets
│       └── styles/             # CSS styles
├── monitoring/                 # Monitoring configurations
└── tests/                      # Test files
```

## 🚀 Getting Started

### Prerequisites
- Go 1.20+
- Python 3.9+
- Node.js 18+
- Terraform 1.5+
- Docker & Docker Compose

### Quick Start

#### 1. Clone the Repository
```bash
git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends
```

#### 2. Set Up Environment Variables
```bash
cp services/backend/.env.example services/backend/.env
# Edit .env with your configuration
```

#### 3. Run with Docker Compose
```bash
cd services/backend
docker-compose up -d
```

#### 4. Deploy Infrastructure
```bash
cd infra
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init
terraform apply
```

## 🏗️ Infrastructure

### Terraform Modules
The infrastructure is managed using Terraform with the following modules:

- **Compute**: ECS instances and auto-scaling
- **Network**: VPC, subnets, security groups
- **Storage**: OBS buckets, RDS PostgreSQL
- **Security**: KMS keys, IAM policies
- **Monitoring**: Cloud Eye, alarms

### Deployment Scripts
- `infra/deploy.sh` - Full infrastructure deployment
- `infra/reactivate-infrastructure.sh` - Quick infrastructure reactivation
- `infra/setup-ecs-subdomains.sh` - ECS deployment with subdomains

## 🛠️ Services

### Backend Services

#### Authentication Service (`services/backend/auth/`)
- Go-based authentication with Google OAuth
- JWT token management
- User and family circle management
- PostgreSQL database integration

#### MCP Server (`services/backend/mcp-server/`)
- Go-based Model Context Protocol server
- Transaction management
- Financial data processing
- REST API endpoints

#### PDF Parser (`services/backend/pdf-parser/`)
- Python-based PDF processing
- OCR integration for scanned documents
- AI-powered data extraction
- OBS storage integration

### Frontend Application (`services/frontend/`)
- Modern web interface
- Real-time transaction tracking
- Family circle management
- Responsive design

## 📊 Monitoring

### Dashboard
Access the monitoring dashboard at `http://<server-ip>:8081/`

### Health Checks
- Authentication: `GET /api/auth/health`
- MCP Server: `GET /api/mcp/health`
- PDF Parser: `GET /api/pdf/health`

## 🔧 Development Guidelines

### Code Conventions

#### Language
- **English Only:** All code, comments, documentation, and commit messages must be in English
- **Consistent Terminology:** Use standardized technical terms (authentication, authorization, transaction, etc.)
- **Clear Naming:** Descriptive variable/function names following language conventions

#### Go (Backend)
- **Package Structure:** Follow standard Go project layout
- **Error Handling:** Proper error wrapping and logging
- **Testing:** Unit tests for all exported functions
- **Documentation:** Godoc comments for all public APIs

#### TypeScript/React (Frontend)
- **Type Safety:** Strict TypeScript configuration
- **Component Structure:** Functional components with hooks
- **State Management:** Zustand for global state
- **Styling:** Tailwind CSS with consistent design tokens

#### Python (PDF Parser)
- **Type Hints:** Comprehensive type annotations
- **Error Handling:** Proper exception handling with context
- **Documentation:** Docstrings for all functions and classes
- **Testing:** Pytest with comprehensive test coverage

### Git Workflow

#### Branch Strategy
```
main (protected)
├── develop (integration)
├── feature/* (new features)
├── bugfix/* (bug fixes)
└── release/* (release preparation)
```

#### Commit Messages
Follow Conventional Commits specification:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Example: `feat(auth): add Google OAuth integration`

#### Pull Request Process
1. **Create Feature Branch:** `git checkout -b feature/description`
2. **Make Changes:** Follow code conventions
3. **Add Tests:** Include unit/integration tests
4. **Update Documentation:** Update README/docs as needed
5. **Create PR:** Link to issue, add description
6. **Code Review:** Address reviewer comments
7. **Merge:** Squash and merge with descriptive message

### Quality Standards

#### Code Quality
- **Linting:** ESLint for TypeScript, golangci-lint for Go, flake8 for Python
- **Formatting:** Prettier for frontend, gofmt for Go, black for Python
- **Static Analysis:** Regular security and vulnerability scans
- **Complexity:** Keep functions small and focused (max 50 lines)

#### Testing
- **Unit Tests:** > 80% coverage for critical paths
- **Integration Tests:** End-to-end testing for key workflows
- **Performance Tests:** Load testing for high-traffic endpoints
- **Security Tests:** Regular penetration testing

#### Documentation
- **API Documentation:** OpenAPI/Swagger specifications
- **Architecture:** Updated diagrams and decision records
- **Deployment:** Clear deployment and operations guides
- **Troubleshooting:** Common issues and solutions

### Development Environment

#### Local Setup
```bash
# Clone repository
git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends

# Setup backend
cd services/backend
cp .env.example .env
# Edit .env with local configuration
go mod download
go run main.go

# Setup frontend
cd ../frontend
npm install
npm run dev

# Setup PDF parser
cd ../pdf-parser
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python src/main.py
```

#### Docker Development
```bash
# Using Docker Compose
cd services/backend
docker-compose up -d

# Access services
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# PDF Parser: http://localhost:8000
```

#### Testing
```bash
# Backend tests
cd services/backend
go test ./... -v

# Frontend tests
cd ../frontend
npm test

# PDF parser tests
cd ../pdf-parser
pytest tests/ -v
```

### Performance Guidelines

#### Backend (Go)
- **Connection Pooling:** Reuse database connections
- **Caching:** Implement Redis caching for frequent queries
- **Async Processing:** Use goroutines for background tasks
- **Monitoring:** Prometheus metrics for all endpoints

#### Frontend (React)
- **Code Splitting:** Lazy load components
- **Image Optimization:** Use WebP format with responsive images
- **Bundle Size:** Keep bundle under 500KB gzipped
- **Caching:** Service workers for offline support

#### Database (PostgreSQL)
- **Indexing:** Proper indexes on query patterns
- **Partitioning:** Large tables partitioned by date
- **Connection Management:** Pool size based on load
- **Backup Strategy:** Point-in-time recovery

### Security Guidelines

#### Code Security
- **Input Validation:** Validate all user inputs
- **SQL Injection:** Use parameterized queries
- **XSS Protection:** Sanitize all user-generated content
- **CSRF Protection:** Implement anti-CSRF tokens

#### Dependency Management
- **Regular Updates:** Weekly dependency updates
- **Vulnerability Scanning:** Snyk/Dependabot integration
- **License Compliance:** Check all dependencies for compatible licenses
- **Pinning Versions:** Lock dependency versions in production

#### Secrets Management
- **Never Hardcode:** Use environment variables or secret managers
- **Rotation Policy:** Regular credential rotation
- **Access Control:** Least privilege principle
- **Audit Logging:** All secret access logged

For detailed development setup, see [DEVELOPMENT.md](docs/DEVELOPMENT.md).

## 🚢 Deployment

### Manual Deployment
```bash
# Deploy to ECS with paths (recommended for DuckDNS)
cd infra
./deploy-with-paths.sh

# Or with subdomains (requires Cloudflare or custom domain)
./setup-ecs-subdomains.sh

# Configure DuckDNS (only main domain needed)
./scripts/configure-duckdns.sh
```

### Automated Deployment
GitHub Actions workflows are available in `.github/workflows/` for:
- CI/CD pipeline
- Automated testing
- Infrastructure deployment
- Security scanning

### Path-based Deployment (Recommended for DuckDNS)
Since DuckDNS doesn't natively support subdomains, we use path-based routing:

```bash
# Deploy with paths (no subdomains required)
cd infra
./deploy-with-paths.sh
```

**Access URLs with paths:**
- Frontend: `http://pulseexpends.duckdns.org/`
- MCP Server API: `http://pulseexpends.duckdns.org/mcp/`
- PDF Parser API: `http://pulseexpends.duckdns.org/pdf/`
- Status Dashboard: `http://pulseexpends.duckdns.org/status/`

See [Paths Deployment Guide](infra/PATHS-DEPLOYMENT-GUIDE.md) for detailed instructions.

## 📝 Documentation

- [Infrastructure Guide](infra/README.md)
- [Authentication System](services/backend/auth/README.md)
- [Deployment Guide](infra/ECS-DEPLOYMENT-GUIDE.md)
- [RDS PostgreSQL Setup](infra/RDS-README.md)
- [API Documentation](services/backend/README.md)

## 🔐 Security

### Secure Credential Management
The project implements a secure credential management system to prevent accidental exposure of sensitive information:

#### Environment Variables
All sensitive configuration is managed through environment variables:
- Database credentials
- API keys (Huawei Cloud, Google OAuth)
- JWT secrets
- Encryption keys

#### Secure Templates
- `terraform.tfvars.example` - Template with placeholders
- `.env.example` files for each service
- Never commit actual credentials to the repository

#### Setup Script
Use the provided script to configure credentials securely:
```bash
cd infra
export HUAWEI_ACCESS_KEY="your_access_key"
export HUAWEI_SECRET_KEY="your_secret_key"
export HUAWEI_PROJECT_ID="your_project_id"
./setup-credentials.sh
```

#### Git Protection
- `.gitignore` excludes all sensitive files (`*.tfvars`, `*.pem`, `.env`, etc.)
- Pre-commit hooks prevent accidental credential commits
- Regular security scans in CI/CD pipeline

### Authentication & Authorization
- **JWT Tokens:** Secure token-based authentication with refresh tokens
- **OAuth 2.0:** Google OAuth integration for social login
- **RBAC:** Role-based access control for different user types
- **Session Management:** Secure session handling with expiration

### Database Security
- **SSL/TLS:** All database connections use SSL encryption
- **Connection Pooling:** Managed connection pools with limits
- **Credential Encryption:** Database credentials encrypted at rest
- **Regular Backups:** Automated backups with retention policies

### Network Security
- **VPC Isolation:** Services run in private VPC
- **Security Groups:** Strict inbound/outbound rules
- **SSL Termination:** Nginx handles SSL termination
- **Rate Limiting:** Protection against brute force attacks

### Compliance
- **PCI DSS Ready:** No PAN/CVV storage in database
- **Data Encryption:** All sensitive data encrypted at rest and in transit
- **Audit Logging:** Comprehensive audit trails for all operations
- **Regular Updates:** Automated security patches and updates

### Security Best Practices
1. **Least Privilege:** Services run with minimal required permissions
2. **Defense in Depth:** Multiple layers of security controls
3. **Regular Audits:** Security reviews and penetration testing
4. **Incident Response:** Documented procedures for security incidents
5. **Monitoring:** Real-time security monitoring and alerting

For detailed security guidelines, see [SECURE_CREDENTIALS.md](SECURE_CREDENTIALS.md).

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For issues and questions:
1. Check the [documentation](infra/README.md)
2. Search existing [issues](https://github.com/dssr1012/PulseExpends/issues)
3. Create a new issue with detailed information

## 🏗️ Architecture & Scalability

### Current Architecture (Small Scale)

```
┌─────────────────────────────────────────────────────────────┐
│                     Load Balancer (Nginx)                   │
│                     Ports: 80 (HTTP), 443 (HTTPS)          │
└───────────────┬─────────────────────────────────┬───────────┘
                │                                 │
    ┌───────────▼───────────┐         ┌──────────▼───────────┐
    │     Frontend (React)   │         │   Backend (Go)       │
    │     Port: 3000         │         │   Port: 8080-8083    │
    │     - Vite + TypeScript│         │   - 109 API Endpoints│
    │     - Zustand State    │         │   - JWT + OAuth      │
    │     - Tailwind CSS     │         │   - PostgreSQL       │
    └───────────┬───────────┘         └──────────┬───────────┘
                │                                 │
    ┌───────────▼───────────┐         ┌──────────▼───────────┐
    │   Static Assets       │         │   Services Layer      │
    │   - CDN (optional)    │         │   - Auth Service      │
    │   - Cache Headers     │         │   - Transaction Service│
    └───────────────────────┘         │   - Notification Service│
                                      └──────────┬───────────┘
                                                 │
                                      ┌──────────▼───────────┐
                                      │   Data Layer         │
                                      │   - PostgreSQL RDS   │
                                      │   - Redis (optional) │
                                      │   - OBS Storage      │
                                      └───────────────────────┘
```

### Scalability Plan

#### Vertical Scaling (Increase Resources)
1. **ECS Instance:** `c6.large.2` → `c6.xlarge.2` → `c6.2xlarge.2`
2. **RDS PostgreSQL:** `rds.pg.c2.medium` → `rds.pg.c2.large` → `rds.pg.c2.xlarge`
3. **Storage:** ESSD → ESSD Auto-scaling
4. **Memory:** Increase RAM as needed

#### Horizontal Scaling (Add Instances)
1. **Load Balancer:** Nginx → Huawei Cloud ELB
2. **Backend Instances:** 1 → 2-4 ECS instances
3. **Database:** Read replicas + Connection pooling
4. **Cache Layer:** Redis cluster for sessions
5. **CDN:** Huawei Cloud CDN for static assets

#### Microservices Architecture (Future)
1. **Auth Service:** Separate authentication service
2. **PDF Parser Service:** Worker pool + message queue
3. **Notification Service:** Event-driven architecture
4. **API Gateway:** Kong/Traefik for routing
5. **Service Mesh:** Istio/Linkerd for service-to-service communication

### Technology Stack

#### Frontend:
- **Framework:** React 18 + TypeScript
- **Build Tool:** Vite
- **State Management:** Zustand
- **Styling:** Tailwind CSS
- **HTTP Client:** Axios
- **Routing:** React Router DOM

#### Backend:
- **Language:** Go 1.22+
- **Web Framework:** Gin
- **ORM:** GORM
- **Authentication:** JWT + OAuth 2.0 (Google)
- **Database:** PostgreSQL 17
- **Storage:** Huawei Cloud OBS
- **PDF Processing:** pdfplumber + OpenAI

#### Infrastructure:
- **Cloud Provider:** Huawei Cloud
- **Compute:** ECS (Elastic Cloud Server)
- **Database:** RDS PostgreSQL
- **Storage:** OBS (Object Storage Service)
- **Networking:** VPC + Security Groups
- **DNS:** DuckDNS (can migrate to Huawei Cloud DNS)

#### DevOps:
- **Infrastructure as Code:** Terraform
- **Containerization:** Docker (planned)
- **CI/CD:** GitHub Actions (planned)
- **Monitoring:** Prometheus + Grafana (planned)
- **Logging:** ELK Stack (planned)

### Current vs. Scalable Capacity

| **Component** | **Current (Small Scale)** | **Scalable (1000+ Users)** |
|---------------|---------------------------|----------------------------|
| **Frontend** | Single ECS instance | Multiple instances + CDN |
| **Backend** | Monolithic Go app | Microservices + API Gateway |
| **Database** | Single RDS instance | Primary + Read replicas |
| **Cache** | In-memory (Go) | Redis cluster |
| **Storage** | OBS bucket | Multiple OBS buckets + lifecycle |
| **Load Balancing** | Nginx reverse proxy | Huawei Cloud ELB |
| **Monitoring** | Basic logs | Prometheus + Grafana + Alerts |

### Scalability Roadmap

**Phase 1 (Current):** Monolithic, single instance, basic authentication
**Phase 2 (100 users):** Load balancer, read replicas, Redis cache
**Phase 3 (1000 users):** Microservices, API gateway, advanced monitoring
**Phase 4 (10,000+ users):** Kubernetes, service mesh, auto-scaling

### Security Implementation

1. **Credential Management:** Environment variables + secure templates
2. **Authentication:** JWT with refresh tokens + OAuth 2.0
3. **Authorization:** Role-based access control (RBAC)
4. **Database:** SSL connections, encrypted credentials
5. **Network:** Security groups, VPC isolation
6. **Compliance:** No PAN/CVV storage, PCI DSS ready

### Performance Metrics

- **API Response Time:** < 200ms (p95)
- **Database Queries:** < 50ms (p95)
- **Frontend Load Time:** < 2s (First Contentful Paint)
- **Concurrent Users:** 50+ (current), 1000+ (scalable)
- **Uptime:** 99.5% (current), 99.9% (target)

## 📊 Project Status

### ✅ Completed Features
- **Authentication System:** JWT + Google OAuth 2.0 with refresh tokens
- **Transaction Management:** Full CRUD operations with validation
- **PDF Parser:** AI-powered PDF extraction with OCR support
- **Frontend Application:** React SPA with TypeScript and Zustand
- **Infrastructure as Code:** Terraform for Huawei Cloud deployment
- **Secure Credential Management:** Environment-based configuration
- **English-Only Codebase:** Consistent language across all files

### 🚧 In Development
- **Family Groups:** Shared expense tracking within family circles
- **Advanced Reporting:** Financial analytics and visualization
- **Mobile Application:** React Native cross-platform app
- **WebSocket Notifications:** Real-time updates for transactions

### 📅 Planned Features
- **Multi-tenant Architecture:** Support for multiple organizations
- **Advanced Analytics:** Machine learning for spending patterns
- **Invoice Processing:** Automated invoice recognition and categorization
- **Budget Planning:** AI-powered budget recommendations
- **Tax Reporting:** Automated tax calculation and reporting

### 🏗️ Architecture Evolution
- **Current:** Monolithic architecture with clear separation of concerns
- **Phase 2:** Service-oriented architecture with message queues
- **Phase 3:** Microservices with API gateway and service mesh
- **Phase 4:** Kubernetes-based orchestration with auto-scaling

### 📈 Performance Metrics
- **API Response Time:** < 200ms (p95) for all endpoints
- **Database Queries:** < 50ms (p95) for common operations
- **Frontend Load Time:** < 2s First Contentful Paint
- **Concurrent Users:** Supports 50+ users (current), scalable to 1000+
- **Uptime:** 99.5% (current target), 99.9% (future target)

### 🔄 Recent Updates
- **Security Enhancement:** Implemented secure credential management system
- **Language Normalization:** Converted entire codebase to English-only
- **Documentation:** Comprehensive architecture and scalability documentation
- **Infrastructure:** Huawei Cloud deployment with Terraform
- **Code Quality:** Consistent coding standards and conventions

### 🎯 Project Goals
1. **User Experience:** Intuitive interface for financial management
2. **Security:** Enterprise-grade security with compliance
3. **Scalability:** Architecture designed for growth
4. **Maintainability:** Clean code with comprehensive testing
5. **Extensibility:** Modular design for future features

### 🤝 Community & Support
- **GitHub Issues:** Bug reports and feature requests
- **Documentation:** Comprehensive guides and API references
- **Contributing:** Open to community contributions
- **Roadmap:** Public roadmap for transparency

### 📞 Contact & Support
For issues, questions, or contributions:
1. **GitHub Issues:** [Create an issue](https://github.com/dssr1012/PulseExpends/issues)
2. **Documentation:** Check the [docs directory](/docs)
3. **Security:** Report security issues to `security@pulseexpends.com`

---

**Project Maintainer:** [DS SR](https://github.com/dssr1012)

**License:** MIT - See [LICENSE](LICENSE) file for details.

**Note:** This is a monorepo combining infrastructure and application code. For the previous separate repositories, see:
- Infrastructure: https://github.com/dssr1012/PulseExpends-Infra (archived)
- Application: https://github.com/dssr1012/PulseExpends (this repository)