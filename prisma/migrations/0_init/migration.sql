-- CreateTable User
CREATE TABLE "User" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "username" TEXT NOT NULL UNIQUE,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL
);

-- CreateTable Vehicle
CREATE TABLE "Vehicle" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "make" TEXT NOT NULL,
    "model" TEXT NOT NULL,
    "year" INTEGER NOT NULL,
    "licensePlate" TEXT NOT NULL,
    "ownerId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Vehicle_ownerId_fkey" FOREIGN KEY ("ownerId") REFERENCES "User" ("id") ON DELETE CASCADE
);

-- CreateIndex
CREATE INDEX "Vehicle_ownerId_idx" ON "Vehicle"("ownerId");

-- CreateTable VehicleUser
CREATE TABLE "VehicleUser" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "userId" TEXT NOT NULL,
    "vehicleId" TEXT NOT NULL,
    "role" TEXT NOT NULL DEFAULT 'viewer',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "VehicleUser_userId_fkey" FOREIGN KEY ("userId") REFERENCES "User" ("id") ON DELETE CASCADE,
    CONSTRAINT "VehicleUser_vehicleId_fkey" FOREIGN KEY ("vehicleId") REFERENCES "Vehicle" ("id") ON DELETE CASCADE
);

-- CreateUnique
CREATE UNIQUE INDEX "VehicleUser_userId_vehicleId_key" ON "VehicleUser"("userId", "vehicleId");

-- CreateIndex
CREATE INDEX "VehicleUser_vehicleId_idx" ON "VehicleUser"("vehicleId");

-- CreateTable MileageEntry
CREATE TABLE "MileageEntry" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "vehicleId" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "mileage" INTEGER NOT NULL,
    "date" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "notes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "MileageEntry_vehicleId_fkey" FOREIGN KEY ("vehicleId") REFERENCES "Vehicle" ("id") ON DELETE CASCADE,
    CONSTRAINT "MileageEntry_userId_fkey" FOREIGN KEY ("userId") REFERENCES "User" ("id") ON DELETE CASCADE
);

-- CreateIndex
CREATE INDEX "MileageEntry_vehicleId_idx" ON "MileageEntry"("vehicleId");

-- CreateIndex
CREATE INDEX "MileageEntry_userId_idx" ON "MileageEntry"("userId");

-- CreateTable PUC
CREATE TABLE "PUC" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "vehicleId" TEXT NOT NULL,
    "expiryDate" TIMESTAMP(3) NOT NULL,
    "certificateNumber" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "PUC_vehicleId_fkey" FOREIGN KEY ("vehicleId") REFERENCES "Vehicle" ("id") ON DELETE CASCADE
);

-- CreateIndex
CREATE INDEX "PUC_vehicleId_idx" ON "PUC"("vehicleId");

-- CreateTable Insurance
CREATE TABLE "Insurance" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "vehicleId" TEXT NOT NULL,
    "expiryDate" TIMESTAMP(3) NOT NULL,
    "policyNumber" TEXT,
    "provider" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Insurance_vehicleId_fkey" FOREIGN KEY ("vehicleId") REFERENCES "Vehicle" ("id") ON DELETE CASCADE
);

-- CreateIndex
CREATE INDEX "Insurance_vehicleId_idx" ON "Insurance"("vehicleId");
