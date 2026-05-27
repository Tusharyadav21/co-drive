import { prisma } from '../prisma';
import type { VehicleWithDetails } from '@/types';

export async function getVehiclesByUserId(userId: string): Promise<VehicleWithDetails[]> {
  return prisma.vehicle.findMany({
    where: {
      OR: [{ ownerId: userId }, { users: { some: { userId } } }],
    },
    include: {
      users: {
        include: {
          user: true,
        },
      },
      mileageEntries: {
        orderBy: { date: 'desc' },
      },
      pucRecords: {
        orderBy: { expiryDate: 'desc' },
      },
      insuranceRecords: {
        orderBy: { expiryDate: 'desc' },
      },
    },
  });
}

export async function getVehicleById(
  vehicleId: string,
  userId: string
): Promise<VehicleWithDetails | null> {
  return prisma.vehicle.findFirst({
    where: {
      id: vehicleId,
      OR: [{ ownerId: userId }, { users: { some: { userId } } }],
    },
    include: {
      users: {
        include: {
          user: true,
        },
      },
      mileageEntries: {
        orderBy: { date: 'desc' },
      },
      pucRecords: {
        orderBy: { expiryDate: 'desc' },
      },
      insuranceRecords: {
        orderBy: { expiryDate: 'desc' },
      },
    },
  });
}

export async function createVehicle(data: {
  name: string;
  make: string;
  model: string;
  year: number;
  licensePlate: string;
  ownerId: string;
}) {
  return prisma.vehicle.create({
    data: {
      ...data,
      users: {
        create: {
          userId: data.ownerId,
          role: 'owner',
        },
      },
    },
  });
}

export async function shareVehicle(vehicleId: string, targetUserId: string) {
  const existing = await prisma.vehicleUser.findUnique({
    where: {
      userId_vehicleId: {
        userId: targetUserId,
        vehicleId,
      },
    },
  });

  if (existing) {
    throw new Error('User already has access to this vehicle');
  }

  return prisma.vehicleUser.create({
    data: {
      userId: targetUserId,
      vehicleId,
      role: 'viewer',
    },
    include: {
      user: true,
    },
  });
}

export async function removeVehicleShare(vehicleUserId: string, requestingUserId: string) {
  const vehicleUser = await prisma.vehicleUser.findUnique({
    where: { id: vehicleUserId },
    include: { vehicle: true },
  });

  if (!vehicleUser) {
    throw new Error('Share not found');
  }

  // Only the owner or the user themselves can remove access
  if (vehicleUser.vehicle.ownerId !== requestingUserId && vehicleUser.userId !== requestingUserId) {
    throw new Error('Unauthorized to remove access');
  }

  return prisma.vehicleUser.delete({
    where: { id: vehicleUserId },
  });
}
