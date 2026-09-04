# Documento técnico — II GongchaCUP 2026

**Equipo:** Chatarry  
**Aplicación:** Sistema de gestión de bebidas  
**Fecha:** 4 de septiembre de 2026

Este documento cubre arquitectura, modelo de datos, decisiones, algoritmo, pruebas, limitaciones y declaración de IA, conforme al §10 del reglamento.

## 1. Arquitectura

Monorepo con dos procesos y una base persistente:

- **Frontend:** Astro 7 + Vue 3 (Composition API) + Tailwind 4. Páginas estáticas que hidratan islas Vue. Cliente HTTP tipado (`frontend/src/lib/api`). Copy en español centralizado en `frontend/src/lib/i18n/es.ts`.
- **Backend:** Go 1.27 + Chi + pgx. El servidor aplica migraciones goose al arrancar. Capas: `api` (HTTP) → `recipe` / `inventory` (reglas) → `store` (SQL) → `domain`.
- **PostgreSQL 18.6** en Docker. En el compose local no se publican 5432 ni 8080: el navegador entra por el puerto 80 (o `WEB_PORT`) y nginx proxea `/api` y `/health` a la API. `PUBLIC_API_URL` vacío = mismo origen; Compose no toma esa variable del `.env` del host para no grabar `localhost:8080` en la imagen.

Flujo de una petición de cálculo: el handler carga recetas e inventario actuales, construye un BOM en memoria y no usa resultados fijos de los casos oficiales.


## 2. Modelo de datos

Tablas principales (`backend/internal/database/migrations/0001_init_schema.sql`):

| Tabla | Rol |
|---|---|
| `products` | Catálogo: id, nombre, tipo (`raw_material` / `semi_finished` / `finished_product`), unidad (`g` / `ml` / `unit`), descripción, `image_ref` |
| `recipes` | Producto resultante, rendimiento de lote, unidad |
| `recipe_components` | Líneas de receta (cantidad > 0) |
| `inventory_balances` | Existencia por producto (cantidad ≥ 0) |
| `inventory_movements` | Libro de descuentos confirmados |
| `events` / `event_demands` | Demanda multi-bebida (EVT001) |
| `calculation_runs` | Instantánea de cada cálculo (directo, inverso, evento, simulación, confirmación) |
| `idempotency_keys` | Evita doble descuento en confirmación |

El inventario no vive en `products`. Un producto del catálogo sin fila de stock aparece en la UI con existencia 0 y se crea al ajustar. `PUT /api/v1/inventory/{id}` exige que la unidad coincida con el catálogo. El seed oficial carga 60 productos, 30 recetas, 113 componentes, 30 balances de materia prima y 7 líneas de EVT001 desde `xlsx_export/*.csv`. Unidades: se conserva `g`/`ml`; la etiqueta española `unidad` se mapea a `unit` sin factor de conversión.

## 3. Decisiones

- **Cantidades como `NUMERIC` / `decimal.Decimal` / string JSON.** Evita `float64`.
- **BOM puro.** El motor no toca la base; cada request recarga recetas e inventario (el jurado puede cambiar datos).
- **Floor en cálculo directo.** Solo unidades completas (§5.3).
- **Simular vs confirmar.** Simular calcula consumo y leftover sin `UPDATE`. Confirmar bloquea filas (`FOR UPDATE`), revalida, descuenta, escribe movimientos y exige `idempotency_key`.
- **Evento.** El API consolida materias primas compartidas y calcula faltantes con `decimal` (`need`, `have`, `shortage ≥ 0`).
- **Imagen.** El reglamento pide imagen o referencia; se guarda `image_ref`.
- **Sin autenticación.** El reto no define usuarios. CORS por allowlist. SQL parametrizado.
- **Código en inglés, UI en español.**

## 4. Algoritmo de cálculo

Sea `factor = cantidad_pedida / rendimiento_lote`.

**Inverso (`Expand`):** escala los componentes inmediatos y, si un componente es semiterminado con receta, lo expande a hojas. Agrega por `(product_id, unit)`. Un semiterminado o terminado **sin receta** se marca en `incomplete` (A9) y no se finge como materia prima anónima.

**Directo (`MaxProduction`):** expande 1 unidad, calcula `floor(disponible / necesidad)` por insumo, toma el mínimo (empate: id lexicográfico), consume `max × necesidad` y leftovers `inventario − consume` (nunca negativos; también cuando `max = 0`). Si hay `incomplete` y no está en `use_direct`, responde error de validación.

**Evento (`ConsolidateEvent`):** expande cada línea, **suma** los raw compartidos una sola vez y compara contra el inventario actual (A6).

**Ciclos:** DFS de tres colores sobre aristas a semiterminados; se rechazan al crear/editar receta (CP06).

## 5. Pruebas

Backend (`go test ./...`):

- CP01 directo PT003 = 10, limitante MP010, leftover taro 0
- CP02 cadena PT002, limitante
- CP03 inverso PT010 × 25
- CP04 evento EVT001 consolidado
- CP05 leftover ST008 = 40
- CP06 ciclo rechazado
- CP07 simular no escribe
- CP08 cantidad inválida
- A9: semiterminado sin receta → `incomplete`; receta vacía → `incomplete`
- A5: inventario 0 → capacidad 0 y leftovers del resto
- Diamante: el mismo semiterminado por dos caminos suma ambas cantidades
- Evento: faltantes en el API; historial `GET /api/v1/calculations` con fecha
- Inventario: unidad distinta a la del catálogo → 400

Frontend: `bun src/lib/recipes/check.ts` (árbol, ciclos, badge “Sin receta”) y `bun src/lib/inventory/check.ts` (alta de existencia 0).

Los números oficiales aparecen solo como oráculos de test, no en el motor.

## 6. Limitaciones

Lo que el reto no pide y no está construido:

- **Sin autenticación.** No hay usuarios, JWT, sesiones ni roles. Quien alcance la URL puede leer y escribir catálogo, recetas e inventario. La mitigación es CORS por allowlist y SQL parametrizado; no hay control de acceso.
- **Sin subida de archivos.** El reglamento admite “imagen o referencia de imagen”. Se guarda `image_ref` (ruta o nombre); no hay multipart ni almacenamiento de binarios.
- **Sin rate limiting ni cuotas.** El API no limita peticiones. La demo está detrás de Caddy/HTTPS.
- **Sin TLS a Postgres en la demo.** `APP_ENV=development` porque el compose local usa `sslmode=disable`. En `production` esa URL se rechaza.
- **Historial de cálculos solo consulta.** `calculation_runs` es append-only; no hay borrado, exportación ni filtro por usuario (no hay usuarios).
- **Unidades oficiales.** Solo `g` / `ml` / `unit`. No hay conversiones (`kg` ↔ `g`).

Estas no son deudas del motor: capacidad, leftovers, faltantes, unidades de catálogo y fecha de cálculo están implementados.

## 7. Declaración de herramientas de IA

Conforme al §6 del reglamento:

| Herramienta | Uso |
|---|---|
| Cursor (agentes de código) | Implementación asistida de frontend Vue/Astro, API Go, tests, Docker/Compose, despliegue y este documento |
| Skills del repo (`vue-best-practices`, estándares de código) | Guía de Composition API y diffs mínimos |
| Documentación pública | Vue 3, Astro, Chi, pgx, Caddy, k3s/Docker |

Lo que el equipo entiende y mantiene:

- El BOM (factor, floor, consolidación, ciclos, `incomplete`) está en `backend/internal/recipe` y se puede explicar en pizarra.
- Confirmación: transacción, bloqueo, idempotencia.
- La UI es un cliente del API; capacidad, insumos y faltantes salen del backend.
- No se reutilizó una app previa. No hubo desarrollo de personas ajenas al equipo.

## 8. Cómo demostrar

1. `docker compose up --build` → http://localhost (solo Docker; `/api` en el mismo origen).
2. Catálogo: crear / editar / eliminar un producto que no esté en uso.
3. Inventario: ajustar un producto sin registro (queda en 0) y uno oficial.
4. Recetas: árbol encadenado, rechazo de ciclo, crear / editar / eliminar.
5. Directo PT003 / inverso PT010×25 / evento EVT001 (faltantes en el API).
6. Simular 3 unidades (el stock no cambia) y confirmar con `idempotency_key`.
7. Cambiar un stock o una receta y repetir el cálculo (A8).
