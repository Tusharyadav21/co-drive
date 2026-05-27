'use server';

import { revalidatePath } from 'next/cache';
import { AddMileageSchema } from '@/types';
import { addMileageEntry } from '../db/mileage';

export async function addMileageAction(formData: unknown, userId: string) {
  try {
    const data = AddMileageSchema.parse(formData);
    const entry = await addMileageEntry({ ...data, userId });
    revalidatePath(`/vehicles/${data.vehicleId}`);
    return { data: entry };
  } catch (error) {
    console.error('Failed to add mileage', error);
    return { error: error instanceof Error ? error.message : 'Failed to add mileage' };
  }
}
