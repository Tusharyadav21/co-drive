import { headers } from 'next/headers';
import { auth } from '@/lib/auth';
import { getVehiclesByUserId } from '@/lib/db/vehicles';
import { EXPIRY_WARNING_DAYS, getExpiryStatus } from '@/lib/expiry';
import { DashboardLogin } from '@/components/dashboard-login';

import { DashboardStats } from '@/components/dashboard-stats';
import { DashboardAlertsBanner } from '@/components/dashboard-alerts-banner';
import { DashboardFleet } from '@/components/dashboard-fleet';
import { Header } from '@/components/header';
import { Navigation } from '@/components/navigation';

interface Vehicle {
  id: string;
  name: string;
  make: string;
  model: string;
  year: number;
  licensePlate: string;
  users: Array<{
    id: string;
    user?: { name?: string | null; email?: string | null } | null;
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

export default async function Home() {
  const session = await auth.api.getSession({
    headers: await headers(),
  });

  if (!session) {
    return <DashboardLogin />;
  }

  const userId = session.user.id;
  const username = session.user.name || session.user.email || '';
  const avatarInitials = username.slice(0, 2).toUpperCase();

  // Fetch vehicles using DB layer
  const vehiclesData = await getVehiclesByUserId(userId);

  // Map database dates and structures to match expected Vehicle frontend interfaces
  const vehicles: Vehicle[] = vehiclesData.map((v) => ({
    id: v.id,
    name: v.name,
    make: v.make,
    model: v.model,
    year: v.year,
    licensePlate: v.licensePlate,
    users: v.users.map((u) => ({
      id: u.id,
      user: u.user ? { name: u.user.name, email: u.user.email } : null,
    })),
    mileageEntries: v.mileageEntries.map((m) => ({
      mileage: m.mileage,
      date: m.date.toISOString(),
      fuelAmount: m.fuelAmount,
    })),
    pucRecords: v.pucRecords.map((p) => ({
      expiryDate: p.expiryDate.toISOString(),
    })),
    insuranceRecords: v.insuranceRecords.map((i) => ({
      expiryDate: i.expiryDate.toISOString(),
    })),
  }));

  // Compute stats for premium look
  const totalVehicles = vehicles.length;

  // Calculate total monthly fuel expenses
  const currentMonthExpenses = vehicles.reduce((total, vehicle) => {
    const monthLogs = vehicle.mileageEntries.filter((entry) => {
      const entryDate = new Date(entry.date);
      const today = new Date();
      return (
        entryDate.getMonth() === today.getMonth() && entryDate.getFullYear() === today.getFullYear()
      );
    });
    return total + monthLogs.reduce((subTotal, entry) => subTotal + (entry.fuelAmount || 0), 0);
  }, 0);

  // Check how many vehicles have an active alert (expired or expiring document)
  const isAlerting = (status: string) => status === 'expired' || status === 'expiring';

  const activeAlerts = vehicles.reduce((count, vehicle) => {
    const hasAlert =
      isAlerting(getExpiryStatus(vehicle.pucRecords[0]?.expiryDate, EXPIRY_WARNING_DAYS.puc)) ||
      isAlerting(
        getExpiryStatus(vehicle.insuranceRecords[0]?.expiryDate, EXPIRY_WARNING_DAYS.insurance)
      );

    return count + (hasAlert ? 1 : 0);
  }, 0);

  return (
    <div className="min-h-screen bg-zinc-50 pb-24 font-sans text-zinc-900 antialiased dark:bg-zinc-950 dark:text-zinc-50">
      {userId && <Header username={username} avatarInitials={avatarInitials} />}
      <main className="container mx-auto flex max-w-lg flex-col gap-6 px-4 py-6">
        <DashboardStats
          totalVehicles={totalVehicles}
          activeAlerts={activeAlerts}
          currentMonthExpenses={currentMonthExpenses}
        />

        <DashboardAlertsBanner activeAlerts={activeAlerts} />

        <DashboardFleet vehicles={vehicles} />
        {userId && <Navigation />}
      </main>
    </div>
  );
}
