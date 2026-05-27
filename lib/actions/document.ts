'use server';

import { revalidatePath } from 'next/cache';
import { UpdatePUCSchema, UpdateInsuranceSchema } from '@/types';
import { updatePUC, updateInsurance } from '../db/documents';

export async function updatePUCAction(formData: unknown) {
  try {
    const data = UpdatePUCSchema.parse(formData);
    const puc = await updatePUC(data);
    revalidatePath(`/vehicles/${data.vehicleId}`);
    return { data: puc };
  } catch (error) {
    console.error('Failed to update PUC', error);
    return { error: error instanceof Error ? error.message : 'Failed to update PUC' };
  }
}

export async function updateInsuranceAction(formData: unknown) {
  try {
    const data = UpdateInsuranceSchema.parse(formData);
    const insurance = await updateInsurance(data);
    revalidatePath(`/vehicles/${data.vehicleId}`);
    return { data: insurance };
  } catch (error) {
    console.error('Failed to update Insurance', error);
    return { error: error instanceof Error ? error.message : 'Failed to update Insurance' };
  }
}
