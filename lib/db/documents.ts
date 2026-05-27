import { prisma } from '../prisma';

export async function updatePUC(data: {
  vehicleId: string;
  expiryDate: string;
  certificateNumber?: string;
}) {
  const [, puc] = await prisma.$transaction([
    prisma.pUC.deleteMany({ where: { vehicleId: data.vehicleId } }),
    prisma.pUC.create({
      data: {
        vehicleId: data.vehicleId,
        expiryDate: new Date(data.expiryDate),
        certificateNumber: data.certificateNumber,
      },
    }),
  ]);
  return puc;
}

export async function updateInsurance(data: {
  vehicleId: string;
  expiryDate: string;
  policyNumber?: string;
  provider?: string;
}) {
  const [, insurance] = await prisma.$transaction([
    prisma.insurance.deleteMany({ where: { vehicleId: data.vehicleId } }),
    prisma.insurance.create({
      data: {
        vehicleId: data.vehicleId,
        expiryDate: new Date(data.expiryDate),
        policyNumber: data.policyNumber,
        provider: data.provider,
      },
    }),
  ]);
  return insurance;
}
