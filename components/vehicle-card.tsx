'use client';

import { Card, CardContent } from '@/components/ui/card';
import Link from 'next/link';
import { Car, Bike, ShieldCheck } from 'lucide-react';
import { EXPIRY_WARNING_DAYS, getExpiryStatus } from '@/lib/expiry';

interface SharedMember {
  id: string;
  user?: { name?: string | null; email?: string | null } | null;
}

interface VehicleCardProps {
  id: string;
  name: string;
  make: string;
  model: string;
  year: number;
  licensePlate: string;
  latestMileage?: { mileage: number; date: string } | null;
  puc?: { expiryDate: string } | null;
  insurance?: { expiryDate: string } | null;
  users?: SharedMember[];
}

export function VehicleCard({
  id,
  name,
  make,
  model,
  year,
  licensePlate,
  latestMileage,
  puc,
  insurance,
  users = [],
}: VehicleCardProps) {
  const pucStatus = getExpiryStatus(puc?.expiryDate, EXPIRY_WARNING_DAYS.puc);
  const insuranceStatus = getExpiryStatus(insurance?.expiryDate, EXPIRY_WARNING_DAYS.insurance);

  const visibleMembers = users.slice(0, 3);
  const extraMembers = users.length - visibleMembers.length;
  const memberInitials = (member: SharedMember) =>
    (member.user?.name || member.user?.email || 'U').slice(0, 2).toUpperCase();

  // Detect vehicle type based on model/name for beautiful icon representation
  const isTwoWheeler =
    name.toLowerCase().includes('bike') ||
    name.toLowerCase().includes('scoot') ||
    make.toLowerCase().includes('honda act') ||
    make.toLowerCase().includes('royal') ||
    make.toLowerCase().includes('yamaha');

  // Format mileage for readability
  const formattedMileage = latestMileage
    ? new Intl.NumberFormat('en-IN').format(latestMileage.mileage)
    : null;

  return (
    <Link
      href={`/vehicles/${id}`}
      className="block transition-transform duration-200 select-none active:scale-[0.98]"
    >
      <Card className="overflow-hidden rounded-3xl border border-zinc-200/60 bg-white shadow-xs transition-colors hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-900 dark:hover:border-zinc-700">
        {/* Top Gradient Banner representing the vehicle visually */}
        <div className="relative flex h-16 items-center justify-between overflow-hidden border-b border-zinc-200/20 bg-linear-to-r from-zinc-900 via-zinc-800 to-zinc-900 p-4 dark:border-zinc-800/40 dark:from-zinc-950 dark:to-zinc-900">
          <div className="pointer-events-none absolute top-0 right-0 h-24 w-24 rounded-full bg-indigo-500/10 blur-2xl dark:bg-indigo-600/5" />

          <div className="relative z-10 flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-white/10 text-white backdrop-blur-md">
              {isTwoWheeler ? (
                <Bike className="h-5 w-5 text-indigo-300" />
              ) : (
                <Car className="h-5 w-5 text-indigo-300" />
              )}
            </div>
            <div>
              <h3 className="text-sm font-bold tracking-tight text-white">{name}</h3>
              <p className="text-[10px] font-medium text-zinc-400">
                {year} • {make} {model}
              </p>
            </div>
          </div>

          {/* Premium license plate badge */}
          <div className="relative z-10 flex items-center gap-1.5 rounded-md border border-zinc-300 bg-zinc-100 px-2.5 py-1 shadow-2xs dark:border-zinc-700 dark:bg-zinc-800/90">
            <div className="h-1.5 w-1.5 rounded-full bg-blue-500" />
            <span className="font-mono text-[10px] font-black tracking-wider text-zinc-800 uppercase dark:text-zinc-200">
              {licensePlate}
            </span>
          </div>
        </div>

        {/* Card Content - Dynamic health overview */}
        <CardContent className="flex flex-col gap-4 p-4">
          {/* Mileage / Recent Log Details */}
          <div className="flex items-center justify-between">
            {formattedMileage ? (
              <div className="flex flex-col">
                <span className="text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
                  Odometer
                </span>
                <span className="text-sm font-extrabold text-zinc-800 dark:text-zinc-200">
                  {formattedMileage} <span className="text-xs font-normal text-zinc-400">km</span>
                </span>
              </div>
            ) : (
              <div className="flex flex-col">
                <span className="text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
                  Odometer
                </span>
                <span className="text-xs font-semibold text-zinc-400 italic">No entry yet</span>
              </div>
            )}

            {/* Overlapping Family/Shared User Avatars */}
            {users.length > 0 && (
              <div className="flex items-center">
                <span className="mr-2 text-[9px] font-bold tracking-wider text-zinc-400 uppercase">
                  Shared
                </span>
                <div className="flex -space-x-1.5">
                  {visibleMembers.map((member) => (
                    <div
                      key={member.id}
                      title={member.user?.name || member.user?.email || 'Member'}
                      className="flex h-5 w-5 items-center justify-center rounded-full border border-white bg-indigo-500 text-[8px] font-bold text-white shadow-2xs dark:border-zinc-900"
                    >
                      {memberInitials(member)}
                    </div>
                  ))}
                  {extraMembers > 0 && (
                    <div className="flex h-5 w-5 items-center justify-center rounded-full border border-white bg-zinc-200 text-[7px] font-bold text-zinc-600 shadow-2xs dark:border-zinc-900 dark:bg-zinc-800 dark:text-zinc-400">
                      +{extraMembers}
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Status Indicators (PUC & Insurance) */}
          <div className="flex w-full gap-2">
            {/* PUC Status Block */}
            <div className="flex flex-1 items-center justify-between rounded-2xl border border-zinc-200/40 bg-zinc-50 p-2 dark:border-zinc-900/60 dark:bg-zinc-950/40">
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="h-3.5 w-3.5 text-zinc-400" />
                <span className="text-[10px] font-bold text-zinc-500 dark:text-zinc-400">PUC</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span
                  className={`h-1.5 w-1.5 rounded-full ${
                    pucStatus === 'active'
                      ? 'bg-emerald-500 shadow-xs shadow-emerald-500'
                      : pucStatus === 'expiring'
                        ? 'animate-pulse bg-amber-500'
                        : 'animate-ping bg-red-500'
                  }`}
                />
                <span
                  className={`text-[10px] font-extrabold capitalize ${
                    pucStatus === 'active'
                      ? 'text-emerald-600 dark:text-emerald-400'
                      : pucStatus === 'expiring'
                        ? 'text-amber-600 dark:text-amber-400'
                        : 'text-red-500'
                  }`}
                >
                  {pucStatus === 'none' ? 'Missing' : pucStatus}
                </span>
              </div>
            </div>

            {/* Insurance Status Block */}
            <div className="flex flex-1 items-center justify-between rounded-2xl border border-zinc-200/40 bg-zinc-50 p-2 dark:border-zinc-900/60 dark:bg-zinc-950/40">
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="h-3.5 w-3.5 text-zinc-400" />
                <span className="text-[10px] font-bold text-zinc-500 dark:text-zinc-400">
                  Insurance
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <span
                  className={`h-1.5 w-1.5 rounded-full ${
                    insuranceStatus === 'active'
                      ? 'bg-emerald-500 shadow-xs shadow-emerald-500'
                      : insuranceStatus === 'expiring'
                        ? 'animate-pulse bg-amber-500'
                        : 'animate-ping bg-red-500'
                  }`}
                />
                <span
                  className={`text-[10px] font-extrabold capitalize ${
                    insuranceStatus === 'active'
                      ? 'text-emerald-600 dark:text-emerald-400'
                      : insuranceStatus === 'expiring'
                        ? 'text-amber-600 dark:text-amber-400'
                        : 'text-red-500'
                  }`}
                >
                  {insuranceStatus === 'none' ? 'Missing' : insuranceStatus}
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
