'use client';

import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ArrowLeft, Plus, Sparkles, PlusCircle, Users } from 'lucide-react';
import Link from 'next/link';
import type { VehicleWithDetails } from '@/types';

import { VehicleHero } from './vehicles/vehicle-hero';
import { MileageTab } from './vehicles/mileage-tab';
import { DocumentsTab } from './vehicles/documents-tab';
import { SharesTab } from './vehicles/shares-tab';

interface VehicleDetailProps {
  vehicle: VehicleWithDetails;
  currentUserId: string;
}

export function VehicleDetail({ vehicle, currentUserId }: VehicleDetailProps) {
  const [activeTab, setActiveTab] = useState('mileage');

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
            <h1 className="truncate text-sm leading-tight font-bold">{vehicle.name}</h1>
            <p className="mt-0.5 text-[10px] leading-none font-semibold text-zinc-500 dark:text-zinc-400">
              {vehicle.year} {vehicle.make} {vehicle.model}
            </p>
          </div>
        </div>
      </header>

      <main className="container mx-auto flex max-w-lg flex-col gap-6 px-4 py-6">
        <VehicleHero vehicle={vehicle} />

        <section className="w-full">
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="shadow-3xs grid h-11 w-full grid-cols-3 rounded-2xl border border-zinc-200/60 bg-white p-1 dark:border-zinc-800 dark:bg-zinc-900">
              <TabsTrigger
                value="mileage"
                className="cursor-pointer rounded-xl text-[10px] font-bold transition-all"
              >
                Logs & Chart
              </TabsTrigger>
              <TabsTrigger
                value="documents"
                className="cursor-pointer rounded-xl text-[10px] font-bold transition-all"
              >
                Documents
              </TabsTrigger>
              <TabsTrigger
                value="shares"
                className="cursor-pointer rounded-xl text-[10px] font-bold transition-all"
              >
                Family Board
              </TabsTrigger>
            </TabsList>

            <TabsContent value="mileage" className="mt-4">
              <MileageTab vehicle={vehicle} />
            </TabsContent>

            <TabsContent value="documents" className="mt-4">
              <DocumentsTab vehicle={vehicle} />
            </TabsContent>

            <TabsContent value="shares" className="mt-4">
              <SharesTab vehicle={vehicle} currentUserId={currentUserId} />
            </TabsContent>
          </Tabs>
        </section>
      </main>

      <div className="fixed right-4 bottom-20 z-30">
        <Link href={`/vehicles/${vehicle.id}/mileage`}>
          <button className="flex h-12 w-12 cursor-pointer items-center justify-center rounded-full bg-indigo-600 text-white shadow-lg shadow-indigo-600/30 transition-all hover:scale-105 hover:bg-indigo-500 active:scale-95">
            <Plus className="h-6 w-6" />
          </button>
        </Link>
      </div>

      <nav className="fixed bottom-4 left-1/2 z-30 flex h-14 w-[calc(100%-32px)] max-w-sm -translate-x-1/2 items-center justify-around rounded-2xl border border-zinc-200/60 bg-white/85 px-4 shadow-lg backdrop-blur-md dark:border-zinc-900/50 dark:bg-zinc-950/85">
        <Link
          href="/"
          className="flex cursor-pointer flex-col items-center gap-0.5 text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200"
        >
          <Sparkles className="h-4.5 w-4.5" />
          <span className="text-[9px] font-semibold">Workspace</span>
        </Link>

        <Link
          href={`/vehicles/${vehicle.id}/mileage`}
          className="flex cursor-pointer flex-col items-center gap-0.5 text-indigo-600 dark:text-indigo-400"
        >
          <PlusCircle className="h-4.5 w-4.5" />
          <span className="text-[9px] font-bold">Log Event</span>
        </Link>

        <div className="flex cursor-not-allowed flex-col items-center gap-0.5 text-zinc-300 dark:text-zinc-700">
          <Users className="h-4.5 w-4.5" />
          <span className="text-[9px] font-medium">Family</span>
        </div>
      </nav>
    </div>
  );
}
