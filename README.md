# Co-Drive - Smart Shared Vehicle Tracker

Co-Drive is a Progressive Web Application (PWA) designed for families and shared fleet operators. It tracks vehicle odometer logs, calculates fuel efficiency, checks PUC and insurance expiry windows, supports browser push subscriptions, and uses Better Auth for Google OAuth and passkey-capable authentication.

---

## Key Features

- **Smart Fleet Dashboard**: High-level visual health, odometer status, and active expiry alerts for family cars and bikes.
- **Automatic Fuel Efficiency (Mileage)**: Calculates distance traveled and fuel economy in `km/l` for refill logs.
- **Document Expiry Checks**: Shows warning indicators and exposes a secured API for upcoming PUC and insurance renewals.
- **Passwordless Passkeys**: Supports device-native passkey flows through Better Auth and WebAuthn.
- **Web Push Subscription Support**: Registers and stores browser push subscriptions for notification delivery workflows.
- **Family Sharing Board**: Shares vehicle view and edit permissions with other members by email.

---

## Tech Stack & Architecture

- **Framework**: Next.js v16.2 (App Router, Turbopack, React 19)
- **Styling**: Tailwind CSS v4 (Harmony HSL customized colors, glassmorphism tokens, and micro-animations)
- **Database**: PostgreSQL (Neon.tech / Supabase)
- **ORM**: Prisma Client v7
- **Authentication**: Better Auth v1 (Google OAuth Provider + Passkeys Biometric Plugin)
- **Icons & Graphics**: Lucide React
- **Runtime Orchestrator**: Bun (recommended) / Node.js

---

## Environment Variables Setup

Before running the application in local development or production, copy the template `.env.example` file to `.env`:

```bash
cp .env.example .env
```

Here is a breakdown of each configuration key and where to obtain them:

| Key                                 | Description                             | How to Obtain / Configuration                                                                                                                                                                                                     |
| :---------------------------------- | :-------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`DATABASE_URL`**                  | Primary PostgreSQL connection string    | Create a PostgreSQL instance on **Neon.tech** or **Supabase.com** and copy the database URI connection string.                                                                                                                    |
| **`POSTGRES_PRISMA_URL`**           | Prisma PostgreSQL connection string     | Usually the same value as `DATABASE_URL`; Prisma falls back to `DATABASE_URL` when this is omitted.                                                                                                                               |
| **`BETTER_AUTH_SECRET`**            | Key to encrypt cookies and auth tokens  | Generate a random 32-character hex key by running `openssl rand -hex 32` or `bun x better-auth secret` in your terminal.                                                                                                          |
| **`BETTER_AUTH_URL`**               | The base backend URL of the auth server | For local development: `http://localhost:3000`. For production: `https://your-production-domain.com`.                                                                                                                             |
| **`NEXT_PUBLIC_APP_URL`**           | Client-side application domain          | For local development: `http://localhost:3000`. For production: `https://your-production-domain.com`.                                                                                                                             |
| **`GOOGLE_CLIENT_ID`**              | Google OAuth Client ID credentials      | Go to [Google Cloud Console](https://console.cloud.google.com/), create a project, go to **Credentials > Create Credentials > OAuth client ID (Web application)**. Configure redirect URIs: `[APP_URL]/api/auth/callback/google`. |
| **`GOOGLE_CLIENT_SECRET`**          | Google OAuth Client Secret key          | Obtained alongside `GOOGLE_CLIENT_ID` in your Google Cloud Console project.                                                                                                                                                       |
| **`NEXT_PUBLIC_VAPID_PUBLIC_KEY`**  | Public VAPID key for push subscriptions | Generate keys using Web-Push CLI by running `npx web-push generate-vapid-keys` in your terminal. This app currently stores subscriptions; server-side push delivery must be added separately.                                     |
| **`CRON_SECRET`**                   | Access key for expiry check cron        | Generate a secure random string (e.g. via `openssl rand -hex 16`). Used to trigger `/api/notifications/check-expiry` securely.                                                                                                    |
| **`NEXT_PUBLIC_DEV_LOGIN_ENABLED`** | Shows the mock developer login UI       | Keep `false` in production. Set to `true` only for local development when using the developer login bypass.                                                                                                                       |
| **`DEV_LOGIN_ENABLED`**             | Enables the mock developer login action | Keep `false` in production. Set to `true` only for local development when using the developer login bypass.                                                                                                                       |
| **`DEV_LOGIN_EMAIL`**               | Mock developer login email              | Local-development identity used by the developer login bypass.                                                                                                                                                                    |
| **`DEV_LOGIN_NAME`**                | Mock developer display name             | Local-development display name used by the developer login bypass.                                                                                                                                                                |

---

## Getting Started

### 1. Local Requirements

Ensure you have **Bun** (highly recommended) or **Node.js (v18+)** installed.

### 2. Install Dependencies

```bash
bun install
```

### 3. Database Migration & Setup

Configure your database string in `.env` and initialize the tables:

```bash
bun x prisma db push
```

### 4. Run Development Server

Start the Next.js development server using Turbopack compiler:

```bash
bun run dev
```

Open **[http://localhost:3000](http://localhost:3000)** in your browser.

---

## Production Deployment Workflow

### 1. Deploying the Database

1. Set up a PostgreSQL instance on **Supabase** or **Neon**.
2. Apply the committed Prisma migration history:
   ```bash
   bun x prisma migrate deploy
   ```

### 2. Deploying to Vercel / Render

Deploy your repository directly to Vercel:

1. Link your GitHub repository in Vercel.
2. In the project settings, add the required environment variables. Ensure both `NEXT_PUBLIC_DEV_LOGIN_ENABLED` and `DEV_LOGIN_ENABLED` are set to `false`.
3. Set the Build Command: `bun run build`.
4. Deploy!

### 3. Configuring Cron Jobs (for daily expiry checks)

Set up a daily automated cron job to check for documents approaching expiry. The current endpoint returns an expiry payload; it does not send Web Push messages by itself.

- **Endpoint to call**: `https://your-production-domain.com/api/notifications/check-expiry`
- **Method**: `GET`
- **Headers**: `Authorization: Bearer <YOUR_CRON_SECRET>`
- **Orchestration**: Configure Vercel Crons inside a `vercel.json` file, or use an external secure scheduler like **cron-job.org** to trigger the API.

---

## Code Quality & Testing Pipelines

The repository features high-quality configuration rules for production stability. Verify before pushing:

- **Linting & Code Quality**: `bun run lint` (ESLint v9 compliant)
- **Code Formatting**: `bun run format` (Prettier code styling)
- **Strict Typecheck**: `bun run typecheck` (tsc diagnostics)
- **Complete Build Check**: `bun run build`
