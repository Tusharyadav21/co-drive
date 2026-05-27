self.addEventListener('push', (event) => {
  const data = event.data ? event.data.json() : {};
  const options = {
    body: data.body || 'Vehicle maintenance reminder',
    icon: '/icon.png',
    badge: '/badge.png',
    data: data.data || {},
  };

  event.waitUntil(self.registration.showNotification(data.title || 'Vehicle Tracker', options));
});

self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  const urlToOpen = event.notification.data.url || '/';

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (let i = 0; i < clientList.length; i++) {
        const client = clientList[i];
        if (client.url === urlToOpen && 'focus' in client) {
          return client.focus();
        }
      }
      if (clients.openWindow) {
        return clients.openWindow(urlToOpen);
      }
    })
  );
});
