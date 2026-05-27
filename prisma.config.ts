import 'dotenv/config';
import { defineConfig } from '@prisma/config';

const url = process.env.POSTGRES_PRISMA_URL || process.env.DATABASE_URL;

if (!url) {
  throw new Error(
    'Database connection string is not set. Define POSTGRES_PRISMA_URL or DATABASE_URL in your environment (.env).'
  );
}

export default defineConfig({
  datasource: { url },
});
