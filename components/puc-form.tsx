'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { FormLayout } from './form-layout';
import { useAuth } from '@/hooks/use-auth';
import { updatePUCAction } from '@/lib/actions/document';

interface PUCFormProps {
  vehicleId: string;
}

export function PUCForm({ vehicleId }: PUCFormProps) {
  const { userId } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [formData, setFormData] = useState({
    expiryDate: '',
    certificateNumber: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
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
      const response = await updatePUCAction({
        vehicleId,
        ...formData,
      });

      if (response.error) {
        throw new Error(response.error);
      }

      router.push(`/vehicles/${vehicleId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update PUC');
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormLayout title="PUC Certificate">
      <form onSubmit={handleSubmit} className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <label htmlFor="expiryDate" className="text-sm font-medium">
            Expiry Date *
          </label>
          <Input
            id="expiryDate"
            name="expiryDate"
            type="date"
            value={formData.expiryDate}
            onChange={handleChange}
            required
          />
        </div>

        <div className="flex flex-col gap-2">
          <label htmlFor="certificateNumber" className="text-sm font-medium">
            Certificate Number
          </label>
          <Input
            id="certificateNumber"
            name="certificateNumber"
            value={formData.certificateNumber}
            onChange={handleChange}
            placeholder="e.g., PUC123456"
          />
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive rounded-lg p-3 text-sm">{error}</div>
        )}

        <Button type="submit" disabled={loading}>
          {loading ? 'Saving...' : 'Save PUC'}
        </Button>
      </form>
    </FormLayout>
  );
}
