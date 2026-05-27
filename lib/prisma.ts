import { PrismaClient } from '@prisma/client';
import { PrismaPg } from '@prisma/adapter-pg';
import { Pool } from 'pg';

const connectionString = process.env.POSTGRES_PRISMA_URL || process.env.DATABASE_URL;

if (!connectionString) {
  throw new Error(
    'Database connection string is not set. Define POSTGRES_PRISMA_URL or DATABASE_URL in your environment.'
  );
}

// Cache the pool and client on the global object so that hot-reloads in
// development and warm serverless invocations in production reuse a single
// connection pool instead of exhausting database connections.
const globalForPrisma = globalThis as unknown as {
  prisma?: PrismaClient;
  pool?: Pool;
};

const pool = globalForPrisma.pool ?? new Pool({ connectionString });
const adapter = new PrismaPg(pool);

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    adapter,
    log: process.env.NODE_ENV === 'production' ? ['error'] : ['error', 'warn'],
    errorFormat: process.env.NODE_ENV === 'production' ? 'minimal' : 'pretty',
  });

globalForPrisma.prisma = prisma;
globalForPrisma.pool = pool;
