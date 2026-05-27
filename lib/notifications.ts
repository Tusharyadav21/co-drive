export async function registerServiceWorker() {
  if (!('serviceWorker' in navigator)) {
    console.warn('Service Workers not supported');
    return null;
  }

  try {
    const registration = await navigator.serviceWorker.register('/sw.js');
    console.warn('Service Worker registered:', registration);
    return registration;
  } catch (error) {
    console.error('Service Worker registration failed:', error);
    return null;
  }
}

export async function requestNotificationPermission() {
  if (!('Notification' in window)) {
    console.warn('Notifications not supported');
    return false;
  }

  if (Notification.permission === 'granted') {
    return true;
  }

  if (Notification.permission !== 'denied') {
    const permission = await Notification.requestPermission();
    return permission === 'granted';
  }

  return false;
}

export async function subscribeToPush(vapidPublicKey: string): Promise<PushSubscription | null> {
  const registration = await navigator.serviceWorker.ready;

  try {
    const subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: vapidPublicKey,
    });
    console.warn('Push subscription created:', subscription);
    return subscription;
  } catch (error) {
    console.error('Push subscription failed:', error);
    return null;
  }
}

export async function savePushSubscription(subscription: PushSubscription, userId: string) {
  try {
    const response = await fetch('/api/notifications/subscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        userId,
        subscription: subscription.toJSON(),
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to save subscription');
    }

    console.warn('Push subscription saved');
  } catch (error) {
    console.error('Failed to save subscription:', error);
  }
}

export async function initializeNotifications(vapidPublicKey: string, userId: string) {
  const hasPermission = await requestNotificationPermission();
  if (!hasPermission) {
    console.warn('Notification permission denied');
    return;
  }

  const swRegistration = await registerServiceWorker();
  if (!swRegistration) {
    console.warn('Service worker registration failed');
    return;
  }

  const subscription = await subscribeToPush(vapidPublicKey);
  if (subscription) {
    await savePushSubscription(subscription, userId);
  }
}
