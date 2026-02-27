# PLM (Product Lifecycle Management)

The PLM manages **products** (with SKU, revision, status), **BOMs** (bill of materials), and **ECOs** (engineering change orders). MES and Traceability can consume product and BOM data for production and genealogy.

---

## Overview

- **Products** — SKU, name, revision (e.g. A, 1.0), status (draft, released, obsolete).
- **BOM (Bill of Materials)** — For each product, list of component lines: child SKU, optional child revision, quantity. Used by MES for “what to build” and by Traceability for component-to-assembly linkage.
- **ECOs** — Engineering change orders: title, description, optional product link, status (draft, approved, implemented).

---

## Run PLM

```bash
./run-plm.sh
```

**Web UI**: http://localhost:8086 or http://localhost:8086/ui

Env: `PLM_LISTEN` (default :8086), `PLM_CONFIG`.

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /products | List. Query: ?status=, ?sku= |
| POST | /products | Create. Body: sku, name, revision?, status? |
| GET | /products/:id | Get product |
| PUT | /products/:id | Update product |
| GET | /products/:id/bom | Get BOM for product |
| POST | /products/:id/bom | Add BOM line. Body: child_sku, child_revision?, quantity, line_number? |
| DELETE | /products/:id/bom/:line_id | Remove BOM line |
| GET | /ecos | List ECOs. Query: ?status=, ?product_id= |
| POST | /ecos | Create ECO. Body: title, description?, product_id? |
| GET | /ecos/:id | Get ECO |
| PUT | /ecos/:id | Update ECO |

---

## Integration

- **MES** — Can call GET /products and GET /products/:id/bom to resolve SKU/revision and BOM for production orders.
- **Traceability** — Can link serial/lot to product revision and BOM for component-to-assembly genealogy.

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for the full system map.
