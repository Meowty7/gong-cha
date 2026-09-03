# GongchaCUP 2026 — Sistema de Gestión de Bebidas

Aplicación web para gestionar productos, recetas encadenadas, inventario y planificación de eventos del reto II GongchaCUP 2026.

El reglamento y los datos oficiales son material académico del reto; no representan recetas comerciales reales de Gong cha.

## Estado del proyecto

El backend Go (API, migraciones, siembra, motor de cálculo BOM, simulación/confirmación transaccional y endpoints versionados) está implementado y probado. El esquema de base de datos y los datos oficiales se cargan idempotentemente. El frontend permanece pendiente de implementación.

## Requisitos

- [Go 1.27.1](https://go.dev/)
- [Bun 1.4.0](https://bun.com/)
- Docker Engine con Docker Compose
- Git

Versiones verificadas para este proyecto:

| Herramienta | Versión |
|---|---|
| Go | 1.27.1 |
| Bun | 1.4.0 |
| Astro | 7.3.1 |
| Vue | 3.5.42 |
| Tailwind CSS | 4.3.3 |
| PostgreSQL | 18.6 |
| Chi | 5.3.2 |
| pgx | 5.10.0 |

## Estructura

```text
.
├── backend/          # API Go + Chi + pgx
├── frontend/         # Astro + Vue + Tailwind (Bun)
├── xlsx_export/      # Datos oficiales convertidos a CSV UTF-8
├── business/         # Investigación del sitio y modelo de negocio
├── compose.yml
├── Reglamento_GongchaCUP_2026.md
└── gong-cha-tarea-analisis.md
```

## Instalación

Clona el repositorio y entra al directorio:

```bash
git clone https://github.com/Meowty7/gong-cha.git
cd gong-cha
```

Instala las dependencias del frontend:

```bash
cd frontend
bun install
cd ..
```

Las dependencias Go se descargarán al añadir el código del backend:

```bash
cd backend
go mod download
cd ..
```

## Configuración

El archivo `.env` contiene las variables locales. Si no existe, créalo desde este ejemplo:

```env
POSTGRES_USER=gongcha
POSTGRES_PASSWORD=gongcha
POSTGRES_DB=gongcha_cup
POSTGRES_PORT=5432

DATABASE_URL=postgres://gongcha:gongcha@localhost:5432/gongcha_cup?sslmode=disable
SERVER_PORT=8080
CORS_ORIGIN=http://localhost:4321

PUBLIC_API_URL=http://localhost:8080
```

> `.env` está excluido del control de versiones. No publiques credenciales reales.

## Ejecución local

### Base de datos

Inicia PostgreSQL:

```bash
docker compose up -d
```

Revisa el estado:

```bash
docker compose ps
```

Detén los servicios:

```bash
docker compose down
```

### Frontend

```bash
cd frontend
bun run dev
```

Astro estará disponible normalmente en `http://localhost:4321`.

### Backend

El servidor aplica las migraciones de `backend/internal/database/migrations/` automáticamente al arrancar. Para ejecutarlo:

```bash
cd backend
go run ./cmd/server
```

La API se expondrá en `http://localhost:8080`.

#### Migraciones

Las migraciones son archivos SQL versionados bajo `backend/internal/database/migrations/`, aplicados con `goose` al iniciar el servidor. Para aplicarlas manualmente contra una base existente:

```bash
cd backend
go run ./cmd/server   # aplica migrations al arrancar
```

No se monta `docker-entrypoint-initdb.d`; el esquema lo gestiona la propia aplicación para evitar conflictos con `goose`.

#### Siembra de datos oficiales

El importador idempotente carga los CSV oficiales (`xlsx_export/`) en la base. Es seguro ejecutarlo varias veces (`ON CONFLICT DO NOTHING`):

```bash
cd backend
go run ./cmd/seed
```

#### Reinicio completo de la base

Para recrear el esquema y los datos desde cero:

```bash
docker compose down -v          # borra el volumen de Postgres
docker compose up -d            # recrea Postgres vacío
cd backend && go run ./cmd/seed # aplica migrations + datos oficiales
```

#### Chequeos de salud

- `GET /health/live` — el proceso responde.
- `GET /health/ready` — el pool de Postgres responde a `PING`.

```bash
curl http://localhost:8080/health/ready
```

#### Ejemplos de API

Todas las rutas de negocio están versionadas bajo `/api/v1`. Las cantidades se transmiten como cadenas para preservar precisión decimal.

```bash
# Capacidad directa (CP01): máximas unidades completas con inventario exclusivo
curl -s localhost:8080/api/v1/calculate/direct \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"PT003","inventory":{"MP010":"350","MP005":"2600","MP008":"600","MP024":"1000","MP001":"600","MP025":"5000"}}'

# Requerimiento inverso (CP03): expandir 25 unidades de PT010
curl -s localhost:8080/api/v1/calculate/inverse \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"PT010","quantity":"25"}'

# Planificación de evento (CP04): consolidar demanda
curl -s localhost:8080/api/v1/calculate/event \
  -H 'Content-Type: application/json' \
  -d '{"event_id":"EV001"}'

# Simulación (CP07): no descuenta
curl -s localhost:8080/api/v1/production/simulate \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"PT001","quantity":"3"}'

# Confirmación: descuenta dentro de una transacción (requiere idempotency_key)
curl -s localhost:8080/api/v1/production/confirm \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"PT001","quantity":"1","idempotency_key":"order-123"}'

# Historial de movimientos de un producto
curl -s localhost:8080/api/v1/inventory/MP025/history
```

#### Recuperación sin red (offline)

Si el entorno no puede descargar dependencias, fija antes de construir:

```bash
export GOSUMDB=off
export GOFLAGS=-mod=mod
```

Las pruebas de integración que necesitan PostgreSQL embebido se saltan automáticamente si no se puede provisionar la base; para forzarlas en CI, define `TEST_DATABASE_URL` con una conexión Postgres real.

#### Pruebas

```bash
cd backend
go test ./...          # unitarias + integración (skip si no hay Postgres)
go test -race ./...    # detector de data races
gofmt -l .             # formato (vacío = limpio)
go vet ./...           # análisis estático
```

## Datos de demostración

Los CSV UTF-8 derivados del archivo oficial se encuentran en `xlsx_export/`:

- `Catalogo_Productos.csv`
- `Recetas.csv`
- `Receta_Detalle.csv`
- `Inventario_Inicial.csv`
- `Demanda_Evento.csv`
- `Casos_Validacion.csv`
- `LEEME.csv`

Conservan los acentos en UTF-8. Abrirlos como Latin-1 produce texto corrupto, por ejemplo `ProducciÃ³n`.

## Pruebas funcionales esperadas

El reto incluye ocho casos de validación en `xlsx_export/Casos_Validacion.csv`, incluidos:

- producción directa y cálculo de insumo crítico;
- expansión de recetas encadenadas;
- cálculo inverso;
- consolidación de demanda de evento;
- inventario insuficiente;
- detección de ciclos;
- simulación frente a confirmación;
- validación de cantidades inválidas.

La implementación debe resolverlos dinámicamente y no codificar valores específicos de los casos.

## Documentación

- [`Reglamento_GongchaCUP_2026.md`](./Reglamento_GongchaCUP_2026.md): reglamento oficial.
- [`gong-cha-tarea-analisis.md`](./gong-cha-tarea-analisis.md): análisis funcional y técnico del reto.
- [`gong-cha-tech-stack.md`](./gong-cha-tech-stack.md): justificación del stack.

## Herramientas de IA

El equipo debe registrar en el documento técnico final las herramientas de IA utilizadas y demostrar comprensión de las partes asistidas, conforme al reglamento.
