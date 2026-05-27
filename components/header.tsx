'use client';

import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { LogOut } from 'lucide-react';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

interface DashboardHeaderProps {
  username: string;
  avatarInitials: string;
}

const deleteCookie = (name: string) => {
  document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
};

export function Header({ username, avatarInitials }: DashboardHeaderProps) {
  const router = useRouter();

  const handleLogout = () => {
    // 1. Clear localStorage for forms compatibility
    localStorage.removeItem('userId');
    localStorage.removeItem('username');

    // 2. Clear cookies to trigger Server Component reload
    deleteCookie('userId');
    deleteCookie('username');

    // 3. Refresh context
    router.refresh();
  };

  return (
    <header className="sticky top-0 z-20 border-b border-zinc-200/50 bg-zinc-50/80 backdrop-blur-md transition-all dark:border-zinc-900/50 dark:bg-zinc-950/80">
      <div className="container mx-auto flex h-16 max-w-lg items-center justify-between px-4">
        <div className="flex items-center gap-2">
          <div className="relative h-12 w-12 overflow-hidden">
            <Image
              src="/apple-touch-icon.png"
              fill={true}
              sizes="48px"
              preload
              alt="Logo"
              className="object-cover"
            />
          </div>
          <span className="text-base font-bold tracking-tight">CoDrive</span>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 rounded-2xl border border-zinc-200/60 bg-white px-3 py-1.5 shadow-2xs dark:border-zinc-800 dark:bg-zinc-900">
            <Avatar className="flex h-6 w-6 items-center justify-center rounded-lg">
              <AvatarFallback className="flex size-full items-center justify-center rounded-lg bg-indigo-500/10 text-[10px] font-bold text-indigo-600 dark:text-indigo-400">
                {avatarInitials}
              </AvatarFallback>
            </Avatar>
            <span className="max-w-[80px] truncate text-xs font-semibold text-zinc-600 dark:text-zinc-300">
              {username}
            </span>
          </div>

          <button
            onClick={handleLogout}
            className="cursor-pointer rounded-xl border border-zinc-200/60 bg-white p-2 text-zinc-500 shadow-2xs transition-colors hover:text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200"
            title="Logout"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </div>
    </header>
  );
}
