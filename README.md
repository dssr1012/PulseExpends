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

## 🔧 Development

### Backend Development
```bash
cd services/backend/auth
go run main.go

cd services/backend/mcp-server
go run main.go

cd services/backend/pdf-parser
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python src/main.py
```

### Frontend Development
```bash
cd services/frontend
npm install
npm start
```

### Testing
```bash
cd services/backend
./run_tests.sh

cd services/backend/pdf-parser
pytest tests/
```

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

### Environment Variables
All sensitive configuration is managed through environment variables:
- Database credentials
- API keys
- OAuth credentials
- Encryption keys

### Security Best Practices
- All secrets are stored in environment variables
- Database connections use SSL/TLS
- API endpoints require authentication
- Regular security updates and patches

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

## 📞 Contact

Project Maintainer: [DS SR](https://github.com/dssr1012)

---

**Note**: This is a monorepo combining infrastructure and application code. For the previous separate repositories, see:
- Infrastructure: https://github.com/dssr1012/PulseExpends-Infra (archived)
- Application: https://github.com/dssr1012/PulseExpends (this repository)