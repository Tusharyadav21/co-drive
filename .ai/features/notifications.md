# Web Push Notifications Setup

This document explains how to set up web push notifications for Vehicle Tracker.

## Overview

The app includes push notification infrastructure for:

- PUC expiry reminders (7 days before expiry)
- Insurance expiry reminders (15 days before expiry)

## Setup Steps

### 1. Generate VAPID Keys

Web Push requires VAPID (Voluntary Application Server Identification) keys. Generate them using:

```bash
npm install -g web-push
web-push generate-vapid-keys
```

This will output:

```
Public Key: <PUBLIC_KEY>
Private Key: <PRIVATE_KEY>
```

### 2. Add Environment Variables

Create/update your `.env.local` file with:

```
NEXT_PUBLIC_VAPID_PUBLIC_KEY=<PUBLIC_KEY>
VAPID_PRIVATE_KEY=<PRIVATE_KEY>
```

The `NEXT_PUBLIC_` prefix makes the public key available to the browser.

### 3. Database Setup (Optional)

Currently, push subscriptions are stored in memory. For production, extend the Prisma schema to store subscriptions:

```prisma
model PushSubscription {
  id        String     @id @default(cuid())
  userId    String
  endpoint  String     @unique
  auth      String
  p256dh    String
  user      User       @relation(fields: [userId], references: [id], onDelete: Cascade)
  createdAt DateTime   @default(now())
  @@index([userId])
}
```

### 4. Background Job for Reminders

To send automated reminders, set up a cron job that calls:

```
GET /api/notifications/check-expiry
```

Example cron services:

- **Vercel Cron Functions**: Use `/api/cron/check-expiry` endpoint
- **External services**: Easycron, Pingmesh, or your own infrastructure

Example Vercel cron function:

```typescript
// app/api/cron/check-expiry/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest) {
  if (req.headers.get('Authorization') !== `Bearer ${process.env.CRON_SECRET}`) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const res = await fetch('http://localhost:3000/api/notifications/check-expiry');
  const data = await res.json();

  return NextResponse.json(data);
}

export const runtime = 'nodejs';
```

### 5. Service Worker

The service worker is automatically registered when notifications are initialized. It's located at `/public/sw.js`.

## How It Works

1. **User Opts In**: When a user visits the app, they're prompted for notification permission
2. **Service Worker Registration**: The browser registers the service worker
3. **Push Subscription**: The app subscribes to push notifications
4. **Subscription Storage**: The subscription endpoint is saved (currently in memory)
5. **Background Check**: A cron job periodically checks for expiring documents
6. **Notification Delivery**: When expiry is detected, a push notification is sent to the user's device

## Testing Notifications

### Manually test the expiry check:

```bash
curl http://localhost:3000/api/notifications/check-expiry
```

### Manually send a test notification:

```typescript
// In browser console
const registration = await navigator.serviceWorker.ready;
registration.showNotification('Test Notification', {
  body: 'This is a test notification',
  icon: '/icon.png',
});
```

## Troubleshooting

- **Notifications not showing**: Check if the browser has granted permission
- **Service worker not registering**: Check browser console for errors
- **VAPID keys missing**: Ensure env variables are set and app is restarted
- **Subscription endpoint invalid**: Ensure correct browser and PWA configuration

## Security Considerations

- Keep `VAPID_PRIVATE_KEY` secret (never share or commit to repository)
- Validate authorization headers in cron endpoints
- Use HTTPS in production (required for service workers and push)
- Implement rate limiting on notification endpoints

## Future Enhancements

- Persistent subscription storage in database
- Multiple notification channels
- User preference controls for notification frequency
- In-app notification center
