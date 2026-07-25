// Co-Drive Push Notifications Client

const VAPID_PUBLIC_KEY = '{{VAPID_PUBLIC_KEY}}';

async function registerServiceWorker() {
  if (!('serviceWorker' in navigator)) {
    console.warn('Service Workers not supported');
    return null;
  }

  try {
    const registration = await navigator.serviceWorker.register('/static/sw.js');
    console.log('Service Worker registered:', registration);
    return registration;
  } catch (error) {
    console.error('Service Worker registration failed:', error);
    return null;
  }
}

async function requestNotificationPermission() {
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

async function subscribeToPush(registration) {
  try {
    const applicationServerKey = urlBase64ToUint8Array(VAPID_PUBLIC_KEY);
    const subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey
    });
    console.log('Push subscription created:', subscription);
    return subscription;
  } catch (error) {
    console.error('Push subscription failed:', error);
    return null;
  }
}

async function savePushSubscription(subscription) {
  try {
    const response = await fetch('/api/notifications/subscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        subscription: subscription.toJSON()
      })
    });

    if (!response.ok) {
      throw new Error('Failed to save subscription');
    }

    console.log('Push subscription saved');
  } catch (error) {
    console.error('Failed to save subscription:', error);
  }
}

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

async function initializeNotifications() {
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

  const subscription = await subscribeToPush(swRegistration);
  if (subscription) {
    await savePushSubscription(subscription);
  }
}

// Auto-initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  if (VAPID_PUBLIC_KEY && VAPID_PUBLIC_KEY !== '{{VAPID_PUBLIC_KEY}}') {
    initializeNotifications();
  }
});

// Expose for manual triggering
window.CoDriveNotifications = {
  initialize: initializeNotifications,
  requestPermission: requestNotificationPermission,
  subscribe: async () => {
    const reg = await registerServiceWorker();
    if (reg) return await subscribeToPush(reg);
  }
};