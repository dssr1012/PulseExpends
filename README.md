# PulseExpends 💰🏗️

**Modern Family Expense Tracker with Hybrid Go + Python Architecture and Huawei Cloud Deployment**

A scalable, cloud-native expense tracking application built with Go for high-performance backend operations and Python for AI-powered document parsing. Designed to scale from 2 users to millions on Huawei Cloud.

## 🏗️ Architecture Overview

### **Hybrid Language Architecture**
- **Go (Golang)**: Core backend, MCP server, API, and business logic
- **Python**: AI-powered PDF/image parsing microservice
- **Repository Pattern**: Abstract data layer for evolutionary scaling

### **Cloud Infrastructure**
- **Primary Region**: Huawei Cloud Santiago, Chile (`la-south-2`)
- **Phase 1**: OBS (Object Storage Service) for JSON ledger storage
- **Phase 2**: Seamless migration to GaussDB/PostgreSQL or GeminiDB/MongoDB
- **Compute**: ECS/CCI containers with CCE/ELB scaling path

## 📁 Repository Structure

### **Application Repository (`dssr1012/PulseExpends`)**
```
pulse-expends/
├── cmd/
│   └── mcp-server/          # Go MCP Server entry point
├── internal/
│   ├── repository/          # Repository pattern interfaces
│   ├── service/            # Business logic services
│   ├── handler/            # HTTP/API handlers
│   ├── middleware/         # HTTP middleware
│   ├── model/             # Data models
│   └── mcp/               # MCP tool definitions
├── pkg/
│   ├── obs/               # OBS storage implementation
│   ├── postgres/          # PostgreSQL implementation
│   ├── mongodb/           # MongoDB implementation
│   └── utils/             # Shared utilities
├── python/
│   └── pdf-parser/        # Python AI parsing microservice
├── api/                   # OpenAPI/Swagger definitions
├── configs/               # Configuration files
├── deployments/
│   ├── k8s/              # Kubernetes manifests
│   └── docker/           # Docker configurations
├── scripts/              # Migration and deployment scripts
└── docs/                 # Documentation
```

### **Infrastructure Repository (`dssr1012/PulseExpends-Infra`)**
```
pulse-expends-infra/
├── modules/
│   ├── network/          # VPC, subnets, security groups
│   ├── compute/          # ECS, CCI, CCE configurations
│   ├── storage/          # OBS buckets, databases
│   └── monitoring/       # Logging, metrics, alerts
├── environments/
│   ├── dev/             # Development environment
│   ├── staging/         # Staging environment
│   └── prod/            # Production environment
└── terraform.tfvars     # Environment variables
```

## 🚀 Core Features

### **Family Expense Management**
- **Family Circles**: Shared real-time expense and income tracking
- **Multi-Method Ledger**: Cash, debit, and credit card tracking
- **Real-time Collaboration**: Synchronized updates across family members

### **AI-Powered Document Processing**
- **Credit Card Statement Parsing**: PDF/Image to structured data
- **Installment Detection**: Automatic installment plan identification
- **Transaction Reconciliation**: Cross-reference with existing ledger data
- **Anomaly Detection**: Identify irregular or unlisted transactions

### **MCP (Model Context Protocol) Integration**
- **OpenClaw Assistant Tools**: Natural language expense tracking
- **Real-time Updates**: MCP server provides live data to AI assistants
- **Secure Operations**: Authentication and authorization built-in

## 🛠️ Technology Stack

### **Backend (Go)**
- **Framework**: Chi Router / Gin / Fiber
- **Database**: Repository pattern with OBS (Phase 1), PostgreSQL/MongoDB (Phase 2)
- **Authentication**: JWT with Huawei Cloud IAM integration
- **MCP Server**: Custom MCP server for OpenClaw integration
- **Validation**: Go-validator with custom financial rules

### **AI Microservice (Python)**
- **Framework**: FastAPI
- **OCR**: Tesseract, EasyOCR, or Huawei OCR Service
- **PDF Processing**: PyPDF2, pdfplumber
- **AI/ML**: Transformers for NLP, Custom models for financial data
- **Queue Processing**: Redis/Celery for async processing

### **Infrastructure (Huawei Cloud)**
- **Compute**: Elastic Cloud Server (ECS), Container Engine (CCE)
- **Storage**: Object Storage Service (OBS), GaussDB, GeminiDB
- **Networking**: VPC, ELB, NAT Gateway, VPN
- **Security**: IAM, Security Groups, Cloud Eye monitoring
- **CI/CD**: CodeArts, SWR (Container Registry)

## 📊 Data Architecture

### **Phase 1: OBS-Based Storage**
```json
{
  "family_circles": {
    "circle_id": "uuid",
    "name": "Family Name",
    "members": ["user_id1", "user_id2"],
    "settings": {...},
    "created_at": "timestamp",
    "updated_at": "timestamp"
  },
  "transactions": {
    "transaction_id": "uuid",
    "circle_id": "uuid",
    "user_id": "uuid",
    "amount": 100.50,
    "currency": "CLP",
    "category": "groceries",
    "payment_method": "credit_card",
    "description": "Supermarket purchase",
    "date": "2024-01-15",
    "metadata": {...},
    "created_at": "timestamp"
  }
}
```

### **Phase 2: Database Migration Path**
```go
// Repository interface allows seamless storage migration
type TransactionRepository interface {
    Create(ctx context.Context, transaction *Transaction) error
    FindByID(ctx context.Context, id string) (*Transaction, error)
    FindByCircle(ctx context.Context, circleID string, filters Filter) ([]Transaction, error)
    Update(ctx context.Context, transaction *Transaction) error
    Delete(ctx context.Context, id string) error
}
```

## 🔧 Getting Started

### **Prerequisites**
- Go 1.21+
- Python 3.10+
- Docker & Docker Compose
- Huawei Cloud Account (la-south-2 region)
- Terraform 1.5+

### **Local Development**
```bash
# Clone the repository
git clone git@github.com:dssr1012/PulseExpends.git
cd PulseExpends

# Start development environment
docker-compose up -d

# Run Go server
go run cmd/mcp-server/main.go

# Run Python parser
cd python/pdf-parser
uvicorn src.main:app --reload

# Access services
# MCP Server: http://localhost:8080
# Python API: http://localhost:8000
# API Docs: http://localhost:8000/docs
```

### **Huawei Cloud Deployment**
```bash
# Clone infrastructure repository
git clone git@github.com:dssr1012/PulseExpends-Infra.git
cd PulseExpends-Infra

# Initialize Terraform
terraform init

# Plan deployment
terraform plan -var-file=environments/dev.tfvars

# Apply infrastructure
terraform apply -var-file=environments/dev.tfvars
```

## 🧪 Testing

```bash
# Run Go tests
go test ./internal/... -v

# Run Python tests
cd python/pdf-parser
pytest tests/ -v

# Integration tests
go test ./tests/integration/...

# Load testing
k6 run scripts/load-test.js
```

## 🔒 Security

### **Data Protection**
- End-to-end encryption for sensitive data
- Huawei Cloud KMS for key management
- OBS server-side encryption
- TLS 1.3 for all communications

### **Authentication & Authorization**
- JWT with short-lived tokens
- Huawei Cloud IAM integration
- Role-based access control (RBAC)
- Audit logging for all operations

### **Compliance**
- GDPR compliance for EU users
- Financial data protection standards
- Regular security audits
- Penetration testing

## 📈 Scaling Strategy

### **Phase 1 (2-100 users)**
- Single ECS instance for Go MCP server
- Python microservice as separate container
- OBS for JSON storage
- Basic monitoring with Cloud Eye

### **Phase 2 (100-10,000 users)**
- CCE (Container Engine) with auto-scaling
- ELB (Elastic Load Balancer) distribution
- GaussDB for PostgreSQL for relational data
- GeminiDB for MongoDB for document storage
- Advanced monitoring and alerting

### **Phase 3 (10,000+ users)**
- Multi-AZ deployment for high availability
- Read replicas for database scaling
- Redis cache for frequent queries
- CDN for static assets
- Advanced AI/ML pipeline for predictions

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### **Development Guidelines**
- Follow Go best practices and effective-go
- Use Python type hints and mypy
- Write comprehensive tests
- Update documentation
- Follow commit message conventions

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Huawei Cloud](https://www.huaweicloud.com/) for infrastructure
- [OpenClaw](https://openclaw.ai/) for MCP integration
- [Go](https://go.dev/) for high-performance backend
- [FastAPI](https://fastapi.tiangolo.com/) for Python microservices
- [Terraform](https://www.terraform.io/) for infrastructure as code

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/dssr1012/PulseExpends/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dssr1012/PulseExpends/discussions)
- **Documentation**: [Project Wiki](https://github.com/dssr1012/PulseExpends/wiki)

## 🌟 Roadmap

### **Q1 2024** - MVP Release
- [ ] Core Go MCP server
- [ ] Python PDF parser microservice
- [ ] OBS storage implementation
- [ ] Basic family circle management
- [ ] Huawei Cloud deployment

### **Q2 2024** - Feature Complete
- [ ] Advanced transaction categorization
- [ ] Budget planning and alerts
- [ ] Multi-currency support
- [ ] Mobile-responsive web interface
- [ ] Database migration capability

### **Q3 2024** - Scaling
- [ ] GaussDB/PostgreSQL integration
- [ ] Advanced AI/ML features
- [ ] Real-time notifications
- [ ] API rate limiting
- [ ] Advanced security features

### **Q4 2024** - Enterprise Ready
- [ ] Multi-tenant support
- [ ] Advanced reporting and analytics
- [ ] Third-party integrations
- [ ] Mobile applications
- [ ] SOC 2 compliance

---

**Built with ❤️ by [DS SR](https://github.com/dssr1012) for modern family finance management**