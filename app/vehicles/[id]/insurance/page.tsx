import { InsuranceForm } from '@/components/insurance-form';

export default async function InsurancePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <InsuranceForm vehicleId={id} />;
}
