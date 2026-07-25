-- Domain tables
CREATE TABLE IF NOT EXISTS vehicles (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name          TEXT NOT NULL,
    make          TEXT NOT NULL,
    model         TEXT NOT NULL,
    year          INTEGER NOT NULL CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM now()) + 1),
    license_plate TEXT NOT NULL,
    owner_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vehicle_users (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_id TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('owner', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, vehicle_id)
);

CREATE TABLE IF NOT EXISTS mileage_entries (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    vehicle_id  TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mileage     INTEGER NOT NULL CHECK (mileage >= 0),
    date        TIMESTAMPTZ NOT NULL DEFAULT now(),
    notes       TEXT,
    fuel_litres REAL,
    fuel_amount REAL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- PUC: one record per vehicle (enforced by PK)
CREATE TABLE IF NOT EXISTS puc (
    vehicle_id         TEXT PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
    expiry_date        DATE NOT NULL,
    certificate_number TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Insurance: one record per vehicle (enforced by PK)
CREATE TABLE IF NOT EXISTS insurance (
    vehicle_id    TEXT PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
    expiry_date   DATE NOT NULL,
    policy_number TEXT,
    provider      TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_vehicles_owner ON vehicles(owner_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_users_user ON vehicle_users(user_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_users_vehicle ON vehicle_users(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_mileage_vehicle ON mileage_entries(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_mileage_user ON mileage_entries(user_id);
