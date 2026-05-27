export type ExpiryStatus = 'expired' | 'expiring' | 'active' | 'none';

const MS_PER_DAY = 1000 * 60 * 60 * 24;

/** Number of days before expiry at which each document type is "expiring soon". */
export const EXPIRY_WARNING_DAYS = {
  puc: 7,
  insurance: 15,
} as const;

/** Whole days from `now` until `expiryDate` (negative if already expired). */
export function getDaysUntilExpiry(expiryDate: string | Date, now: Date = new Date()): number {
  const expiry = new Date(expiryDate);
  return Math.ceil((expiry.getTime() - now.getTime()) / MS_PER_DAY);
}

/** Classifies an expiry date into a status, using the given warning window. */
export function getExpiryStatus(
  expiryDate: string | Date | null | undefined,
  warningDays: number,
  now: Date = new Date()
): ExpiryStatus {
  if (!expiryDate) return 'none';
  const days = getDaysUntilExpiry(expiryDate, now);
  if (days < 0) return 'expired';
  if (days <= warningDays) return 'expiring';
  return 'active';
}
