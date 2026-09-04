# GongchaCUP 2026 — Sistema de Gestión de Bebidas

Aplicación web para catálogo, recetas encadenadas, inventario, cálculo directo/inverso, eventos y producción (simular / confirmar).

Los datos oficiales son material académico del reto; no son recetas comerciales de Gong cha.

## Arranque (solo Docker)

Requisito: [Docker](https://docs.docker.com/engine/install/) con el plugin Compose. No hace falta Go, Bun ni PostgreSQL en el host.

```bash
git clone https://github.com/Meowty7/gong-cha.git
cd gong-cha
docker compose up --build
```

Abre **http://localhost**

- UI y API en el mismo origen: el navegador llama a `/api` y `/health`; nginx lo reenvía. No se publica el puerto 8080 ni Postgres.
- Salud: `curl http://localhost/health/ready`

El seed oficial corre al levantar (`xlsx_export/`). Es idempotente. El API aplica las migraciones al arrancar.

Si el puerto 80 está ocupado:

```bash
WEB_PORT=8080 docker compose up --build
```

Luego usa http://localhost:8080

No pongas `PUBLIC_API_URL=http://localhost:8080` en un `.env`. Compose no pasa esa variable al build: la UI queda en mismo origen. Si la imagen se construyó con `:8080`, el navegador no carga datos.

### Reinicio de datos

```bash
docker compose down -v
docker compose up --build
```

### Variables opcionales

`cp .env.example .env` solo si quieres cambiar usuario/clave de Postgres o `WEB_PORT`. Sin `.env` valen los valores de ejemplo. Deja `PUBLIC_API_URL` vacío.

## Qué cubre la app

| Área | Ruta UI | En la pantalla | API |
|---|---|---|---|
| Catálogo | `/catalog` | crear, editar, eliminar | `/api/v1/products` |
| Inventario | `/inventory` | ajustar cantidad (no hay borrar fila) | `/api/v1/inventory` |
| Recetas | `/recipes` | crear, editar, eliminar | `/api/v1/recipes` |
| Directo / inverso | `/calculations` | capacidad y BOM | `/api/v1/calculate/direct`, `/inverse` |
| Eventos | `/events` | demanda consolidada | `/api/v1/calculate/event` |
| Producción | `/production` | simular / confirmar | `/api/v1/production/simulate`, `/confirm` |

Las cantidades viajan como cadenas decimales. El cálculo no hardcodea los casos CP01–CP08; usa el BOM y el inventario vivos.

## Pruebas (opcional, hace falta Go / Bun)

```bash
cd backend && go test ./...
cd frontend && bun src/lib/recipes/check.ts && bun src/lib/inventory/check.ts
```

Casos oficiales: `xlsx_export/Casos_Validacion.csv`.

## Documentación

- [`Reglamento_GongchaCUP_2026.md`](./Reglamento_GongchaCUP_2026.md)
- [`Documento_Tecnico_GongchaCUP_2026.pdf`](./Documento_Tecnico_GongchaCUP_2026.pdf)
- [`xlsx_export/LEEME.csv`](./xlsx_export/LEEME.csv)

## Estructura

```text
backend/          API Go (Chi, pgx, goose)
frontend/         Astro + Vue + Tailwind
xlsx_export/      CSV oficiales UTF-8
compose.yml       Stack completo: Postgres + seed + API + UI
```
