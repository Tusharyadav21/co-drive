import { prisma } from '@/lib/prisma';
import { ApiError } from '@/lib/api-handler';

/**
 * Ensures the acting user owns or has been granted access to the vehicle.
 * Throws 403 otherwise. This is the data-layer guard that prevents a user from
 * reading or mutating vehicles that are not theirs.
 *
 * NOTE: this verifies *authorization* given a userId. The app still identifies
 * users by username alone (no secret/session), so the userId itself is not yet
 * authenticated — adding a real login secret is a separate, product-level step.
 */
export async function assertVehicleAccess(userId: string, vehicleId: string) {
  const vehicle = await prisma.vehicle.findFirst({
    where: {
      id: vehicleId,
      OR: [{ ownerId: userId }, { users: { some: { userId } } }],
    },
    select: { id: true },
  });

  if (!vehicle) {
    throw new ApiError(403, 'You do not have access to this vehicle');
  }
}

/** Ensures the acting user is the owner of the vehicle. Throws 403 otherwise. */
export async function assertVehicleOwner(userId: string, vehicleId: string) {
  const vehicle = await prisma.vehicle.findFirst({
    where: { id: vehicleId, ownerId: userId },
    select: { id: true },
  });

  if (!vehicle) {
    throw new ApiError(403, 'Only the vehicle owner can perform this action');
  }
}
