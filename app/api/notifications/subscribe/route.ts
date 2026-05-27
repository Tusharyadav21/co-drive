import { prisma } from '@/lib/prisma';
import { NextRequest } from 'next/server';
import { z } from 'zod';
import type { Prisma } from '@prisma/client';
import { handleApiRequest } from '@/lib/api-handler';

const subscribeSchema = z.object({
  userId: z.string().min(1, 'User ID is required'),
  subscription: z
    .object({ endpoint: z.string().url('Invalid subscription endpoint') })
    .passthrough(),
});

export async function POST(req: NextRequest) {
  return handleApiRequest(async () => {
    const body = await req.json();
    const { userId, subscription } = subscribeSchema.parse(body);

    // Persist the subscription so it survives restarts and is shared across
    // instances. Keyed by endpoint so re-subscribing updates in place.
    await prisma.pushSubscription.upsert({
      where: { endpoint: subscription.endpoint },
      create: {
        userId,
        endpoint: subscription.endpoint,
        data: subscription as Prisma.InputJsonValue,
      },
      update: {
        userId,
        data: subscription as Prisma.InputJsonValue,
      },
    });

    return { success: true };
  }, 'Failed to store subscription');
}

export async function GET() {
  return handleApiRequest(async () => {
    return {
      subscriptions: await prisma.pushSubscription.count(),
    };
  }, 'Failed to fetch subscriptions');
}
