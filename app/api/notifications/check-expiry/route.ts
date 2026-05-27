import { prisma } from '@/lib/prisma';
import { NextRequest } from 'next/server';
import type { Prisma } from '@prisma/client';
import { handleApiRequest, ApiError } from '@/lib/api-handler';
import { EXPIRY_WARNING_DAYS } from '@/lib/expiry';

type PucWithVehicle = Prisma.PUCGetPayload<{
  include: { vehicle: { include: { owner: { include: { pushSubscriptions: true } } } } };
}>;
type InsuranceWithVehicle = Prisma.InsuranceGetPayload<{
  include: { vehicle: { include: { owner: { include: { pushSubscriptions: true } } } } };
}>;

interface PushSubscriptionData {
  keys?: {
    auth?: string;
    p256dh?: string;
  };
}

function addDays(date: Date, days: number) {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

export async function GET(req: NextRequest) {
  return handleApiRequest(async () => {
    // This endpoint exposes data across all users, so it is intended to be
    // called by a trusted scheduler. When CRON_SECRET is configured, require it.
    const secret = process.env.CRON_SECRET;
    if (secret) {
      const auth = req.headers.get('authorization');
      if (auth !== `Bearer ${secret}`) {
        throw new ApiError(401, 'Unauthorized');
      }
    }

    const now = new Date();
    const pucWindow = addDays(now, EXPIRY_WARNING_DAYS.puc);
    const insuranceWindow = addDays(now, EXPIRY_WARNING_DAYS.insurance);

    const [expiringPUCs, expiringInsurance] = await Promise.all([
      prisma.pUC.findMany({
        where: { expiryDate: { lte: pucWindow, gte: now } },
        include: { vehicle: { include: { owner: { include: { pushSubscriptions: true } } } } },
      }),
      prisma.insurance.findMany({
        where: { expiryDate: { lte: insuranceWindow, gte: now } },
        include: { vehicle: { include: { owner: { include: { pushSubscriptions: true } } } } },
      }),
    ]);

    const notifications = [
      ...expiringPUCs.map((puc: PucWithVehicle) => ({
        user: puc.vehicle.owner.name,
        type: 'PUC',
        vehicle: puc.vehicle.name,
        expiryDate: puc.expiryDate,
        daysLeft: Math.ceil((puc.expiryDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24)),
        endpoint: puc.vehicle.owner.pushSubscriptions?.[0]?.endpoint,
        auth: (puc.vehicle.owner.pushSubscriptions?.[0]?.data as unknown as PushSubscriptionData)
          ?.keys?.auth,
        p256dh: (puc.vehicle.owner.pushSubscriptions?.[0]?.data as unknown as PushSubscriptionData)
          ?.keys?.p256dh,
      })),
      ...expiringInsurance.map((ins: InsuranceWithVehicle) => ({
        user: ins.vehicle.owner.name,
        type: 'Insurance',
        vehicle: ins.vehicle.name,
        expiryDate: ins.expiryDate,
        daysLeft: Math.ceil((ins.expiryDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24)),
        endpoint: ins.vehicle.owner.pushSubscriptions?.[0]?.endpoint,
        auth: (ins.vehicle.owner.pushSubscriptions?.[0]?.data as unknown as PushSubscriptionData)
          ?.keys?.auth,
        p256dh: (ins.vehicle.owner.pushSubscriptions?.[0]?.data as unknown as PushSubscriptionData)
          ?.keys?.p256dh,
      })),
    ];

    return {
      success: true,
      notificationsToSend: notifications.length,
      notifications,
    };
  }, 'Failed to check expiry');
}
