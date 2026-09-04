# Universal Agent Contract (`AGENTS.md`)

> **Repository**: Co-Drive (`co-drive`)  
> **Governance Level**: Strict Production Engineering Standards  
> **Target Audience**: All Autonomous AI Agents, Pair Programmers, and Engineering Contributors

---

## 1. Core Operating Directives

Every AI agent operating in this repository **MUST** adhere strictly to the following deterministic rules:

1. **Zero Hallucination & Token Optimization**: Never invent APIs, data models, or terminology. Consult [CONTEXT.md](file:///Users/tusharyadav/Dev/co-drive/CONTEXT.md) for official ubiquitous domain terms.
2. **CodeGraph-First Exploration**: When `.codegraph/` exists, **ALWAYS** use `codegraph_explore` (MCP) or `codegraph explore` (CLI) **BEFORE** using ripgrep or reading raw files. Trace call graphs and types deterministically.
3. **6-Phase Engineering Lifecycle**: All modifications must progress through the 6-phase workflow defined in [.agents/workflows/unified-agent-workflow.md](file:///Users/tusharyadav/Dev/co-drive/.agents/workflows/unified-agent-workflow.md).
4. **Architectural Governance**: All code must conform to the invariants in [docs/Architecture.md](file:///Users/tusharyadav/Dev/co-drive/docs/Architecture.md) and accepted ADRs in [docs/Decisions.md](file:///Users/tusharyadav/Dev/co-drive/docs/Decisions.md).
5. **Strict Verification Gate**: No task is complete without running and passing `make check` (Typecheck/Vet + Test Suite + Binary Build) with exit code 0.

---

## 2. Fast Navigation & Master Indexes

| Resource | File Location | Purpose |
| :--- | :--- | :--- |
| **Unified Workflow** | [`.agents/workflows/unified-agent-workflow.md`](file:///Users/tusharyadav/Dev/co-drive/.agents/workflows/unified-agent-workflow.md) | Mandatory 6-phase development lifecycle |
| **Documentation Hub** | [`docs/README.md`](file:///Users/tusharyadav/Dev/co-drive/docs/README.md) | Central repository sitemap and documentation portal |
| **Domain Language** | [`CONTEXT.md`](file:///Users/tusharyadav/Dev/co-drive/CONTEXT.md) | Ubiquitous vocabulary, entities, and domain formulas |
| **System Architecture** | [`docs/Architecture.md`](file:///Users/tusharyadav/Dev/co-drive/docs/Architecture.md) | Topologies, component boundaries, and security model |
| **ADR Register** | [`docs/Decisions.md`](file:///Users/tusharyadav/Dev/co-drive/docs/Decisions.md) | Architecture Decision Records & standard ADR template |
| **Product Scope / PRD** | [`docs/PRD.md`](file:///Users/tusharyadav/Dev/co-drive/docs/PRD.md) | Product scope, user personas, and feature specifications |
| **Operations Runbook** | [`docs/Runbook.md`](file:///Users/tusharyadav/Dev/co-drive/docs/Runbook.md) | Deployments, environment variables, and disaster recovery |
| **Interactive Archify Suite** | [`docs/archify.html`](file:///Users/tusharyadav/Dev/co-drive/docs/archify.html) | Interactive 5-mode Archify architecture, sequence, and lifecycle visualizer |

---

## 3. Specialized Governance Rules

All agents must load and enforce modular rule files located under [`.agents/rules/`](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/):

- **[Architecture Invariants](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/architecture.md)**: Pure-Go PostgreSQL rules, connection pool constraints, sub-resource ownership checks, standardized JSON envelope responses.
- **[Production Safety](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/production.md)**: Protected files (`.env`, `go.mod`, migration scripts), data loss prevention, schema migration safety.
- **[Testing Standards](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/testing.md)**: Test hierarchy, zero swallowed assertions, schema isolation, deterministic test runners.
- **[CodeGraph Protocol](file:///Users/tusharyadav/Dev/co-drive/.agents/rules/codegraph.md)**: Guidelines for symbol discovery and call graph navigation before reading files.

---

## 4. Verification Gate Commands

Before concluding any implementation or handing off tasks, execute:

```bash
# Primary Verification Gate (Vet + Tests + Multi-Binary Compilation)
make check

# Granular Checks
make test          # Executes full Go test suite across all internal/ and pkg/ packages
make build         # Builds bin/server and bin/migrate binaries
make migrate-up    # Verifies database migrations against PostgreSQL instance
```

---

## 5. Multi-Agent Handoff Protocol

When context windows truncate or a peer agent resumes work, the executing agent must produce a structured handoff note (in PR comments or `handoff.md`):

1. **Completed Tasks**: Exact features, tests, or bug fixes implemented.
2. **Files Modified**: Clickable relative links with concise explanation of changes.
3. **Verification Results**: Exact output/exit code of `make check`.
4. **Active State / Next Steps**: Explicit blockers or remaining items for the incoming agent.
