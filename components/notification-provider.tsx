'use client';

import { useEffect } from 'react';
import { initializeNotifications } from '@/lib/notifications';

export function NotificationProvider() {
  useEffect(() => {
    const initNotifications = async () => {
      const vapidPublicKey = process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY;

      if (!vapidPublicKey) {
        console.warn('Notifications not configured (missing VAPID key)');
        return;
      }

      // Subscriptions are stored per user, so we can only register once the
      // user is identified.
      const userId = localStorage.getItem('userId');
      if (!userId) return;

      await initializeNotifications(vapidPublicKey, userId);
    };

    initNotifications();
  }, []);

  return null;
}
