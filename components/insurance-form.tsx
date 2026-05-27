'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { FormLayout } from './form-layout';
import { useAuth } from '@/hooks/use-auth';
import { updateInsuranceAction } from '@/lib/actions/document';

interface InsuranceFormProps {
  vehicleId: string;
}

export function InsuranceForm({ vehicleId }: InsuranceFormProps) {
  const { userId } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [formData, setFormData] = useState({
    expiryDate: '',
    policyNumber: '',
    provider: '',
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
      const response = await updateInsuranceAction({
        vehicleId,
        ...formData,
      });

      if (response.error) {
        throw new Error(response.error);
      }

      router.push(`/vehicles/${vehicleId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update Insurance');
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormLayout title="Insurance Policy">
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
          <label htmlFor="provider" className="text-sm font-medium">
            Provider
          </label>
          <Input
            id="provider"
            name="provider"
            value={formData.provider}
            onChange={handleChange}
            placeholder="e.g., HDFC Insurance"
          />
        </div>

        <div className="flex flex-col gap-2">
          <label htmlFor="policyNumber" className="text-sm font-medium">
            Policy Number
          </label>
          <Input
            id="policyNumber"
            name="policyNumber"
            value={formData.policyNumber}
            onChange={handleChange}
            placeholder="e.g., POL123456"
          />
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive rounded-lg p-3 text-sm">{error}</div>
        )}

        <Button type="submit" disabled={loading}>
          {loading ? 'Saving...' : 'Save Insurance'}
        </Button>
      </form>
    </FormLayout>
  );
}
