import { prisma } from '../prisma';

export async function addMileageEntry(data: {
  vehicleId: string;
  userId: string;
  mileage: number;
  notes?: string;
  fuelLitres?: number;
  fuelAmount?: number;
  date?: string;
}) {
  return prisma.mileageEntry.create({
    data: {
      vehicleId: data.vehicleId,
      userId: data.userId,
      mileage: data.mileage,
      notes: data.notes,
      fuelLitres: data.fuelLitres,
      fuelAmount: data.fuelAmount,
      date: data.date ? new Date(data.date) : new Date(),
    },
  });
}

export async function getMileageEntries(vehicleId: string, userId: string) {
  return prisma.mileageEntry.findMany({
    where: {
      vehicleId,
      vehicle: {
        OR: [{ ownerId: userId }, { users: { some: { userId } } }],
      },
    },
    orderBy: { date: 'desc' },
  });
}
