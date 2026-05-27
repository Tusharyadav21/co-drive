# Project Architecture

## High-Level Structure

```
app/
  api/                        ← Serverless API routes
  vehicles/                   ← Vehicle-specific pages
    [id]/
      page.tsx                ← Server Component (data fetching)
    layout.tsx                ← Layout for Vehicles routes
  layout.tsx                  ← Layout for app folder and routing logic (if logged in or other wise redirect to login page)
  page.tsx                    ← Dashboard
  loading.tsx                 ← Route-level skeleton
  error.tsx                   ← Error boundary ("use client")
components/
  ui/                         ← shadcn/ui components
  vehicles/                   ← Shared components
lib/
  prisma.ts                   ← Singleton Prisma client
  db/                         ← Database query functions
  actions/                    ← Server actions ("use server")
  utils/                      ← Pure utilities
prisma/                       ← Schema & migrations
types/                        ← TypeScript & Zod definitions
```

## Data Model

- `User`: Application users.
- `Vehicle`: Vehicle details (make, model, plate).
- `VehicleUser`: Many-to-many join (User <-> Vehicle) with roles (`owner`, `viewer`).
- `MileageEntry`: Tracking vehicle mileage over time.
- `PUC`: Pollution Under Control certificate tracking.
- `Insurance`: Vehicle insurance policies.
- `Service`: Maintain Service Info.

## Conventions

- **Component File**: Max 150 lines. Extract sub-components. PascalCase naming (e.g., `VehicleCard.tsx`).
- **Database Queries**: All raw queries and Prisma calls must be in `lib/db/`. Do not put Prisma calls in UI components or pages.
- **Styling**: Tailwind CSS via utility classes. Use `cn()` (clsx + tailwind-merge) for conditional classes. Use `lucide-react` for icons.
- **Error Handling**: Use `error.tsx` (`"use client"`) for unexpected boundaries. Use `notFound()` for missing records.
