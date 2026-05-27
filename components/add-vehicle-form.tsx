'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { FormLayout } from './form-layout';
import { useAuth } from '@/hooks/use-auth';
import { createVehicleAction } from '@/lib/actions/vehicle';

export function AddVehicleForm() {
  const { userId } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [formData, setFormData] = useState({
    name: '',
    make: '',
    model: '',
    year: new Date().getFullYear(),
    licensePlate: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: name === 'year' ? parseInt(value) : value,
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
      const response = await createVehicleAction({ ...formData }, userId);

      if (response.error) {
        throw new Error(response.error);
      }

      if (response.data) {
        router.push(`/vehicles/${response.data.id}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create vehicle');
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormLayout title="Add New Vehicle">
      <form onSubmit={handleSubmit} className="flex flex-col gap-6">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="flex flex-col gap-2">
            <label htmlFor="name" className="text-sm font-medium">
              Vehicle Name *
            </label>
            <Input
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="e.g., My Honda Civic"
              required
            />
          </div>
          <div className="flex flex-col gap-2">
            <label htmlFor="make" className="text-sm font-medium">
              Make *
            </label>
            <Input
              id="make"
              name="make"
              value={formData.make}
              onChange={handleChange}
              placeholder="e.g., Honda"
              required
            />
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="flex flex-col gap-2">
            <label htmlFor="model" className="text-sm font-medium">
              Model *
            </label>
            <Input
              id="model"
              name="model"
              value={formData.model}
              onChange={handleChange}
              placeholder="e.g., Civic"
              required
            />
          </div>
          <div className="flex flex-col gap-2">
            <label htmlFor="year" className="text-sm font-medium">
              Year *
            </label>
            <Input
              id="year"
              name="year"
              type="number"
              value={formData.year}
              onChange={handleChange}
              min="1900"
              max={new Date().getFullYear() + 1}
              required
            />
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <label htmlFor="licensePlate" className="text-sm font-medium">
            License Plate *
          </label>
          <Input
            id="licensePlate"
            name="licensePlate"
            value={formData.licensePlate}
            onChange={handleChange}
            placeholder="e.g., ABC123"
            required
          />
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive rounded-lg p-3 text-sm">{error}</div>
        )}

        <Button type="submit" disabled={loading}>
          {loading ? 'Creating...' : 'Create Vehicle'}
        </Button>
      </form>
    </FormLayout>
  );
}
