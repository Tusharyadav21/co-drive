import { notFound } from 'next/navigation';
import { headers } from 'next/headers';
import { VehicleDetail } from '@/components/vehicle-detail';
import { getVehicleById } from '@/lib/db/vehicles';
import { auth } from '@/lib/auth';

export default async function VehiclePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const session = await auth.api.getSession({
    headers: await headers(),
  });

  if (!session) {
    return notFound();
  }

  const userId = session.user.id;

  const vehicleData = await getVehicleById(id, userId);

  if (!vehicleData) {
    return notFound();
  }

  return <VehicleDetail vehicle={vehicleData} currentUserId={userId} />;
}
