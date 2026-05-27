'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { FormLayout } from './form-layout';
import { useAuth } from '@/hooks/use-auth';
import { addMileageAction } from '@/lib/actions/mileage';

interface AddMileageFormProps {
  vehicleId: string;
}

export function AddMileageForm({ vehicleId }: AddMileageFormProps) {
  const { userId } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [formData, setFormData] = useState({
    mileage: '',
    date: new Date().toISOString().split('T')[0],
    notes: '',
    fuelLitres: '',
    fuelAmount: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId) {
      setError('User not found');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const response = await addMileageAction(
        {
          vehicleId,
          mileage: parseInt(formData.mileage),
          date: formData.date,
          notes: formData.notes,
          fuelLitres: formData.fuelLitres ? parseFloat(formData.fuelLitres) : undefined,
          fuelAmount: formData.fuelAmount ? parseFloat(formData.fuelAmount) : undefined,
        },
        userId
      );

      if (response.error) {
        throw new Error(response.error);
      }

      router.push(`/vehicles/${vehicleId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add mileage');
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormLayout title="Add Mileage & Fuel Entry">
      <form onSubmit={handleSubmit} className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <label htmlFor="mileage" className="text-sm font-medium">
            Current Odometer Reading (km) *
          </label>
          <Input
            id="mileage"
            name="mileage"
            type="number"
            value={formData.mileage}
            onChange={handleChange}
            placeholder="e.g., 50000"
            required
            min="0"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="flex flex-col gap-2">
            <label htmlFor="fuelLitres" className="text-sm font-medium">
              Fuel (Litres)
            </label>
            <Input
              id="fuelLitres"
              name="fuelLitres"
              type="number"
              step="0.01"
              value={formData.fuelLitres}
              onChange={handleChange}
              placeholder="e.g., 25.21"
              min="0"
            />
          </div>
          <div className="flex flex-col gap-2">
            <label htmlFor="fuelAmount" className="text-sm font-medium">
              Total Amount
            </label>
            <Input
              id="fuelAmount"
              name="fuelAmount"
              type="number"
              step="0.01"
              value={formData.fuelAmount}
              onChange={handleChange}
              placeholder="e.g., 2500"
              min="0"
            />
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <label htmlFor="date" className="text-sm font-medium">
            Date *
          </label>
          <Input
            id="date"
            name="date"
            type="date"
            value={formData.date}
            onChange={handleChange}
            required
          />
        </div>

        <div className="flex flex-col gap-2">
          <label htmlFor="notes" className="text-sm font-medium">
            Notes
          </label>
          <textarea
            id="notes"
            name="notes"
            value={formData.notes}
            onChange={handleChange}
            placeholder="e.g., After service, on highway"
            className="focus:ring-ring rounded-lg border px-3 py-2 text-sm focus:ring-2 focus:outline-none"
            rows={3}
          />
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive rounded-lg p-3 text-sm">{error}</div>
        )}

        <Button type="submit" disabled={loading}>
          {loading ? 'Adding...' : 'Add Entry'}
        </Button>
      </form>
    </FormLayout>
  );
}
