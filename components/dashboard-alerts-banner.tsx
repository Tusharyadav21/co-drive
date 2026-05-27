import { AlertTriangle } from 'lucide-react';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';

interface DashboardAlertsBannerProps {
  activeAlerts: number;
}

export function DashboardAlertsBanner({ activeAlerts }: DashboardAlertsBannerProps) {
  if (activeAlerts === 0) return null;

  return (
    <Alert className="shadow-3xs flex items-start gap-3 rounded-2xl border-amber-200/60 bg-amber-50 p-4 text-amber-800 dark:border-amber-900/30 dark:bg-amber-950/20 dark:text-amber-300 [&>svg]:text-amber-600 dark:[&>svg]:text-amber-400">
      <AlertTriangle className="mt-0.5 h-4.5 w-4.5 shrink-0" />
      <div className="flex flex-1 flex-col gap-0.5">
        <AlertTitle className="text-xs leading-tight font-bold">Renewal Alerts Pending</AlertTitle>
        <AlertDescription className="text-[10px] leading-tight text-amber-700 dark:text-amber-400">
          Some vehicles have expired or expiring PUC / Insurance policies. Tap vehicle cards to view
          or update documents.
        </AlertDescription>
      </div>
    </Alert>
  );
}
