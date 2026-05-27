'use server';

import { revalidatePath } from 'next/cache';
import { CreateVehicleSchema, ShareVehicleSchema, RemoveShareSchema } from '@/types';
import { createVehicle, shareVehicle, removeVehicleShare } from '../db/vehicles';
import { prisma } from '../prisma';

export async function createVehicleAction(formData: unknown, ownerId: string) {
  try {
    const data = CreateVehicleSchema.parse(formData);
    const vehicle = await createVehicle({ ...data, ownerId });
    revalidatePath('/');
    return { data: vehicle };
  } catch (error) {
    console.error('Failed to create vehicle', error);
    return { error: error instanceof Error ? error.message : 'Failed to create vehicle' };
  }
}

export async function shareVehicleAction(formData: unknown, ownerId: string) {
  try {
    const data = ShareVehicleSchema.parse(formData);

    // Find target user
    const targetUser = await prisma.user.findUnique({
      where: { email: data.email },
    });

    if (!targetUser) {
      return { error: 'User not found' };
    }

    if (targetUser.id === ownerId) {
      return { error: 'Cannot share with yourself' };
    }

    const share = await shareVehicle(data.vehicleId, targetUser.id);
    revalidatePath(`/vehicles/${data.vehicleId}`);
    return { data: share };
  } catch (error) {
    console.error('Failed to share vehicle', error);
    return { error: error instanceof Error ? error.message : 'Failed to share vehicle' };
  }
}

export async function removeShareAction(formData: unknown, requestingUserId: string) {
  try {
    const data = RemoveShareSchema.parse(formData);
    await removeVehicleShare(data.vehicleUserId, requestingUserId);
    // Since we don't have vehicleId easily accessible without another DB hit, revalidate all paths
    revalidatePath('/');
    return { data: { success: true } };
  } catch (error) {
    console.error('Failed to remove share', error);
    return { error: error instanceof Error ? error.message : 'Failed to remove share' };
  }
}
