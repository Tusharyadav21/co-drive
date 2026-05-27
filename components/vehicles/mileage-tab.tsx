'use client';

import { History, TrendingUp, Fuel } from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import type { VehicleWithDetails } from '@/types';

export function MileageTab({ vehicle }: { vehicle: VehicleWithDetails }) {
  // Calculate individual efficiencies and overall average in a chronological (older-to-newer) reference
  const entriesWithEfficiency = vehicle.mileageEntries.map((entry, index) => {
    // Since vehicle.mileageEntries is sorted by date desc, the historically older entry is at index + 1
    const previousEntry = vehicle.mileageEntries[index + 1];
    let efficiency: number | null = null;
    let distance: number | null = null;

    if (previousEntry && entry.mileage > previousEntry.mileage && entry.fuelLitres) {
      distance = entry.mileage - previousEntry.mileage;
      efficiency = distance / entry.fuelLitres;
    }

    return {
      ...entry,
      distance,
      efficiency,
    };
  });

  const validEfficiencies = entriesWithEfficiency.filter((e) => e.efficiency !== null) as Array<
    (typeof vehicle.mileageEntries)[0] & { distance: number; efficiency: number }
  >;

  const averageEfficiency =
    validEfficiencies.length > 0
      ? validEfficiencies.reduce((sum, e) => sum + e.efficiency, 0) / validEfficiencies.length
      : null;

  const totalTrackedDistance = validEfficiencies.reduce((sum, e) => sum + e.distance, 0);

  const chartData = [...vehicle.mileageEntries].reverse().map((entry) => ({
    date: new Date(entry.date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    }),
    Odometer: entry.mileage,
    Expense: entry.fuelAmount || 0,
  }));

  return (
    <div className="flex flex-col gap-4">
      {averageEfficiency !== null && (
        <div className="grid grid-cols-2 gap-3">
          <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
            <span className="flex items-center gap-1 text-[10px] font-bold tracking-wider text-zinc-400 uppercase">
              <Fuel className="h-3.5 w-3.5 text-emerald-500" />
              Avg Mileage
            </span>
            <div className="flex items-baseline gap-1">
              <span className="text-xl font-black text-emerald-600 dark:text-emerald-400">
                {averageEfficiency.toFixed(2)}
              </span>
              <span className="text-[10px] font-semibold text-zinc-400">km/l</span>
            </div>
          </div>

          <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
            <span className="flex items-center gap-1 text-[10px] font-bold tracking-wider text-zinc-400 uppercase">
              <TrendingUp className="h-3.5 w-3.5 text-indigo-500" />
              Tracked Run
            </span>
            <div className="flex items-baseline gap-1">
              <span className="text-xl font-black text-zinc-800 dark:text-zinc-200">
                {totalTrackedDistance.toLocaleString('en-IN')}
              </span>
              <span className="text-[10px] font-semibold text-zinc-400">km</span>
            </div>
          </div>
        </div>
      )}

      {chartData.length > 1 && (
        <div className="shadow-3xs rounded-3xl border border-zinc-200/60 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
          <div className="mb-4 flex items-center gap-1.5">
            <TrendingUp className="h-4 w-4 text-indigo-500" />
            <span className="text-xs font-bold text-zinc-500 dark:text-zinc-400">
              Odometer Logs (km)
            </span>
          </div>
          <div className="h-[180px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData} margin={{ top: 5, right: 5, left: -25, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorOdometer" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#4f46e5" stopOpacity={0.2} />
                    <stop offset="95%" stopColor="#4f46e5" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  strokeDasharray="3 3"
                  vertical={false}
                  stroke="rgba(200,200,200,0.15)"
                />
                <XAxis dataKey="date" fontSize={9} tickLine={false} axisLine={false} />
                <YAxis fontSize={9} tickLine={false} axisLine={false} domain={['auto', 'auto']} />
                <Tooltip
                  contentStyle={{
                    background: 'rgba(255, 255, 255, 0.95)',
                    borderRadius: '16px',
                    border: '1px solid rgba(200,200,200,0.4)',
                    fontSize: '10px',
                    color: '#18181b',
                    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.05)',
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="Odometer"
                  stroke="#4f46e5"
                  strokeWidth={2.5}
                  fillOpacity={1}
                  fill="url(#colorOdometer)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold tracking-wider text-zinc-400 uppercase">
            Activity Feed
          </span>
          <span className="text-[10px] font-semibold text-zinc-400">
            {entriesWithEfficiency.length} entries
          </span>
        </div>

        {entriesWithEfficiency.length === 0 ? (
          <div className="shadow-3xs flex flex-col items-center gap-3 rounded-3xl border border-zinc-200/60 bg-white p-8 text-center dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-zinc-50 text-zinc-400 dark:bg-zinc-950">
              <History className="h-5 w-5" />
            </div>
            <div>
              <h4 className="text-xs font-bold">No logs listed yet</h4>
              <p className="mt-1 max-w-[180px] text-[10px] leading-normal text-zinc-400">
                Add mileage logs, refills, or fuel prices using the floating add button below.
              </p>
            </div>
          </div>
        ) : (
          <div className="relative flex flex-col gap-3 before:absolute before:top-4 before:bottom-4 before:left-5 before:w-0.5 before:bg-zinc-200/60 dark:before:bg-zinc-800/60">
            {entriesWithEfficiency.map((entry) => (
              <div key={entry.id} className="relative flex flex-col gap-1.5 pl-10">
                <div className="absolute top-1.5 left-3.5 z-10 flex h-3 w-3 items-center justify-center rounded-full border-2 border-white bg-indigo-600 ring-4 ring-indigo-500/10 dark:border-zinc-900" />

                <div className="shadow-3xs flex flex-col gap-2 rounded-2xl border border-zinc-200/60 bg-white p-3.5 dark:border-zinc-800 dark:bg-zinc-900">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-extrabold text-zinc-800 dark:text-zinc-200">
                      {entry.mileage.toLocaleString('en-IN')}{' '}
                      <span className="text-[10px] font-normal text-zinc-400">km</span>
                    </span>
                    <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-[9px] font-bold text-zinc-400 dark:bg-zinc-800">
                      {new Date(entry.date).toLocaleDateString('en-IN', {
                        day: 'numeric',
                        month: 'short',
                        year: 'numeric',
                      })}
                    </span>
                  </div>

                  {(entry.fuelLitres || entry.fuelAmount || entry.efficiency) && (
                    <div className="flex items-center gap-3 rounded-xl border border-emerald-500/10 bg-emerald-500/5 px-2.5 py-1.5 dark:bg-emerald-500/5">
                      <Fuel className="h-3.5 w-3.5 text-emerald-500" />
                      <div className="flex flex-wrap items-center gap-1.5">
                        {entry.fuelLitres && (
                          <span className="text-[10px] font-bold text-emerald-600 dark:text-emerald-400">
                            {entry.fuelLitres} Litres
                          </span>
                        )}
                        {entry.fuelLitres && entry.fuelAmount && (
                          <span className="text-[10px] text-zinc-300 dark:text-zinc-800">•</span>
                        )}
                        {entry.fuelAmount && (
                          <span className="text-[10px] font-bold text-emerald-600 dark:text-emerald-400">
                            ₹{entry.fuelAmount} Spent
                          </span>
                        )}
                        {entry.efficiency && (
                          <>
                            {(entry.fuelLitres || entry.fuelAmount) && (
                              <span className="text-[10px] text-zinc-300 dark:text-zinc-800">
                                •
                              </span>
                            )}
                            <span className="text-[10px] font-extrabold text-emerald-600 dark:text-emerald-400">
                              {entry.efficiency.toFixed(2)} km/l
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                  )}

                  {entry.notes && (
                    <p className="rounded-xl bg-zinc-50 p-2 text-[10px] leading-tight text-zinc-500 italic dark:bg-zinc-950 dark:text-zinc-400">
                      "{entry.notes}"
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
