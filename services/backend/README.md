# PulseExpends Backend

Este directorio contiene el código del backend para el sistema PulseExpends, incluyendo el servidor MCP y el parser de PDFs.

## Estructura

```
backend/
├── mcp-server/           # Servidor MCP (Go)
│   ├── main.go           # Servidor MCP original
│   ├── mcp-server-enhanced.go  # Servidor MCP mejorado con api REST
│   ├── go.mod            # Dependencias de Go
│   └── go.sum            # Checksums de dependencias
├── pdf-parser/           # Parser de PDFs (Python)
│   ├── app.py            # Aplicación FastAPI para parsear PDFs
│   └── requirements.txt  # Dependencias de Python
├── nginx-config/         # configuration de Nginx
│   └── nginx.conf        # configuration de reverse proxy
└── systemd-services/     # Archivos de service systemd
    ├── pulseexpends-mcp-server.service
    └── pulseexpends-pdf-parser.service
```

## Servidor MCP

El servidor MCP proporciona una api REST para gestionar transactiones financieras.

### Endpoints

- `GET /transactions` - get todas las transactiones
- `POST /transactions` - Agregar una nueva transaction
- `GET /summary` - get resumen de gastos por categoría
- `GET /health` - Health check del servidor

### execution

```bash
cd backend/mcp-server
go run mcp-server-enhanced.go
```

El servidor se ejecutará en `http://localhost:8080`.

## PDF Parser

El PDF Parser es una application FastAPI que extrae texto de archivos PDF y analiza transactiones financieras.

### Endpoints

- `POST /parse` - Subir y procesar un archivo PDF
- `GET /health` - Health check del service

### execution

```bash
cd backend/pdf-parser
pip install -r requirements.txt
python app.py
```

El servidor se ejecutará en `http://localhost:8000`.

## configuration de Nginx

El archivo `nginx.conf` configura Nginx como reverse proxy para servir el frontend y enrutar las peticiones a los servicios backend.

## Servicios systemd

Los archivos `.service` permiten ejecutar los servicios como demonios en sistemas Linux.

### Instalación

```bash
sudo cp backend/systemd-services/*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable pulseexpends-mcp-server
sudo systemctl enable pulseexpends-pdf-parser
sudo systemctl start pulseexpends-mcp-server
sudo systemctl start pulseexpends-pdf-parser
```

## Dependencias

### Go (MCP Server)
- gorilla/mux v1.8.1
- rs/cors v1.10.1

### Python (PDF Parser)
- fastapi==0.104.1
- uvicorn[standard]==0.24.0
- pymupdf==1.23.8
- pytesseract==0.3.10
- Pillow==10.1.0
- python-multipart==0.0.6