'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Sparkles, PlusCircle, User } from 'lucide-react';

export function Navigation() {
  const pathname = usePathname();

  const isWorkspace = pathname === '/';
  const isNew = pathname === '/vehicles/new';
  const isProfile = pathname === '/profile';

  return (
    <nav className="fixed bottom-4 left-1/2 z-30 flex h-14 w-[calc(100%-32px)] max-w-sm -translate-x-1/2 items-center justify-around rounded-2xl border border-zinc-200/60 bg-white/80 px-4 shadow-lg backdrop-blur-md dark:border-zinc-900/50 dark:bg-zinc-950/80">
      <Link
        href="/"
        className={`flex cursor-pointer flex-col items-center gap-0.5 transition-colors ${
          isWorkspace
            ? 'font-bold text-indigo-600 dark:text-indigo-400'
            : 'font-medium text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-100'
        }`}
      >
        <Sparkles className="h-4.5 w-4.5" />
        <span className="text-[9px]">Workspace</span>
      </Link>

      <Link
        href="/vehicles/new"
        className={`flex cursor-pointer flex-col items-center gap-0.5 transition-colors ${
          isNew
            ? 'font-bold text-indigo-600 dark:text-indigo-400'
            : 'font-medium text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-100'
        }`}
      >
        <PlusCircle className="h-4.5 w-4.5" />
        <span className="text-[9px]">New</span>
      </Link>

      <Link
        href="/profile"
        className={`flex cursor-pointer flex-col items-center gap-0.5 transition-colors ${
          isProfile
            ? 'font-bold text-indigo-600 dark:text-indigo-400'
            : 'font-medium text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-100'
        }`}
      >
        <User className="h-4.5 w-4.5" />
        <span className="text-[9px]">Profile</span>
      </Link>
    </nav>
  );
}
