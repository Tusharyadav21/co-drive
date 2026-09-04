# Product Requirements Document (PRD) (`docs/PRD.md`)

> **Product Name**: Co-Drive  
> **Product Type**: Vehicle Fleet & Regulatory Compliance Management System  
> **Target Release**: Production Ready

---

## 1. Executive Summary & Vision

**Co-Drive** is a lightweight, zero-daemon, production-grade vehicle fleet management platform. It empowers individual vehicle owners and commercial fleet managers to effortlessly track vehicle inventories, monitor fuel efficiency, log odometer milestones, and track regulatory compliance documents (PUC & Insurance) with automated proactive expiration warnings and cloud backups.

---

## 2. Target User Personas

1. **Individual Vehicle Owner**: Needs a secure, frictionless tool to track family car documents (PUC/Insurance expiry dates), log fuel refills, and monitor calculated fuel economy (km/L).
2. **Small Fleet Manager**: Manages multiple commercial vehicles, needing rapid search, centralized document compliance status, and tamper-resistant audit logs.
3. **System Administrator / Self-Hoster**: Deploys the service on lightweight hardware or cloud platforms (VPS, Docker, Cloud Run) connecting to hosted PostgreSQL (Neon, Supabase) or local PostgreSQL instances.

---

## 3. Core Functional Requirements

### 3.1 Authentication & Driver Profile Management
- **Passwordless Authentication**: 6-digit Email OTP via SMTP with cryptographic verification and rate limiting. Eliminates passwords, credential stuffing, and legacy password reset flows.
- **Default & Derived Display Name**: First-time OTP registration automatically computes a formatted display name from the email address, defaulting to `"Driver"`.
- **Driver Profile & Details**: Dedicated driver profile page (`/profile`) allowing drivers to add and edit contact numbers, bio, residential address, emergency contact, and avatar image, stored in the relational `user_details` table (`GET/PUT /api/users/me`).
- **Rolling Session Security**: Server-side PostgreSQL session management with `HttpOnly`, `SameSite=Lax`, `Max-Age=5184000` (60 days) cookies, automatic 30-day activity refresh for indefinite login persistence, Double-Submit CSRF cookie protection, and instant revocation (`/api/auth/logout`).

### 3.2 Vehicle Management
- **Vehicle Inventory**: CRUD operations for vehicles (Name, Make, Model, Year, License Plate).
- **Ownership Isolation**: Strict user-level tenant boundary ensuring users only access their own vehicles.

### 3.3 Mileage & Fuel Tracking
- **Odometer Milestones**: Log mileage entries with timestamps and optional notes.
- **Fuel Ingestion**: Log fuel entries (Odometer, Litres, Cost).
- **Dynamic Efficiency**: Automated calculation of fuel economy ($km/L$) between successive chronological fuel entries.

### 3.4 Regulatory Document Compliance
- **Document Tracking**: Register `PUC` and `INSURANCE` documents with policy numbers, providers, and expiration dates.
- **Expiry Classification**:
  - `EXPIRED` (Date $\le$ Today)
  - `EXPIRING_SOON` (Within next 30 days)
  - `VALID` (> 30 days remaining)

### 3.5 Managed Cloud Storage & Provider Independence
- **PostgreSQL Connectivity**: Standard `DATABASE_URL` supporting Neon (serverless branching & point-in-time recovery), Supabase, AWS RDS, or local PostgreSQL daemons.
- **Connection Pooling**: Built-in production pooling via `pgx/v5` (`MaxOpenConns: 25`, `MaxIdleConns: 5`).

---

## 4. Non-Functional Requirements (NFRs)

| Requirement | Metric / Specification |
| :--- | :--- |
| **Portability** | `CGO_ENABLED=0` pure-Go compilation with zero system C dependencies. |
| **Latency** | $P_{99}$ response time $< 25\text{ms}$ under standard operations. |
| **Data Integrity** | Strict ACID guarantees and foreign key cascades via PostgreSQL. |
| **Security** | OWASP Top 10 compliance: Bcrypt hashing, anti-IDOR gates, CSRF protection, secure cookie flags. |
| **Resource Footprint** | Memory footprint $< 40\text{MB}$ RSS under normal operation. |
