# AI Collaboration System

This is the central entry point for all AI agents working on the Vehicle Tracker application.
To avoid context bloat, our engineering standards and documentation are modularized.

## Available Context

Read these files based on the task you are performing:

- **[.ai/agent-rules.md](file:///Users/suven/Desktop/repo/vehicle-tracker-app/.ai/agent-rules.md)**: Universal behavioral rules for all AI agents (Read this first).
- **[.ai/nextjs-standards.md](file:///Users/suven/Desktop/repo/vehicle-tracker-app/.ai/nextjs-standards.md)**: Next.js v16.2 App Router best practices, conventions, and requirements.
- **[.ai/architecture.md](file:///Users/suven/Desktop/repo/vehicle-tracker-app/.ai/architecture.md)**: Project folder structure, data models, and database interactions.
- **[.ai/features/notifications.md](file:///Users/suven/Desktop/repo/vehicle-tracker-app/.ai/features/notifications.md)**: Setup and workflows for Web Push Notifications.
- **[.ai/features/betterauth.md](file:///Users/suven/Desktop/repo/vehicle-tracker-app/.ai/features/betterauth.md)**: BetterAuth Best Practices for authentication.

## Core Directives

1. **Always refer to `.ai/nextjs-standards.md`** before creating a new component or route.
2. **Always refer to `.ai/architecture.md`** before writing database queries or changing the schema.
3. Keep the codebase strictly typed.
4. If you add a new system, document it in a new file inside `.ai/features/` and link it here.
