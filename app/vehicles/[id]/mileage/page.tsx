import { AddMileageForm } from '@/components/add-mileage-form';

export default async function MileagePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <AddMileageForm vehicleId={id} />;
}
