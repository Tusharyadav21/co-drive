# Ubiquitous Domain Language & Context (`CONTEXT.md`)

> **Single Source of Truth**: All agents, engineers, and documentation must use the exact terms defined herein. Strictly **NO** invented synonyms, alternative frameworks, or ungrounded data models.

---

## 1. System Identity

- **Application Name**: `Co-Drive` (hyphenated as `co-drive` in module paths and repository names).
- **Core Purpose**: Lightweight, production-grade Go vehicle fleet & document management backend, utilizing pure-Go PostgreSQL (Neon-ready, Supabase, or local) with connection pooling, passwordless Email OTP authentication, and extensible driver profile management.
- **Architecture Style**: Go Standard Clean Architecture (HTTP Routers/Handlers -> Domain Services -> Repository Layer -> PostgreSQL Database Driver).

---

## 2. Ubiquitous Vocabulary (Domain Terms)

| Canonical Term | Definition | Strictly Forbidden Synonyms / Anti-Patterns |
| :--- | :--- | :--- |
| **`User`** | An authenticated identity record owning vehicles, credentials, and session tokens. | *Account holder, Client, Member* |
| **`User Details`** | Extended driver profile information (phone, bio, address, emergency contact, avatar) stored in `user_details`. | *User metadata, Extra info* |
| **`Vehicle`** | A motor asset owned by a `User`, containing identification attributes (Make, Model, Year, License Plate). | *Car, Automobile, Machine, Transport* |
| **`Mileage Entry`** | An odometer reading recorded at a specific timestamp for a `Vehicle`, with optional fuel details. | *Distance record, Trip meter, Run log* |
| **`Vehicle Document`** | A regulatory compliance record associated with a `Vehicle` having a specific `DocumentType` (`puc` or `insurance`). | *Certificate, License paper, File* |
| **`PUC`** | Pollution Under Control certificate required for vehicular regulatory compliance. | *Smog check, Emissions test* |
| **`Insurance`** | Mandatory vehicle insurance policy record tracking policy number, provider, and expiration date. | *Coverage, Assurance* |
| **`OTP`** | 6-digit numeric One-Time Password sent via SMTP email for passwordless verification with short TTL. | *Pin code, Verification code, 2FA token* |
| **`Session`** | Server-side PostgreSQL session entity identified by a cryptographically random token in an HTTP-only cookie. | *JWT, Bearer token, Auth header* |
| **`CSRF Token`** | Double-submit CSRF cookie validated on state-changing HTTP requests (`POST`, `PUT`, `DELETE`). | *Anti-tamper token, X-CSRF* |
| **`Database URL`** | PostgreSQL connection string (`DATABASE_URL`) formatted as `postgres://user:pass@host:port/db?sslmode=mode`. | *DB path, SQLite file* |

---

## 3. Mathematical Formulas & Business Rules

### 3.1 Fuel Efficiency Calculation
Efficiency is computed on-the-fly dynamically across chronological pairs of mileage entries with fuel:
$$\text{Efficiency (km/L)} = \frac{\text{Current Odometer (km)} - \text{Previous Odometer (km)}}{\text{Fuel Added (Litres)}}$$

- **Rules**:
  - If previous fuel entry does not exist: `Efficiency = 0`.
  - If $\text{Fuel Added} \le 0$: `Efficiency = 0` (prevent divide-by-zero).
  - If $\text{Current Odometer} < \text{Previous Odometer}$ (rollback/tampering): `Efficiency = 0`.

### 3.2 Document Expiry Windows & Status Calculation
- **`EXPIRED`**: Expiration date $\le \text{Today}$.
- **`EXPIRING_SOON`**: Expiration date $> \text{Today}$ and $\le \text{Today} + 30\text{ days}$.
- **`VALID`**: Expiration date $> \text{Today} + 30\text{ days}$.

---

## 4. Entity Models & Field Constraints

### `User`
- `id` (UUID string, Primary Key)
- `email` (string, Unique, Valid Email Format)
- `full_name` (string, Defaults to 'Driver')
- `email_verified` (BOOLEAN, default FALSE)
- `created_at` / `updated_at` (TIMESTAMPTZ)

### `UserDetails`
- `user_id` (string, Primary Key, Foreign Key -> `users.id` ON DELETE CASCADE)
- `phone` (string, Optional)
- `bio` (string, Optional)
- `address` (string, Optional)
- `emergency_contact` (string, Optional)
- `avatar_url` (string, Optional)
- `updated_at` (TIMESTAMPTZ)

### `Vehicle`
- `id` (UUID string, Primary Key)
- `user_id` (string, Foreign Key -> `users.id` ON DELETE CASCADE)
- `name` (string, Required)
- `make` (string, Optional)
- `model` (string, Optional)
- `year` (int, Defaults to 2020 if $\le 0$)
- `license_plate` (string, Optional)
- `created_at` / `updated_at` (TIMESTAMPTZ)

### `VehicleDocument`
- `id` (UUID string, Primary Key)
- `vehicle_id` (string, Foreign Key -> `vehicles.id` ON DELETE CASCADE)
- `type` (TEXT CHECK `puc` OR `insurance`)
- `reference` (string, Required)
- `issuer` (string, Optional)
- `expiry_date` (TIMESTAMPTZ, Required)
- `updated_at` (TIMESTAMPTZ)

### `MileageEntry`
- `id` (UUID string, Primary Key)
- `vehicle_id` (string, Foreign Key -> `vehicles.id` ON DELETE CASCADE)
- `mileage` (int, Required, $\ge 0$)
- `date` (TIMESTAMPTZ, Required)
- `notes` (string, Optional)
- `fuel_litres` (float64, Optional, $> 0$)
- `fuel_amount` (float64, Optional, $\ge 0$)
- `created_at` (TIMESTAMPTZ)

---

## 5. Security & Isolation Invariants

1. **Sub-Resource Ownership Validation**:
   Any operation mutating or querying vehicle sub-resources (mileage entries, documents) **MUST** perform an ownership check (`ownsVehicle`) against `session.UserID`. If the vehicle does not belong to the user, respond with `404 Not Found` (never `403 Forbidden`) to prevent enumeration attacks.
2. **Session Cookies & Rolling Lifecycle**:
   Cookie name: `session_id`. Flags: `HttpOnly=true`, `SameSite=Lax`, `Path=/`, `Max-Age=5184000` (60 days), `Secure=true` (when serving over HTTPS / behind TLS proxies). Automatic rolling refresh extends the session to 60 days in PostgreSQL and re-issues the cookie whenever an active user interacts with $< 30$ days remaining (50% threshold).
3. **Double-Submit CSRF**:
   Cookie name: `csrf_token`. Request header required for unsafe methods: `X-CSRF-Token`.
