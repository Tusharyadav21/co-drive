'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { User, Shield, Sparkles } from 'lucide-react';

interface UsernameInputProps {
  onSubmit: (username: string) => void;
  loading?: boolean;
}

export function UsernameInput({ onSubmit, loading = false }: UsernameInputProps) {
  const [username, setUsername] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (username.trim()) {
      onSubmit(username.trim());
    }
  };

  return (
    <div className="w-full">
      {/* Decorative Brand Card */}
      <div className="relative mb-8 overflow-hidden rounded-3xl bg-linear-to-tr from-zinc-900 via-zinc-800 to-zinc-900 p-6 text-white shadow-xl dark:border dark:border-zinc-800">
        <div className="absolute top-0 right-0 h-32 w-32 rounded-full bg-violet-600/10 blur-3xl" />
        <div className="absolute bottom-0 left-0 h-32 w-32 rounded-full bg-indigo-500/10 blur-3xl" />

        <div className="mb-6 flex items-center gap-2">
          <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-indigo-600 text-white shadow-lg shadow-indigo-600/40">
            <Sparkles className="h-5 w-5" />
          </div>
          <span className="text-lg font-semibold tracking-tight">CoDrive</span>
        </div>

        <h2 className="mb-2 text-2xl font-bold tracking-tight">Sync your family fleet</h2>
        <p className="max-w-[280px] text-xs leading-relaxed text-zinc-400">
          Track fuel logs, service history, and document expiry together in real time.
        </p>

        {/* Floating Mini Feature Badges */}
        <div className="mt-6 flex flex-wrap gap-2">
          <span className="inline-flex items-center gap-1 rounded-full border border-indigo-500/20 bg-indigo-500/10 px-2 py-1 text-[10px] font-medium text-indigo-300">
            <Shield className="h-2.5 w-2.5" /> Shared Access
          </span>
          <span className="inline-flex items-center gap-1 rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2 py-1 text-[10px] font-medium text-emerald-300">
            ✓ Smart Reminders
          </span>
        </div>
      </div>

      {/* Main Login Card */}
      <div className="bg-card rounded-3xl border border-zinc-200/60 p-6 shadow-xs dark:border-zinc-800">
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          <div className="flex flex-col gap-2">
            <label
              htmlFor="username"
              className="text-xs font-semibold tracking-wider text-zinc-500 uppercase dark:text-zinc-400"
            >
              Enter your username
            </label>
            <div className="relative flex items-center">
              <span className="absolute left-4 text-zinc-400">
                <User className="h-4.5 w-4.5" />
              </span>
              <Input
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="e.g. john_doe"
                disabled={loading}
                autoFocus
                className="h-12 rounded-2xl border-zinc-200/80 bg-zinc-50 pr-4 pl-11 text-sm font-medium transition-all focus-visible:ring-indigo-600 focus-visible:ring-offset-0 dark:border-zinc-800 dark:bg-zinc-900"
              />
            </div>
            <p className="text-[10px] leading-tight text-zinc-400">
              Type any username to access your vehicles. If it's a new username, a brand new vault
              will be created instantly.
            </p>
          </div>

          <Button
            type="submit"
            disabled={loading || !username.trim()}
            className="h-12 cursor-pointer rounded-2xl bg-indigo-600 text-sm font-semibold text-white shadow-md shadow-indigo-600/20 transition-all duration-300 hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-indigo-600 dark:hover:bg-indigo-500"
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <svg className="h-4 w-4 animate-spin text-white" fill="none" viewBox="0 0 24 24">
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                Logging in...
              </span>
            ) : (
              'Enter Workspace'
            )}
          </Button>
        </form>
      </div>
    </div>
  );
}
