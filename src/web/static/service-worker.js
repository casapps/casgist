// service-worker.js - CasGists PWA Service Worker
const CACHE_NAME = 'casgists-v1';
const urlsToCache = [
  '/',
  '/static/css/tailwind.css',
  '/static/css/app.css',
  '/static/js/app.js',
  '/static/js/htmx.min.js',
  '/static/js/prism.js',
  '/static/manifest.json',
  '/static/icons/icon-192x192.png',
  '/static/icons/icon-512x512.png',
  '/static/fonts/inter/inter.woff2',
  '/healthz',
  '/auth/login',
  '/user/dashboard'
];

// Cache strategies
const CACHE_STRATEGIES = {
  CACHE_FIRST: 'cache-first',
  NETWORK_FIRST: 'network-first',
  STALE_WHILE_REVALIDATE: 'stale-while-revalidate',
  NETWORK_ONLY: 'network-only',
  CACHE_ONLY: 'cache-only'
};

// URL patterns and their cache strategies
const URL_PATTERNS = [
  { pattern: /\.(css|js|woff2?|png|jpg|jpeg|gif|svg|ico)$/, strategy: CACHE_STRATEGIES.CACHE_FIRST },
  { pattern: /^\/static\//, strategy: CACHE_STRATEGIES.CACHE_FIRST },
  { pattern: /^\/api\/v1\/gists\/public/, strategy: CACHE_STRATEGIES.STALE_WHILE_REVALIDATE },
  { pattern: /^\/api\/v1\/explore/, strategy: CACHE_STRATEGIES.STALE_WHILE_REVALIDATE },
  { pattern: /^\/healthz$/, strategy: CACHE_STRATEGIES.NETWORK_FIRST },
  { pattern: /^\/api\/v1\//, strategy: CACHE_STRATEGIES.NETWORK_FIRST },
  { pattern: /.*/, strategy: CACHE_STRATEGIES.NETWORK_FIRST } // Default fallback
];

// Install event - cache essential resources
self.addEventListener('install', (event) => {
  console.log('[ServiceWorker] Install event');
  
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('[ServiceWorker] Caching essential resources');
        return cache.addAll(urlsToCache);
      })
      .then(() => {
        console.log('[ServiceWorker] All essential resources cached');
        // Skip waiting to activate new service worker immediately
        return self.skipWaiting();
      })
      .catch((error) => {
        console.error('[ServiceWorker] Cache installation failed:', error);
      })
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('[ServiceWorker] Activate event');
  
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (cacheName !== CACHE_NAME) {
              console.log('[ServiceWorker] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => {
        console.log('[ServiceWorker] New service worker activated');
        // Claim all clients immediately
        return self.clients.claim();
      })
  );
});

// Fetch event - serve from cache with different strategies
self.addEventListener('fetch', (event) => {
  // Skip non-GET requests
  if (event.request.method !== 'GET') {
    return;
  }

  // Skip chrome-extension and other non-http requests
  if (!event.request.url.startsWith('http')) {
    return;
  }

  const url = new URL(event.request.url);
  const strategy = getStrategyForUrl(url.pathname);

  event.respondWith(
    handleRequest(event.request, strategy)
      .catch((error) => {
        console.error('[ServiceWorker] Fetch failed:', error);
        return handleFallback(event.request);
      })
  );
});

// Get cache strategy for URL
function getStrategyForUrl(pathname) {
  for (const { pattern, strategy } of URL_PATTERNS) {
    if (pattern.test(pathname)) {
      return strategy;
    }
  }
  return CACHE_STRATEGIES.NETWORK_FIRST; // Default fallback
}

// Handle request based on strategy
async function handleRequest(request, strategy) {
  switch (strategy) {
    case CACHE_STRATEGIES.CACHE_FIRST:
      return handleCacheFirst(request);
    
    case CACHE_STRATEGIES.NETWORK_FIRST:
      return handleNetworkFirst(request);
    
    case CACHE_STRATEGIES.STALE_WHILE_REVALIDATE:
      return handleStaleWhileRevalidate(request);
    
    case CACHE_STRATEGIES.CACHE_ONLY:
      return handleCacheOnly(request);
    
    case CACHE_STRATEGIES.NETWORK_ONLY:
      return handleNetworkOnly(request);
    
    default:
      return handleNetworkFirst(request);
  }
}

// Cache first strategy - serve from cache, fallback to network
async function handleCacheFirst(request) {
  const cache = await caches.open(CACHE_NAME);
  const cachedResponse = await cache.match(request);
  
  if (cachedResponse) {
    return cachedResponse;
  }
  
  const networkResponse = await fetch(request);
  
  // Cache successful responses
  if (networkResponse.ok) {
    cache.put(request, networkResponse.clone());
  }
  
  return networkResponse;
}

// Network first strategy - try network, fallback to cache
async function handleNetworkFirst(request) {
  try {
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    // Network failed, try cache
    const cache = await caches.open(CACHE_NAME);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }
    
    throw error;
  }
}

// Stale while revalidate - serve from cache, update in background
async function handleStaleWhileRevalidate(request) {
  const cache = await caches.open(CACHE_NAME);
  const cachedResponse = await cache.match(request);
  
  // Start network request in background
  const networkResponsePromise = fetch(request).then((response) => {
    if (response.ok) {
      cache.put(request, response.clone());
    }
    return response;
  }).catch(() => {
    // Network failed, but we might have cache
    return null;
  });
  
  // Return cached version immediately if available
  if (cachedResponse) {
    return cachedResponse;
  }
  
  // Wait for network if no cache
  return networkResponsePromise;
}

// Cache only strategy
async function handleCacheOnly(request) {
  const cache = await caches.open(CACHE_NAME);
  return cache.match(request);
}

// Network only strategy
async function handleNetworkOnly(request) {
  return fetch(request);
}

// Fallback handler for failed requests
async function handleFallback(request) {
  const url = new URL(request.url);
  
  // For HTML pages, serve offline fallback
  if (request.headers.get('Accept').includes('text/html')) {
    return caches.match('/offline.html') || 
           new Response('<!DOCTYPE html><html><head><title>Offline - CasGists</title></head><body><h1>You are offline</h1><p>Please check your internet connection.</p></body></html>', {
             headers: { 'Content-Type': 'text/html' }
           });
  }
  
  // For API requests, return offline JSON response
  if (url.pathname.startsWith('/api/')) {
    return new Response(JSON.stringify({
      error: {
        code: 'OFFLINE',
        message: 'You are currently offline. Please check your internet connection.',
        offline: true
      }
    }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' }
    });
  }
  
  // For other requests, return a generic offline response
  return new Response('Offline', { status: 503 });
}

// Background sync for offline actions
self.addEventListener('sync', (event) => {
  console.log('[ServiceWorker] Background sync:', event.tag);
  
  if (event.tag === 'background-sync') {
    event.waitUntil(handleBackgroundSync());
  }
});

// Handle background sync
async function handleBackgroundSync() {
  try {
    console.log('[ServiceWorker] Performing background sync');
    
    // Get pending actions from IndexedDB
    const pendingActions = await getPendingActions();
    
    for (const action of pendingActions) {
      try {
        await syncAction(action);
        await removePendingAction(action.id);
        console.log('[ServiceWorker] Synced action:', action.type);
      } catch (error) {
        console.error('[ServiceWorker] Failed to sync action:', action.type, error);
      }
    }
    
  } catch (error) {
    console.error('[ServiceWorker] Background sync failed:', error);
  }
}

// Get pending actions from IndexedDB (stub - would need full implementation)
async function getPendingActions() {
  // In a full implementation, this would read from IndexedDB
  return [];
}

// Sync individual action (stub - would need full implementation)
async function syncAction(action) {
  // In a full implementation, this would replay the action
  switch (action.type) {
    case 'CREATE_GIST':
      return fetch('/api/v1/gists', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(action.data)
      });
    
    case 'UPDATE_GIST':
      return fetch(`/api/v1/gists/${action.gistId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(action.data)
      });
    
    default:
      console.warn('[ServiceWorker] Unknown action type:', action.type);
  }
}

// Remove pending action from IndexedDB (stub)
async function removePendingAction(actionId) {
  // In a full implementation, this would remove from IndexedDB
}

// Push notification handling
self.addEventListener('push', (event) => {
  console.log('[ServiceWorker] Push received');
  
  let notificationData = { title: 'CasGists', body: 'New notification' };
  
  if (event.data) {
    try {
      notificationData = event.data.json();
    } catch (error) {
      console.error('[ServiceWorker] Invalid push data:', error);
    }
  }
  
  const notificationOptions = {
    body: notificationData.body,
    icon: '/static/icons/icon-192x192.png',
    badge: '/static/icons/icon-72x72.png',
    vibrate: [200, 100, 200],
    data: notificationData.data || {},
    actions: [
      {
        action: 'open',
        title: 'Open',
        icon: '/static/icons/icon-72x72.png'
      },
      {
        action: 'close',
        title: 'Close'
      }
    ]
  };
  
  event.waitUntil(
    self.registration.showNotification(notificationData.title, notificationOptions)
  );
});

// Notification click handling
self.addEventListener('notificationclick', (event) => {
  console.log('[ServiceWorker] Notification clicked');
  
  event.notification.close();
  
  if (event.action === 'close') {
    return;
  }
  
  // Default action or 'open' action
  const urlToOpen = event.notification.data.url || '/';
  
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true })
      .then((clientList) => {
        // Check if there's already a window/tab open with the target URL
        for (const client of clientList) {
          if (client.url === urlToOpen && 'focus' in client) {
            return client.focus();
          }
        }
        
        // No existing window, open new one
        if (clients.openWindow) {
          return clients.openWindow(urlToOpen);
        }
      })
  );
});

// Message handling for communication with main thread
self.addEventListener('message', (event) => {
  console.log('[ServiceWorker] Message received:', event.data);
  
  switch (event.data.type) {
    case 'SKIP_WAITING':
      self.skipWaiting();
      break;
    
    case 'GET_VERSION':
      event.ports[0].postMessage({ version: CACHE_NAME });
      break;
    
    case 'CLEAR_CACHE':
      event.waitUntil(
        caches.delete(CACHE_NAME)
          .then(() => event.ports[0].postMessage({ success: true }))
          .catch((error) => event.ports[0].postMessage({ success: false, error }))
      );
      break;
    
    case 'CACHE_URLs':
      event.waitUntil(
        cacheUrls(event.data.urls)
          .then(() => event.ports[0].postMessage({ success: true }))
          .catch((error) => event.ports[0].postMessage({ success: false, error }))
      );
      break;
    
    default:
      console.warn('[ServiceWorker] Unknown message type:', event.data.type);
  }
});

// Cache specific URLs
async function cacheUrls(urls) {
  const cache = await caches.open(CACHE_NAME);
  return cache.addAll(urls);
}

// Periodic background sync (if supported)
self.addEventListener('periodicsync', (event) => {
  console.log('[ServiceWorker] Periodic sync:', event.tag);
  
  if (event.tag === 'content-sync') {
    event.waitUntil(handlePeriodicSync());
  }
});

// Handle periodic sync
async function handlePeriodicSync() {
  try {
    console.log('[ServiceWorker] Performing periodic sync');
    
    // Update cached content in background
    const cache = await caches.open(CACHE_NAME);
    
    const urlsToUpdate = [
      '/api/v1/gists/public',
      '/api/v1/explore/trending',
      '/healthz'
    ];
    
    for (const url of urlsToUpdate) {
      try {
        const response = await fetch(url);
        if (response.ok) {
          await cache.put(url, response);
          console.log('[ServiceWorker] Updated cache for:', url);
        }
      } catch (error) {
        console.warn('[ServiceWorker] Failed to update cache for:', url, error);
      }
    }
    
  } catch (error) {
    console.error('[ServiceWorker] Periodic sync failed:', error);
  }
}

console.log('[ServiceWorker] Service worker loaded');