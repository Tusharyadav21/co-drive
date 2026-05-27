import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { FileText, ShieldCheck } from 'lucide-react';
import Link from 'next/link';
import { getDaysUntilExpiry } from '@/lib/expiry';
import type { VehicleWithDetails } from '@/types';

const getStatusBadge = (days: number) => {
  if (days < 0) {
    return (
      <Badge className="rounded-full border-none bg-red-500/10 px-2 py-0.5 text-[10px] font-bold text-red-500 hover:bg-red-500/15">
        Expired
      </Badge>
    );
  } else if (days <= 7) {
    return (
      <Badge className="rounded-full border-none bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold text-amber-500 hover:bg-amber-500/15">
        Expiring Soon
      </Badge>
    );
  } else if (days <= 15) {
    return (
      <Badge className="rounded-full border-none bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold text-amber-500 hover:bg-amber-500/15">
        2 Weeks Left
      </Badge>
    );
  }
  return (
    <Badge className="rounded-full border-none bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold text-emerald-500 hover:bg-emerald-500/15">
      Active
    </Badge>
  );
};

export function DocumentsTab({ vehicle }: { vehicle: VehicleWithDetails }) {
  const latestPuc = vehicle.pucRecords[0];
  const latestInsurance = vehicle.insuranceRecords[0];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <span className="text-xs font-bold tracking-wider text-zinc-400 uppercase">
          Mandatory Certificates
        </span>
        <span className="text-[10px] font-semibold text-zinc-400">2 active</span>
      </div>

      <div className="shadow-3xs flex flex-col gap-4 rounded-3xl border border-zinc-200/60 bg-white p-4.5 dark:border-zinc-800 dark:bg-zinc-900">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-violet-500/10 text-violet-600 dark:text-violet-400">
              <FileText className="h-4.5 w-4.5" />
            </div>
            <div>
              <h4 className="text-xs font-bold">PUC Certificate</h4>
              <p className="text-[9px] text-zinc-400">Pollution Under Control</p>
            </div>
          </div>
          {latestPuc &&
            getStatusBadge(getDaysUntilExpiry(new Date(latestPuc.expiryDate).toISOString()))}
        </div>

        {latestPuc ? (
          <div className="flex flex-col gap-2 rounded-2xl border border-zinc-200/40 bg-zinc-50 p-3 dark:border-zinc-900/60 dark:bg-zinc-950/40">
            <div className="flex items-center justify-between text-[10px]">
              <span className="font-medium text-zinc-400">Expiry Date</span>
              <span className="font-bold">
                {new Date(latestPuc.expiryDate).toLocaleDateString('en-IN', {
                  day: 'numeric',
                  month: 'short',
                  year: 'numeric',
                })}
              </span>
            </div>
            {latestPuc.certificateNumber && (
              <div className="flex items-center justify-between text-[10px]">
                <span className="font-medium text-zinc-400">Certificate No.</span>
                <span className="font-mono font-bold">{latestPuc.certificateNumber}</span>
              </div>
            )}
          </div>
        ) : (
          <p className="rounded-2xl border border-dashed border-zinc-200 bg-zinc-50 py-4 text-center text-[10px] text-zinc-400 italic dark:border-zinc-800 dark:bg-zinc-950/40">
            No PUC records verified yet.
          </p>
        )}

        <Link href={`/vehicles/${vehicle.id}/puc`} className="block">
          <Button className="h-11 w-full cursor-pointer rounded-2xl bg-zinc-900 text-xs font-semibold text-white shadow-xs hover:bg-zinc-800 dark:bg-zinc-800 dark:hover:bg-zinc-700">
            {latestPuc ? 'Update PUC File' : 'Add PUC Records'}
          </Button>
        </Link>
      </div>

      <div className="shadow-3xs flex flex-col gap-4 rounded-3xl border border-zinc-200/60 bg-white p-4.5 dark:border-zinc-800 dark:bg-zinc-900">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <ShieldCheck className="h-4.5 w-4.5" />
            </div>
            <div>
              <h4 className="text-xs font-bold">Insurance Policy</h4>
              <p className="text-[9px] text-zinc-400">Comprehensive Vehicle Protection</p>
            </div>
          </div>
          {latestInsurance &&
            getStatusBadge(getDaysUntilExpiry(new Date(latestInsurance.expiryDate).toISOString()))}
        </div>

        {latestInsurance ? (
          <div className="flex flex-col gap-2 rounded-2xl border border-zinc-200/40 bg-zinc-50 p-3 dark:border-zinc-900/60 dark:bg-zinc-950/40">
            <div className="flex items-center justify-between text-[10px]">
              <span className="font-medium text-zinc-400">Expiry Date</span>
              <span className="font-bold">
                {new Date(latestInsurance.expiryDate).toLocaleDateString('en-IN', {
                  day: 'numeric',
                  month: 'short',
                  year: 'numeric',
                })}
              </span>
            </div>
            {latestInsurance.provider && (
              <div className="flex items-center justify-between text-[10px]">
                <span className="font-medium text-zinc-400">Provider</span>
                <span className="font-bold">{latestInsurance.provider}</span>
              </div>
            )}
            {latestInsurance.policyNumber && (
              <div className="flex items-center justify-between text-[10px]">
                <span className="font-medium text-zinc-400">Policy No.</span>
                <span className="font-mono font-bold">{latestInsurance.policyNumber}</span>
              </div>
            )}
          </div>
        ) : (
          <p className="rounded-2xl border border-dashed border-zinc-200 bg-zinc-50 py-4 text-center text-[10px] text-zinc-400 italic dark:border-zinc-800 dark:bg-zinc-950/40">
            No insurance policy loaded.
          </p>
        )}

        <Link href={`/vehicles/${vehicle.id}/insurance`} className="block">
          <Button className="h-11 w-full cursor-pointer rounded-2xl bg-zinc-900 text-xs font-semibold text-white shadow-xs hover:bg-zinc-800 dark:bg-zinc-800 dark:hover:bg-zinc-700">
            {latestInsurance ? 'Update Insurance File' : 'Add Insurance Records'}
          </Button>
        </Link>
      </div>
    </div>
  );
}
