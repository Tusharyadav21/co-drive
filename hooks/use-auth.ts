import { authClient } from '@/lib/auth-client';

export function useAuth() {
  const { data: session, isPending } = authClient.useSession();

  return {
    userId: session?.user?.id || null,
    user: session?.user || null,
    loading: isPending,
  };
}
