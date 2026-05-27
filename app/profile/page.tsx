import { headers } from 'next/headers';
import { redirect } from 'next/navigation';
import { auth } from '@/lib/auth';
import { prisma } from '@/lib/prisma';
import { ProfileClient } from '@/components/profile-client';
import { Navigation } from '@/components/navigation';

export default async function ProfilePage() {
  const session = await auth.api.getSession({
    headers: await headers(),
  });

  if (!session) {
    redirect('/');
  }

  const userId = session.user.id;

  // Retrieve user-centric database statistics
  const [totalVehicles, mileageEntries] = await Promise.all([
    prisma.vehicle.count({
      where: {
        OR: [{ ownerId: userId }, { users: { some: { userId } } }],
      },
    }),
    prisma.mileageEntry.findMany({
      where: {
        userId: userId,
      },
      select: {
        fuelAmount: true,
        fuelLitres: true,
      },
    }),
  ]);

  const totalRefills = mileageEntries.filter(
    (m) => m.fuelLitres !== null || m.fuelAmount !== null
  ).length;
  const totalSpent = mileageEntries.reduce((sum, m) => sum + (m.fuelAmount || 0), 0);

  const mappedUser = {
    id: session.user.id,
    name: (session.user.name || session.user.email.split('@')[0]) as string,
    email: session.user.email,
    image: session.user.image,
  };

  const stats = {
    totalVehicles,
    totalRefills,
    totalSpent,
  };

  return (
    <>
      <ProfileClient user={mappedUser} stats={stats} />
      <Navigation />
    </>
  );
}
