import { Car, Bike } from 'lucide-react';
import type { VehicleWithDetails } from '@/types';

export function VehicleHero({ vehicle }: { vehicle: VehicleWithDetails }) {
  const isTwoWheeler =
    vehicle.name.toLowerCase().includes('bike') ||
    vehicle.name.toLowerCase().includes('scoot') ||
    vehicle.make.toLowerCase().includes('honda act') ||
    vehicle.make.toLowerCase().includes('royal') ||
    vehicle.make.toLowerCase().includes('yamaha');

  return (
    <section className="relative overflow-hidden rounded-3xl bg-linear-to-tr from-zinc-900 via-zinc-800 to-zinc-900 p-5 text-white shadow-xl dark:border dark:border-zinc-800">
      <div className="absolute top-0 right-0 h-32 w-32 rounded-full bg-indigo-500/10 blur-3xl" />
      <div className="relative z-10 flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/10 text-white backdrop-blur-md">
            {isTwoWheeler ? (
              <Bike className="h-6 w-6 text-indigo-300" />
            ) : (
              <Car className="h-6 w-6 text-indigo-300" />
            )}
          </div>
          <div>
            <span className="text-[9px] leading-none font-bold tracking-wider text-zinc-400 uppercase">
              Vehicle Vault
            </span>
            <h2 className="mt-0.5 text-xl leading-none font-bold tracking-tight">{vehicle.name}</h2>
            <span className="mt-2 inline-flex items-center gap-1.5 rounded-md border border-zinc-300 bg-zinc-100 px-2 py-0.5 shadow-2xs dark:border-zinc-700 dark:bg-zinc-800/90">
              <div className="h-1.5 w-1.5 rounded-full bg-blue-500" />
              <span className="font-mono text-[10px] leading-none font-black tracking-wider text-zinc-800 uppercase dark:text-zinc-200">
                {vehicle.licensePlate}
              </span>
            </span>
          </div>
        </div>

        <div className="flex flex-col items-end gap-1">
          <span className="text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
            Fleet Access
          </span>
          <div className="flex -space-x-1">
            {vehicle.users.map((vu) => (
              <div
                key={vu.id}
                className="flex h-6 w-6 items-center justify-center rounded-full border-2 border-zinc-900 bg-indigo-600 text-[8px] font-bold text-white shadow-2xs"
                title={vu.user?.name || vu.user?.email || vu.userId}
              >
                {(vu.user?.name || vu.user?.email || 'U').slice(0, 2).toUpperCase()}
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
