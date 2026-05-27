'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  ArrowLeft,
  Mail,
  KeyRound,
  Bell,
  LogOut,
  CheckCircle2,
  Car,
  Fuel,
  TrendingUp,
} from 'lucide-react';
import Link from 'next/link';
import Image from 'next/image';
import { authClient } from '@/lib/auth-client';
import { initializeNotifications } from '@/lib/notifications';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

interface ProfileClientProps {
  user: {
    id: string;
    name: string;
    email: string;
    image?: string | null;
  };
  stats: {
    totalVehicles: number;
    totalRefills: number;
    totalSpent: number;
  };
}

export function ProfileClient({ user, stats }: ProfileClientProps) {
  const router = useRouter();
  const [registeringPasskey, setRegisteringPasskey] = useState(false);
  const [notificationsStatus, setNotificationsStatus] = useState<string>('default');

  const vapidPublicKey = process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY || null;

  useEffect(() => {
    // Check push notifications permission status on client
    if ('Notification' in window) {
      const status = Notification.permission;
      setTimeout(() => {
        setNotificationsStatus(status);
      }, 0);
    }
  }, []);

  const handleRegisterPasskey = async () => {
    setRegisteringPasskey(true);
    try {
      const passkeyName = `${user.name}'s Passkey (${new Date().toLocaleDateString('en-IN')})`;
      const result = await authClient.passkey.addPasskey({
        name: passkeyName,
      });

      if (result.error) {
        throw new Error(result.error.message || 'Failed to register passkey');
      }

      toast.success(
        'Passkey registered successfully! You can now log in securely using biometrics.'
      );
    } catch (err) {
      console.error(err);
      toast.error(err instanceof Error ? err.message : 'Biometric registration failed');
    } finally {
      setRegisteringPasskey(false);
    }
  };

  const handleEnableNotifications = async () => {
    if (!vapidPublicKey) {
      toast.error('Push notifications are not configured (VAPID key missing)');
      return;
    }

    try {
      await initializeNotifications(vapidPublicKey, user.id);
      if ('Notification' in window) {
        setNotificationsStatus(Notification.permission);
        if (Notification.permission === 'granted') {
          toast.success('Push notifications successfully enabled!');
        } else {
          toast.error('Notification permission denied by browser.');
        }
      }
    } catch (err) {
      console.error(err);
      toast.error('Failed to configure notifications.');
    }
  };

  const handleSignOut = async () => {
    try {
      await authClient.signOut();
      localStorage.removeItem('userId'); // Cleanup local user tracking if any
      toast.success('Logged out successfully');
      router.push('/');
      router.refresh();
    } catch (err) {
      console.error(err);
      toast.error('Failed to log out.');
    }
  };

  const avatarInitials = user.name.slice(0, 2).toUpperCase();

  return (
    <div className="min-h-screen bg-zinc-50 pb-24 font-sans text-zinc-900 antialiased dark:bg-zinc-950 dark:text-zinc-50">
      <header className="sticky top-0 z-20 border-b border-zinc-200/50 bg-zinc-50/80 backdrop-blur-md transition-all dark:border-zinc-900/50 dark:bg-zinc-950/80">
        <div className="container mx-auto flex h-16 max-w-lg items-center gap-3 px-4">
          <Link href="/">
            <button className="cursor-pointer rounded-xl border border-zinc-200/60 bg-white p-2 text-zinc-500 shadow-2xs transition-colors hover:text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200">
              <ArrowLeft className="h-4 w-4" />
            </button>
          </Link>
          <div className="min-w-0 flex-1">
            <h1 className="text-sm leading-tight font-bold">Your Profile</h1>
            <p className="mt-0.5 text-[10px] font-semibold text-zinc-500 dark:text-zinc-400">
              Manage your credentials & system preferences
            </p>
          </div>
        </div>
      </header>

      <main className="container mx-auto flex max-w-lg flex-col gap-6 px-4 py-6">
        {/* User Card */}
        <section className="relative overflow-hidden rounded-3xl bg-linear-to-tr from-zinc-900 via-zinc-800 to-zinc-900 p-5 text-white shadow-xl dark:border dark:border-zinc-800">
          <div className="absolute top-0 right-0 h-32 w-32 rounded-full bg-indigo-500/10 blur-3xl" />
          <div className="relative z-10 flex items-center gap-4">
            {user.image ? (
              <Image
                src={user.image}
                alt={user.name}
                width={56}
                height={56}
                className="h-14 w-14 rounded-2xl border-2 border-white/10 object-cover shadow-md"
              />
            ) : (
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl border-2 border-white/10 bg-indigo-600 text-base font-black text-white shadow-md">
                {avatarInitials}
              </div>
            )}
            <div className="min-w-0 flex-1">
              <span className="flex items-center gap-1 text-[9px] leading-none font-bold tracking-wider text-indigo-300 uppercase">
                <CheckCircle2 className="h-3 w-3 fill-emerald-400/20 text-emerald-400" /> Verified
                Account
              </span>
              <h2 className="mt-0.5 truncate text-lg font-bold tracking-tight">{user.name}</h2>
              <p className="mt-0.5 flex items-center gap-1 truncate text-xs text-zinc-400">
                <Mail className="h-3 w-3 text-zinc-400" />
                {user.email}
              </p>
            </div>
          </div>
        </section>

        {/* Dynamic Stats Grid */}
        <section className="grid grid-cols-3 gap-3">
          <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
            <span className="flex items-center gap-1.5 text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
              <Car className="h-3.5 w-3.5 text-indigo-500" /> Vehicles
            </span>
            <span className="text-xl font-black">{stats.totalVehicles}</span>
          </div>

          <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
            <span className="flex items-center gap-1.5 text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
              <Fuel className="h-3.5 w-3.5 text-emerald-500" /> Refills
            </span>
            <span className="text-xl font-black">{stats.totalRefills}</span>
          </div>

          <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
            <span className="flex items-center gap-1.5 text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
              <TrendingUp className="h-3.5 w-3.5 text-indigo-500" /> Spent
            </span>
            <div className="flex items-baseline gap-0.5 overflow-hidden text-ellipsis">
              <span className="text-[10px] font-bold text-zinc-400">₹</span>
              <span className="text-lg font-black tracking-tight">
                {stats.totalSpent.toLocaleString('en-IN')}
              </span>
            </div>
          </div>
        </section>

        {/* Security & Credentials (Passkeys) Card */}
        <section className="shadow-3xs flex flex-col gap-4 rounded-3xl border border-zinc-200/60 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
          <div className="flex items-center gap-2 border-b border-zinc-100 pb-2 dark:border-zinc-800">
            <KeyRound className="h-4 w-4 text-indigo-500" />
            <h3 className="text-xs font-bold tracking-wide text-zinc-500 uppercase dark:text-zinc-400">
              Security Credentials
            </h3>
          </div>
          <div>
            <h4 className="text-xs font-bold">Biometric Logins (Passkeys)</h4>
            <p className="mt-1 text-[10px] leading-normal text-zinc-400">
              Register a passkey to securely sign in using your device's fingerprint or facial
              recognition scanner. No password required.
            </p>
          </div>
          <Button
            onClick={handleRegisterPasskey}
            disabled={registeringPasskey}
            className="h-10 w-full cursor-pointer rounded-2xl bg-indigo-600 text-xs font-semibold text-white shadow-md shadow-indigo-600/10 transition-all hover:bg-indigo-500"
          >
            {registeringPasskey ? 'Registering...' : 'Register Secure Passkey'}
          </Button>
        </section>

        {/* Notifications Preference Card */}
        <section className="shadow-3xs flex flex-col gap-4 rounded-3xl border border-zinc-200/60 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
          <div className="flex items-center gap-2 border-b border-zinc-100 pb-2 dark:border-zinc-800">
            <Bell className="h-4 w-4 text-emerald-500" />
            <h3 className="text-xs font-bold tracking-wide text-zinc-500 uppercase dark:text-zinc-400">
              Push Notifications
            </h3>
          </div>
          <div className="flex items-start justify-between gap-4">
            <div className="flex flex-col gap-0.5">
              <h4 className="text-xs font-bold">Expiry & Maintenance Alerts</h4>
              <p className="text-[10px] leading-normal text-zinc-400">
                Receive proactive warnings directly on your device when your vehicle's PUC,
                insurance, or scheduled service window is approaching.
              </p>
            </div>
            <div className="flex shrink-0 items-center gap-1.5 rounded-full border border-zinc-200/60 bg-zinc-50 px-2.5 py-1 text-[9px] font-extrabold uppercase dark:border-zinc-800 dark:bg-zinc-950">
              <span
                className={`h-1.5 w-1.5 rounded-full ${notificationsStatus === 'granted' ? 'bg-emerald-500 shadow-xs shadow-emerald-500' : 'bg-amber-500'}`}
              />
              <span
                className={
                  notificationsStatus === 'granted'
                    ? 'text-emerald-600 dark:text-emerald-400'
                    : 'text-amber-600'
                }
              >
                {notificationsStatus === 'granted'
                  ? 'Enabled'
                  : notificationsStatus === 'denied'
                    ? 'Blocked'
                    : 'Configure'}
              </span>
            </div>
          </div>
          {notificationsStatus !== 'granted' && (
            <Button
              onClick={handleEnableNotifications}
              className="h-10 w-full cursor-pointer rounded-2xl bg-zinc-900 text-xs font-bold text-white shadow-md transition-all hover:bg-zinc-800 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-100"
            >
              Configure Push Notifications
            </Button>
          )}
        </section>

        {/* Action Options (Logout) */}
        <section className="mt-2 flex flex-col gap-3">
          <button
            onClick={handleSignOut}
            className="flex h-11 w-full cursor-pointer items-center justify-center gap-2 rounded-2xl border-2 border-red-500/20 bg-red-500/5 text-xs font-bold text-red-500 transition-all hover:bg-red-500/10 active:scale-98"
          >
            <LogOut className="h-4 w-4" />
            Secure Log Out
          </button>
        </section>
      </main>
    </div>
  );
}
