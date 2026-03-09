<div align="center">

# 📡 Sales Radar

**Sales analytics backend for multi-source marketplace data aggregation**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docker.com)

> 🚧 **Work in Progress** — foundation layer complete, core features under development

</div>

---

## 💡 What It Will Do

A backend service that **aggregates sales data from multiple marketplaces** into a single analytics platform:

- 📊 **Unified dashboard** — sales metrics from Ozon, Wildberries, and other sources
- 🔄 **Automated sync** — scheduled data pulls via marketplace APIs
- 📈 **Trend analysis** — identify patterns across channels
- 🔐 **Multi-user** — token-based auth with audit logging

---

## ✅ What's Done

| Component | Status |
|-----------|--------|
| Project structure (Go clean arch) | ✅ |
| PostgreSQL + Goose migrations | ✅ |
| Config loader with validation | ✅ |
| Config & validation tests | ✅ |
| Docker Compose (app + DB) | ✅ |
| User & token tables | ✅ |
| Dataset & audit log tables | ✅ |
| Makefile (run, build, migrate) | ✅ |

---

## 🛠 Tech Stack

| | Technology | Purpose |
|-|------------|---------|
| 🔧 | **Go** | Backend service |
| 🗄 | **PostgreSQL** | Data storage |
| 🔄 | **Goose** | SQL migrations |
| 📦 | **Docker Compose** | Deployment |
| ✅ | **Testing** | Config & validation coverage |

---

## 🚀 Quick Start

```bash
git clone https://github.com/AlexMayka/sales-radar.git
cd sales-radar
docker-compose up -d    # Start PostgreSQL
make migrate-up         # Run migrations
make run                # Start the service
```

---

## 📁 Structure

```
cmd/                → Entry point
internal/
├── config/         → Config loader + validation + tests
└── utils/          → Env helpers + validation utils
migrations/         → Goose SQL migrations (users, tokens, datasets, audit)
```

## License

MIT
