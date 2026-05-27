import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { ArrowLeft } from 'lucide-react';
import { useRouter } from 'next/navigation';

interface FormLayoutProps {
  title: string;
  children: React.ReactNode;
}

export function FormLayout({ title, children }: FormLayoutProps) {
  const router = useRouter();

  return (
    <div className="min-h-screen bg-zinc-50 pb-16 font-sans dark:bg-zinc-950">
      {/* Premium Sticky Header with Blur */}
      <header className="sticky top-0 z-20 border-b border-zinc-200/50 bg-zinc-50/80 backdrop-blur-md transition-all dark:border-zinc-900/50 dark:bg-zinc-950/80">
        <div className="container mx-auto flex h-16 max-w-lg items-center gap-3 px-4">
          <button
            onClick={() => router.back()}
            className="cursor-pointer rounded-xl border border-zinc-200/60 bg-white p-2 text-zinc-500 shadow-2xs transition-colors hover:text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200"
          >
            <ArrowLeft className="h-4 w-4" />
          </button>
          <span className="truncate text-sm leading-tight font-bold">{title}</span>
        </div>
      </header>

      {/* Centered card - optimized layout for mobile first */}
      <main className="container mx-auto max-w-lg px-4 py-6">
        <Card className="overflow-hidden rounded-3xl border border-zinc-200/60 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-900">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-lg font-bold tracking-tight text-zinc-800 dark:text-white">
              {title}
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-2">{children}</CardContent>
        </Card>
      </main>
    </div>
  );
}
