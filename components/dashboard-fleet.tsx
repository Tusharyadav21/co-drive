import Link from 'next/link';
import { Car } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { VehicleCard } from '@/components/vehicle-card';

interface Vehicle {
  id: string;
  name: string;
  make: string;
  model: string;
  year: number;
  licensePlate: string;
  users: Array<{
    id: string;
    user?: { name?: string | null } | null;
  }>;
  mileageEntries: Array<{
    mileage: number;
    date: string;
    fuelAmount?: number | null;
  }>;
  pucRecords: Array<{
    expiryDate: string;
  }>;
  insuranceRecords: Array<{
    expiryDate: string;
  }>;
}

interface DashboardFleetProps {
  vehicles: Vehicle[];
}

export function DashboardFleet({ vehicles }: DashboardFleetProps) {
  const totalVehicles = vehicles.length;

  return (
    <section className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-bold tracking-wider text-zinc-400 uppercase">Your Fleet</h3>
        <span className="rounded-full bg-zinc-200/50 px-2 py-0.5 text-[10px] font-bold text-zinc-400 dark:bg-zinc-900">
          {totalVehicles} active
        </span>
      </div>

      {totalVehicles === 0 ? (
        <div className="shadow-3xs flex flex-col items-center gap-4 rounded-3xl border border-zinc-200/60 bg-white p-8 text-center dark:border-zinc-800 dark:bg-zinc-900">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-zinc-50 text-zinc-400 dark:bg-zinc-950 dark:text-zinc-500">
            <Car className="h-8 w-8" />
          </div>
          <div>
            <h4 className="mb-1 text-sm font-bold">No Vehicles Loaded</h4>
            <p className="mx-auto max-w-[220px] text-[11px] leading-normal text-zinc-400">
              Add your first car or bike to start tracking logs and document expiry with your
              family.
            </p>
          </div>
          <Link href="/vehicles/new">
            <Button className="h-10 cursor-pointer rounded-2xl bg-indigo-600 px-6 text-xs font-semibold text-white shadow-md shadow-indigo-600/10">
              Add Your First Vehicle
            </Button>
          </Link>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {vehicles.map((vehicle) => (
            <VehicleCard
              key={vehicle.id}
              id={vehicle.id}
              name={vehicle.name}
              make={vehicle.make}
              model={vehicle.model}
              year={vehicle.year}
              licensePlate={vehicle.licensePlate}
              latestMileage={vehicle.mileageEntries[0]}
              puc={vehicle.pucRecords[0]}
              insurance={vehicle.insuranceRecords[0]}
              users={vehicle.users}
            />
          ))}
        </div>
      )}
    </section>
  );
}
