# Tech Stack & Programming Languages for QSR Franchise Model (Gong Cha)

## 1. Executive Summary

For a multi-unit beverage franchise business model (high volume, low unit cost, perishable inventory, multi-tier franchising, recipe Bill-of-Materials), no single language handles everything alone. A modern QSR inventory and reporting system typically splits into three layers:

```
┌──────────────────────────────────────────────────────────────┐
│                    Reporting & Analytics                     │
│               Python (Pandas/Polars) + SQL + BI              │
├──────────────────────────────────────────────────────────────┤
│               Inventory & Business Logic Core                │
│             TypeScript (NestJS) OR C# (.NET 8/9)             │
├──────────────────────────────────────────────────────────────┤
│              POS Ingestion & High-Throughput Hub             │
│                    Go (Golang) OR Node.js                    │
├──────────────────────────────────────────────────────────────┤
│                 Database & Event Storage                     │
│               PostgreSQL + Redis + ClickHouse                │
└──────────────────────────────────────────────────────────────┘
```

---

## 2. Best Languages by Functional Domain

### A. Calculations & Business Logic (COGS, Recipe BOM, Royalties)

**Primary Recommendation:** **Python** or **C# (.NET)**

#### Key Domain Calculations
1. **Recipe Deconstruction (Bill of Materials - BOM):**
   - 1 Large Brown Sugar Boba = 250ml fresh milk + 50g tapioca pearls + 30ml brown sugar syrup + 1 cup (24oz) + 1 PP lid/film + 1 giant straw.
   - Deduction must calculate yield rates, brewing wastage (tea brewed every 4 hours), and prep variance.
2. **Theoretical vs. Actual Variance (Food Cost Variance):**
   $$\text{Variance \%} = \frac{\text{Actual Usage} - \text{Theoretical Usage (from POS)}}{\text{Theoretical Usage}} \times 100$$
3. **Franchise Royalty & Marketing Fee Calculations:**
   - Multi-tier fee engine based on gross sales minus allowed discounts, localized tax handling, and currency conversion across 30+ countries.

#### Why Python?
- Unmatched libraries for numerical analysis: `pandas`, `polars`, `numpy`.
- Easy rule engines to model localized franchise agreements, supplier price changes, and promotions.

#### Why C# (.NET)?
- Strong decimal precision arithmetic (`decimal` type avoids floating-point errors).
- Mature enterprise patterns for complex financial ledgers and multi-tenant domain models.

---

### B. Automated Reporting & Export Engines

**Primary Recommendation:** **Python**

| Report Type | Target Output | Recommended Libraries / Tools |
|---|---|---|
| **Store Operational Sheets** (daily shift audits, waste logs) | PDF / Web | `WeasyPrint` (HTML/CSS to PDF), `ReportLab` |
| **Franchisee Financials & Audits** | Excel with formulas & formatting | `XlsxWriter`, `OpenPyXL` |
| **Corporate BI & Franchise Dashboards** | Interactive dashboards | `Metabase`, `Apache Superset`, or `Power BI` |
| **Data Aggregation & Rollups** | SQL views / materialized tables | `dbt` (data build tool) + `PostgreSQL` / `ClickHouse` |

---

### C. Real-Time Inventory & Multi-Store Management Engine

**Primary Recommendation:** **TypeScript (Node.js / NestJS)** or **C# (.NET)** or **Go**

#### Key Requirements
- **Multi-tier location hierarchy:** Global HQ $\rightarrow$ Master Franchisee $\rightarrow$ Regional Warehouse $\rightarrow$ Store $\rightarrow$ Kiosk.
- **Batch / Expiration Tracking:** Perishable items (cooked pearls: 4h shelf life; brewed tea: 4h; open milk: 48h).
- **Par Level & Automated Reordering:** Reorder point triggered by current stock + lead time + sales velocity.

#### Language Comparison

| Language | Strengths | Best Used For |
|---|---|---|
| **TypeScript (NestJS)** | Unified stack (shares types with frontend POS/kiosk/dashboard), rapid development, large ecosystem. | Core API, Inventory CRUD, Admin portals. |
| **C# (.NET 8/9)** | High performance, bulletproof transaction management (ACID), strong typing, built-in dependency injection. | Enterprise ERP, multi-tenant accounting systems. |
| **Go (Golang)** | Minimal memory footprint, high concurrency, single binary deployment. | Ingesting webhook streams from 2,000+ POS terminals simultaneously. |

---

## 3. Recommended Architectural Stacks

### Stack Option 1: Modern, Agile & Fast-to-Market (Recommended)
- **Backend API & Core Logic:** TypeScript (`NestJS` or `FastAPI` in Python)
- **Calculation & Report Workers:** Python (`Polars`, `Celery`/`Temporal`, `WeasyPrint`, `XlsxWriter`)
- **Primary Database:** PostgreSQL (with Row-Level Security for multi-tenant franchise isolation)
- **Caching & Locks:** Redis (inventory stock reservation during peak order rushes)
- **Frontend / Dashboard:** React / Next.js with Tailwind CSS

### Stack Option 2: Heavy Enterprise / Franchise Core
- **Core Backend:** C# (.NET 8/9 Web API) + Entity Framework Core
- **Database:** Microsoft SQL Server or PostgreSQL
- **Reporting:** SQL Server Reporting Services (SSRS) or custom .NET ClosedXML/QuestPDF
- **Best if:** You need deep ERP integration (SAP, Microsoft Dynamics 365, Oracle NetSuite).

---

## 4. Key Data Models for this Business Model

### Core Entities Needed:
1. `Tenants` (HQ, Master Franchisee, Store, Kiosk)
2. `Ingredients` (SKU, unit of measure, cost per unit, shelf life)
3. `Recipes` & `RecipeItems` (Drink ID $\rightarrow$ ingredient decomposition + yield factors)
4. `InventoryBalances` (Store ID, Ingredient ID, Batch ID, Expiry Date, Qty on Hand, Par Level)
5. `StockMovements` (PO receipts, transfers, POS sales deductions, waste logs)
6. `RoyaltyRules` (Tiered %, fixed rates, territory overrides)
7. `DailySalesRollups` (Store ID, Date, Gross, Net, COGS, Royalty Due)
