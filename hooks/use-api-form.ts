import { useState } from 'react';
import { useRouter } from 'next/navigation';

export function useApiForm<T, R = unknown>(
  apiEndpoint: string,
  successRedirect?: string | ((data: R) => string)
) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

  async function submitForm(data: T) {
    setError('');
    setLoading(true);
    try {
      const response = await fetch(apiEndpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const result = await response.json();
        throw new Error(result.error || 'Something went wrong');
      }

      const resultData = (await response.json()) as R;

      if (successRedirect) {
        const path =
          typeof successRedirect === 'function' ? successRedirect(resultData) : successRedirect;
        router.push(path);
      }

      return resultData;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unexpected error occurred');
      return null;
    } finally {
      setLoading(false);
    }
  }

  return { loading, error, setError, submitForm };
}
