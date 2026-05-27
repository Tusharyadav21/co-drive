import { z } from 'zod';
import type { User, Vehicle, VehicleUser, MileageEntry, PUC, Insurance } from '@prisma/client';

export const CreateVehicleSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  make: z.string().min(1, 'Make is required'),
  model: z.string().min(1, 'Model is required'),
  year: z
    .number()
    .int()
    .min(1900)
    .max(new Date().getFullYear() + 1),
  licensePlate: z.string().min(1, 'License plate is required'),
});

export const ShareVehicleSchema = z.object({
  vehicleId: z.string().min(1, 'Vehicle ID is required'),
  email: z.string().email('Valid email is required'),
});

export const RemoveShareSchema = z.object({
  vehicleUserId: z.string().min(1, 'VehicleUser ID is required'),
});

export const AddMileageSchema = z.object({
  vehicleId: z.string().min(1, 'Vehicle ID is required'),
  mileage: z.number().int().positive('Mileage must be positive'),
  notes: z.string().optional(),
  fuelLitres: z.number().positive('Must be positive').optional(),
  fuelAmount: z.number().positive('Must be positive').optional(),
  date: z.string().optional(),
});

export const UpdatePUCSchema = z.object({
  vehicleId: z.string().min(1, 'Vehicle ID is required'),
  expiryDate: z.string().min(1, 'Expiry date is required'),
  certificateNumber: z.string().optional(),
});

export const UpdateInsuranceSchema = z.object({
  vehicleId: z.string().min(1, 'Vehicle ID is required'),
  expiryDate: z.string().min(1, 'Expiry date is required'),
  policyNumber: z.string().optional(),
  provider: z.string().optional(),
});

// Infered Types
export type CreateVehicleInput = z.infer<typeof CreateVehicleSchema>;
export type ShareVehicleInput = z.infer<typeof ShareVehicleSchema>;
export type RemoveShareInput = z.infer<typeof RemoveShareSchema>;
export type AddMileageInput = z.infer<typeof AddMileageSchema>;
export type UpdatePUCInput = z.infer<typeof UpdatePUCSchema>;
export type UpdateInsuranceInput = z.infer<typeof UpdateInsuranceSchema>;

// Extened types for Frontend
export type VehicleWithDetails = Vehicle & {
  users: (VehicleUser & { user: User | null })[];
  mileageEntries: MileageEntry[];
  pucRecords: PUC[];
  insuranceRecords: Insurance[];
};
