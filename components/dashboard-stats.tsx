import { Car, AlertTriangle, TrendingUp } from 'lucide-react';

interface DashboardStatsProps {
  totalVehicles: number;
  activeAlerts: number;
  currentMonthExpenses: number;
}

export function DashboardStats({
  totalVehicles,
  activeAlerts,
  currentMonthExpenses,
}: DashboardStatsProps) {
  return (
    <section className="grid grid-cols-3 gap-3">
      <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
        <span className="text-[10px] font-bold tracking-wider text-zinc-400 uppercase">
          Vehicles
        </span>
        <div className="flex items-baseline gap-1">
          <span className="text-2xl font-black">{totalVehicles}</span>
          <Car className="h-3 w-3 text-zinc-400" />
        </div>
      </div>

      <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
        <span className="text-[10px] font-bold tracking-wider text-zinc-400 uppercase">Alerts</span>
        <div className="flex items-baseline gap-1">
          <span
            className={`text-2xl font-black ${
              activeAlerts > 0 ? 'text-amber-500' : 'text-emerald-500'
            }`}
          >
            {activeAlerts}
          </span>
          <AlertTriangle
            className={`h-3 w-3 ${
              activeAlerts > 0 ? 'animate-bounce text-amber-500' : 'text-zinc-400'
            }`}
          />
        </div>
      </div>

      <div className="shadow-3xs flex flex-col gap-1.5 rounded-2xl border border-zinc-200/60 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900">
        <span className="text-[10px] font-bold tracking-wider text-zinc-400 uppercase">
          Fuel (Mtd)
        </span>
        <div className="flex items-baseline gap-0.5 overflow-hidden text-ellipsis">
          <span className="text-[10px] font-bold text-zinc-400">₹</span>
          <span className="text-xl font-black tracking-tight">
            {currentMonthExpenses.toFixed(0)}
          </span>
          <TrendingUp className="ml-0.5 h-3 w-3 text-emerald-500" />
        </div>
      </div>
    </section>
  );
}
