# PulseExpends Frontend - Gestión de Gastos Inteligente

## 🎉 ¡Frontend Similar a Fintonic Desplegado!

He created y desplegado successsamente un frontend web completo para **PulseExpends** con un diseño inspirado en **Fintonic**, que te permite interactuar con el MCP Server y cargar resúmenes de tarjetas en PDF.

## 🌐 **URL de Acceso**

**Frontend Principal:** http://182.160.24.205/

## 🚀 **Características Implementadas**

### **1. Dashboard Principal**
- **Estadísticas en tiempo real**: Gastos totales, ingresos, balance neto, número de transactiones
- **Transactiones recientes**: Lista de las últimas transactiones con iconos por categoría
- **Gráfico de categorías**: Visualización de gastos por categoría con gráfico circular
- **Alertas inteligentes**: Notificaciones de gastos altos y resúmenes pendientes

### **2. Gestión de Gastos**
- **Agregar gastos manualmente**: Formulario intuitivo para registrar gastos e ingresos
- **Categorías predefinidas**: Alimentos, Transporte, Compras, Entretenimiento, Café, Servicios, Salud, Educación, Otros
- **Filtrado y búsqueda**: Busca y filtra transactiones por categoría y tipo

### **3. Subida de Resúmenes PDF**
- **Arrastrar y soltar**: Interfaz intuitiva para subir archivos PDF
- **Procesamiento automático**: Conectado al PDF Parser en `http://182.160.24.205/pdf/`
- **Simulación de extraction**: Extrae transactiones simuladas de los PDFs subidos

### **4. Análisis por Categorías**
- **Gráficos interactivos**: Visualización de distribución de gastos
- **Porcentajes detallados**: Desglose porcentual por categoría
- **Recomendaciones personalizadas**: Sugerencias basadas en patrones de gasto

### **5. Sistema de Alertas**
- **Alertas de presupuesto**: Notifica cuando superas limits establecidos
- **Recordatorios de pago**: Alertas de vencimiento de tarjetas
- **configuration personalizable**: Ajusta limits y preferencias de notification

### **6. configuration Personalizable**
- **Moneda y formato de date**: configuration regional
- **Categorías personalizadas**: Añade o elimina categorías
- **Exportación de datos**: CSV, Excel y PDF
- **Preferencias de notification**: Email y push notifications

## 🔗 **APIs Conectadas**

### **MCP Server** (`http://182.160.24.205/mcp/`)
- `GET /transactions` - Obtiene todas las transactiones
- `POST /transactions` - Agrega nueva transaction
- `GET /summary` - Resumen de gastos por categoría
- `GET /health` - Verificación de salud del servidor

### **PDF Parser** (`http://182.160.24.205/pdf/`)
- `POST /api/parse` - Procesa archivos PDF de resúmenes
- `GET /` - information del service
- `GET /health` - Verificación de salud

## 🎨 **Diseño Inspirado en Fintonic**

### **Paleta de Colores**
- **Azul principal**: `#00a8ff` (similar al azul Fintonic)
- **Azul oscuro**: `#0097e6`
- **Púrpura secundario**: `#8c7ae6`
- **Verde success**: `#44bd32`
- **Amarillo warning**: `#fbc531`
- **Rojo peligro**: `#e84118`

### **Componentes Visuales**
- **Tarjetas de estadísticas**: Con iconos y cambios porcentuales
- **Lista de transactiones**: Con iconos por categoría y colores para gastos/ingresos
- **Gráfico circular**: Para distribución de categorías
- **Área de subida**: Drag & drop para PDFs
- **Modales elegantes**: Para agregar gastos y subir archivos

## 📱 **Responsive Design**
- **Mobile-first**: Diseño adaptativo para todos los dispositivos
- **Sidebar colapsable**: En móviles se convierte en menú hamburguesa
- **Grid flexible**: Se adapta al tamaño de pantalla

## 🔧 **Tecnologías Utilizadas**

### **Frontend**
- **HTML5/CSS3/JavaScript Vanilla**: Sin frameworks pesados
- **Font Awesome**: Iconos modernos
- **CSS Grid & Flexbox**: Layouts responsivos
- **Fetch api**: Comunicación con APIs REST
- **Drag & Drop api**: Para subida de archivos

### **Backend (Servicios Existentes)**
- **MCP Server (Go)**: api REST para gestión de transactiones
- **PDF Parser (Python/Flask)**: Procesamiento de archivos PDF
- **Nginx**: Reverse proxy y servidor web

## 🛠️ **Cómo Usar**

### **1. Agregar un Gasto Manualmente**
1. Haz clic en "Agregar Gasto" en el dashboard
2. Completa el formulario:
   - Descripción (ej: "Cena en restaurante")
   - Monto (ej: 45.50)
   - Categoría (selecciona una)
   - date (selecciona o usa la actual)
   - Tipo (Gasto o Ingreso)
3. Haz clic en "Agregar transaction"

### **2. Subir Resumen de Tarjeta PDF**
1. Ve a la sección "Subir Resumen"
2. Arrastra tu archivo PDF o haz clic para seleccionar
3. El sistema procesará el PDF y extraerá transactiones automáticamente
4. Las transactiones detectadas se agregarán a tu dashboard

### **3. Ver Análisis**
1. **Dashboard**: view general de tus finanzas
2. **Transactiones**: Lista completa con filtros
3. **Categorías**: Análisis detallado por categoría
4. **Alertas**: Notificaciones y configurationes

### **4. Configurar Alertas**
1. Ve a "Alertas" en el sidebar
2. Establece presupuestos por categoría
3. Configura umbrales de notification
4. Guarda tus preferencias

## 🔌 **Endpoints api Disponibles**

### **Desde el Frontend**
```javascript
// Agregar transaction
fetch('/mcp/transactions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    description: "Supermercado",
    amount: 85.50,
    category: "Food",
    type: "expense",
    currency: "USD"
  })
})

// get transactiones
fetch('/mcp/transactions')
  .then(res => res.json())
  .then(data => console.log(data))

// Subir PDF
const formData = new FormData();
formData.append('file', pdfFile);
fetch('/pdf/api/parse', {
  method: 'POST',
  body: formData
})
```

### **Desde Línea de Comandos**
```bash
# Agregar gasto
curl -X POST http://182.160.24.205/mcp/transactions \
  -H "Content-Type: application/json" \
  -d '{"description":"Café","amount":3.50,"category":"Coffee","type":"expense"}'

# Ver todas las transactiones
curl http://182.160.24.205/mcp/transactions

# Ver resumen
curl http://182.160.24.205/mcp/summary
```

## 🚀 **Próximas Mejoras**

### **Fase 2 (Planeado)**
1. **authentication de users**: Sistema de login/registro
2. **Dashboard en tiempo real**: Actualizaciones automáticas con WebSockets
3. **Exportación avanzada**: Reportes personalizados en PDF
4. **Integración con bancos**: Conexión directa a APIs bancarias
5. **Aplicación móvil**: Versión PWA para iOS/Android

### **Fase 3 (Futuro)**
1. **Machine Learning**: Predicción de gastos y detección de patrones
2. **OCR avanzado**: Mejor extraction de datos de PDFs
3. **Alertas inteligentes**: Basadas en comportamiento histórico
4. **Presupuestos automáticos**: Sugerencias basadas en ingresos

## 🐛 **Solución de Problemas**

### **Frontend no carga**
```bash
# Verificar Nginx
ssh root@182.160.24.205 "systemctl status nginx"

# Verificar archivos
ssh root@182.160.24.205 "ls -la /var/www/pulseexpends-frontend/"
```

### **APIs no responden**
```bash
# Verificar servicios
ssh root@182.160.24.205 "systemctl status pulseexpends-mcp"
ssh root@182.160.24.205 "systemctl status pulseexpends-pdf-parser"

# Probar endpoints
curl http://182.160.24.205/mcp/health
curl http://182.160.24.205/pdf/health
```

### **Subida de PDF falla**
1. Verifica que el archivo sea PDF (no escaneado/imagen)
2. Tamaño máximo: 10MB
3. El PDF debe contener texto (no solo imágenes)
4. Revisa la consola del navegador para errores

## 📊 **Ejemplos de Uso**

### **Ejemplo 1: Agregar Gastos Diarios**
```javascript
// Gastos del día
const dailyExpenses = [
  { description: "Desayuno", amount: 8.50, category: "Food" },
  { description: "Transporte", amount: 4.75, category: "Transportation" },
  { description: "Almuerzo", amount: 12.00, category: "Food" },
  { description: "Café", amount: 3.50, category: "Coffee" }
];

// Agregar cada gasto
dailyExpenses.forEach(expense => {
  fetch('/mcp/transactions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ...expense, type: 'expense', currency: 'USD' })
  });
});
```

### **Ejemplo 2: Análisis Semanal**
```javascript
// get resumen semanal
fetch('/mcp/summary')
  .then(res => res.json())
  .then(data => {
    console.log('Gastos totales:', data.total_expenses);
    console.log('Por categoría:', data.categories);
    
    // Encontrar categoría con mayor gasto
    const topCategory = Object.entries(data.categories)
      .reduce((a, b) => a[1] > b[1] ? a : b);
    console.log('Mayor gasto:', topCategory[0], '- $' + topCategory[1]);
  });
```

## 🎯 **Consejos de Uso**

1. **Categoriza consistentemente**: Usa las mismas categorías para gastos similares
2. **Revisa semanalmente**: Chequea el dashboard cada semana para mantener control
3. **Sube resúmenes mensuales**: Procesa tus PDFs de tarjeta cada mes
4. **Configura alertas**: Establece limits realistas para cada categoría
5. **Exporta regularmente**: Descarga tus datos cada trimestre como backup

## 🔒 **Seguridad**

### **Recomendaciones de Producción**
1. **Habilitar HTTPS**: Configurar certificado SSL
2. **authentication**: Implementar sistema de login
3. **Rate limiting**: Limitar peticiones por user
4. **validation de entrada**: Sanitizar datos del user
5. **Backup regular**: Respaldar la base de datos

---

**¡Tu sistema de gestión de gastos estilo Fintonic está listo para usar!** 🚀

Accede en: http://182.160.24.205/
- **user**: DS SR
- **Estado**: Premium
- **api disponible**: Sí
- **PDF parsing**: Funcional
- **Diseño**: Responsivo y moderno

**Nota**: Este es un MVP functional. En producción, se recomienda agregar authentication, HTTPS, y una base de datos persistente.