'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Trash2, UserPlus } from 'lucide-react';
import type { VehicleWithDetails } from '@/types';
import { shareVehicleAction, removeShareAction } from '@/lib/actions/vehicle';

export function SharesTab({
  vehicle,
  currentUserId,
}: {
  vehicle: VehicleWithDetails;
  currentUserId: string;
}) {
  const [shareUsername, setShareUsername] = useState('');
  const [sharingError, setSharingError] = useState('');
  const [sharingSuccess, setSharingSuccess] = useState('');
  const [sharingLoading, setSharingLoading] = useState(false);

  const handleShareSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!shareUsername.trim()) return;

    setSharingLoading(true);
    setSharingError('');
    setSharingSuccess('');

    try {
      const response = await shareVehicleAction(
        {
          vehicleId: vehicle.id,
          username: shareUsername.trim(),
        },
        currentUserId
      );

      if (response.error) {
        throw new Error(response.error);
      }

      setSharingSuccess(`Shared successfully with @${shareUsername}`);
      setShareUsername('');
    } catch (err) {
      setSharingError(err instanceof Error ? err.message : 'Failed to share vehicle');
    } finally {
      setSharingLoading(false);
    }
  };

  const handleRemoveShare = async (vehicleUserId: string) => {
    if (!confirm('Are you sure you want to remove access for this member?')) return;

    try {
      const response = await removeShareAction({ vehicleUserId }, currentUserId);
      if (response.error) {
        throw new Error(response.error);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to remove access');
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <span className="text-xs font-bold tracking-wider text-zinc-400 uppercase">
          Active Board
        </span>
        <span className="text-[10px] font-semibold text-zinc-400">
          {vehicle.users.length} members
        </span>
      </div>

      <div className="shadow-3xs flex flex-col gap-4 rounded-3xl border border-zinc-200/60 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
        <div className="flex flex-col gap-3">
          {vehicle.users.map((vu) => {
            const memberUsername = vu.user?.name || vu.user?.email || vu.userId;
            // Hacky check for owner if schema doesn't clearly provide it on Vehicle type
            // (Wait, Vehicle type has ownerId? Not in VehicleWithDetails unless added,
            // but we can assume role==="owner")
            const isOwner = vu.role === 'owner' || vehicle.users.length === 1;

            return (
              <div
                key={vu.id}
                className="flex items-center justify-between rounded-2xl border border-zinc-200/40 bg-zinc-50 p-3 dark:border-zinc-900/60 dark:bg-zinc-950/40"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-indigo-500/10 text-xs font-bold text-indigo-600 dark:text-indigo-400">
                    {memberUsername.slice(0, 2).toUpperCase()}
                  </div>
                  <div>
                    <span className="block text-xs font-bold">@{memberUsername}</span>
                    <span className="text-[9px] font-semibold text-zinc-400 capitalize">
                      {vu.role}
                    </span>
                  </div>
                </div>

                {!isOwner && (
                  <button
                    onClick={() => handleRemoveShare(vu.id)}
                    className="cursor-pointer rounded-xl border border-transparent p-2 text-zinc-400 shadow-2xs transition-colors hover:border-red-500/20 hover:bg-red-500/5 hover:text-red-500 dark:hover:bg-red-500/10"
                    title="Revoke Access"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                )}
              </div>
            );
          })}
        </div>

        <div className="my-1 h-px bg-zinc-200/60 dark:bg-zinc-800" />

        <form onSubmit={handleShareSubmit} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1.5">
            <span className="text-[10px] font-bold tracking-wider text-zinc-400 uppercase">
              Invite Family / Friends
            </span>
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <span className="absolute top-3 left-3 text-xs text-zinc-400">@</span>
                <Input
                  value={shareUsername}
                  onChange={(e) => setShareUsername(e.target.value)}
                  placeholder="e.g. brother_username"
                  disabled={sharingLoading}
                  className="h-10 rounded-xl border-zinc-200 bg-zinc-50 pr-3 pl-7 text-xs font-semibold dark:border-zinc-800 dark:bg-zinc-950"
                />
              </div>
              <Button
                type="submit"
                disabled={sharingLoading || !shareUsername.trim()}
                className="h-10 cursor-pointer rounded-xl bg-indigo-600 px-4 text-xs font-semibold text-white shadow-md shadow-indigo-600/10 hover:bg-indigo-500 disabled:opacity-50"
              >
                <UserPlus className="mr-1 h-3.5 w-3.5" /> Invite
              </Button>
            </div>
          </div>

          {sharingError && (
            <p className="animate-pulse text-[10px] font-medium text-red-500">{sharingError}</p>
          )}
          {sharingSuccess && (
            <p className="animate-pulse text-[10px] font-medium text-emerald-500">
              {sharingSuccess}
            </p>
          )}
        </form>
      </div>
    </div>
  );
}
