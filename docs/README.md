# Central Documentation Hub (`docs/README.md`)

Welcome to the **Co-Drive Central Documentation Hub**. This repository enforces architectural consistency, token optimization, and deterministic development standards across all human engineers and autonomous AI agents.

---

## 🗺️ Master Documentation Sitemap

| Category | Primary Document | Description | Target Readers |
| :--- | :--- | :--- | :--- |
| **System Architecture** | [`docs/Architecture.md`](file:///Users/tusharyadav/Dev/co-drive/docs/Architecture.md) | Topologies, database engine, component models, and data flows. | Architects, Backend Engineers, AI Agents |
| **Architecture Decisions (ADRs)** | [`docs/Decisions.md`](file:///Users/tusharyadav/Dev/co-drive/docs/Decisions.md) | ADR Register, architectural invariants, and standard ADR template. | All Contributors & AI Planners |
| **Domain Language & Rules** | [`CONTEXT.md`](file:///Users/tusharyadav/Dev/co-drive/CONTEXT.md) | Ubiquitous vocabulary, formulas, entity constraints, and anti-patterns. | Product Managers, Developers, Agents |
| **Product Scope & PRD** | [`docs/PRD.md`](file:///Users/tusharyadav/Dev/co-drive/docs/PRD.md) | Product scope, user personas, functional specifications, and roadmap. | Product Owners, QA, Developers |
| **Operations & Runbook** | [`docs/Runbook.md`](file:///Users/tusharyadav/Dev/co-drive/docs/Runbook.md) | Deployment guides, env vars reference, backups, and observability. | DevOps, SREs, Maintainers |
| **OpenAPI Specification** | [`docs/openapi.yaml`](file:///Users/tusharyadav/Dev/co-drive/docs/openapi.yaml) | OpenAPI 3.0 API schema definitions for all endpoints. | API Consumers, Client Developers |
| **Interactive Archify Suite** | [`docs/archify.html`](file:///Users/tusharyadav/Dev/co-drive/docs/archify.html) | Interactive 5-mode Archify architecture, sequence, and lifecycle visualizer. | All Engineers, Architects, Reviewers |

---

## 🧭 Multi-Agent Engineering Protocols

All development in this repository is governed by the following strict guidelines:

1. **Universal Contract**: Read [AGENTS.md](file:///Users/tusharyadav/Dev/co-drive/AGENTS.md) / [CLAUDE.md](file:///Users/tusharyadav/Dev/co-drive/CLAUDE.md) for core requirements and verification steps.
2. **6-Phase Workflow**: Follow [.agents/workflows/unified-agent-workflow.md](file:///Users/tusharyadav/Dev/co-drive/.agents/workflows/unified-agent-workflow.md) from discovery through multi-gate verification.
3. **Specialized Governance Rules**:
   - [Architecture Invariants (`.agents/rules/architecture.md`)](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/architecture.md)
   - [Production & Safety (`.agents/rules/production.md`)](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/production.md)
   - [Testing Standards (`.agents/rules/testing.md`)](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/testing.md)
   - [CodeGraph Protocol (`.agents/rules/codegraph.md`)](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/codegraph.md)

---

## ⚡ Quick Verification Gate

```bash
# Run full static analysis, test suite, and compilation
make check
```
