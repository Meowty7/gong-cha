# GongchaCUP 2026 — Sistema de Gestión de Bebidas

Aplicación web para gestionar productos, recetas encadenadas, inventario y planificación de eventos del reto II GongchaCUP 2026.

El reglamento y los datos oficiales son material académico del reto; no representan recetas comerciales reales de Gong cha.

## Estado del proyecto

El entorno y la estructura del monorepo están preparados. La lógica de la aplicación y el esquema de base de datos se implementarán durante la competencia.

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
├── docker-compose.yml
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

Cuando exista el servidor en `backend/cmd/server`, ejecútalo con:

```bash
cd backend
go run ./cmd/server
```

La API se expondrá en `http://localhost:8080`.

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
