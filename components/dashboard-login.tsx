'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle, Fingerprint, Terminal } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { authClient } from '@/lib/auth-client';
import { devLoginAction } from '@/lib/actions/dev-login';

export function DashboardLogin() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const router = useRouter();

  const isDevLoginEnabled = process.env.NEXT_PUBLIC_DEV_LOGIN_ENABLED === 'true';

  const handleGoogleLogin = async () => {
    setLoading(true);
    setError('');
    try {
      await authClient.signIn.social({
        provider: 'google',
        callbackURL: '/',
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to login with Google');
      setLoading(false);
    }
  };

  const handlePasskeyLogin = async () => {
    setLoading(true);
    setError('');
    try {
      const { data: _data, error } = await authClient.signIn.passkey();
      if (error) {
        throw new Error(error.message || 'Passkey login failed');
      }
      router.push('/');
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to login with Passkey');
      setLoading(false);
    }
  };

  const handleDevLogin = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await devLoginAction();
      if (res.error) {
        throw new Error(res.error);
      }
      router.push('/');
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Developer login failed');
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 p-4 dark:bg-zinc-950">
      <div className="pointer-events-none absolute top-1/4 left-1/2 h-[350px] w-[350px] -translate-x-1/2 rounded-full bg-indigo-500/10 blur-[80px] dark:bg-indigo-600/5" />
      <div className="relative z-10 flex w-full max-w-sm flex-col gap-4">
        <div className="mb-6 text-center">
          <h1 className="mb-2 text-3xl font-extrabold tracking-tight text-zinc-900 dark:text-white">
            CoDrive
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            The collaborative home for your shared vehicles
          </p>
        </div>

        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Login Failed</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {isDevLoginEnabled && (
          <div className="group relative mb-2 rounded-xl bg-gradient-to-r from-amber-500 via-orange-500 to-yellow-500 p-[1px] shadow-md transition-all duration-300 hover:shadow-lg hover:shadow-orange-500/20 dark:from-amber-600 dark:via-orange-600 dark:to-yellow-600">
            <div className="pointer-events-none absolute -inset-0.5 animate-pulse rounded-xl bg-gradient-to-r from-amber-500 to-orange-500 opacity-30 blur transition duration-300 group-hover:opacity-60" />
            <Button
              onClick={handleDevLogin}
              disabled={loading}
              className="relative flex h-12 w-full items-center justify-center rounded-[11px] border-0 bg-white font-semibold text-zinc-900 shadow-inner hover:bg-zinc-50 dark:bg-zinc-950 dark:text-zinc-50 dark:hover:bg-zinc-900"
            >
              <Terminal className="mr-2 h-5 w-5 animate-pulse text-amber-500 group-hover:animate-bounce dark:text-amber-400" />
              Developer Bypass (Local Env)
            </Button>
          </div>
        )}

        <Button
          onClick={handleGoogleLogin}
          disabled={loading}
          className="h-12 w-full border border-zinc-200 bg-white text-zinc-800 shadow-sm hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
        >
          <svg
            className="mr-2 h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              fill="#4285F4"
            />
            <path
              d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              fill="#34A853"
            />
            <path
              d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              fill="#FBBC05"
            />
            <path
              d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              fill="#EA4335"
            />
          </svg>
          Sign in with Google
        </Button>

        <div className="relative flex items-center py-2">
          <div className="grow border-t border-zinc-200 dark:border-zinc-800"></div>
          <span className="mx-4 shrink-0 text-xs text-zinc-400">or</span>
          <div className="grow border-t border-zinc-200 dark:border-zinc-800"></div>
        </div>

        <Button
          onClick={handlePasskeyLogin}
          disabled={loading}
          className="h-12 w-full bg-zinc-900 text-white shadow-sm hover:bg-zinc-800 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-200"
        >
          <Fingerprint className="mr-2 h-5 w-5" />
          Sign in with Passkey
        </Button>
      </div>
    </div>
  );
}
