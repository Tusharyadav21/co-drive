import { PUCForm } from '@/components/puc-form';

export default async function PUCPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <PUCForm vehicleId={id} />;
}
