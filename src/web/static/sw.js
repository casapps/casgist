// CasGists Service Worker
// Provides offline functionality and caching for the PWA

const CACHE_NAME = 'casgists-v1';
const STATIC_CACHE_NAME = 'casgists-static-v1';
const DYNAMIC_CACHE_NAME = 'casgists-dynamic-v1';

// Files to cache immediately (App Shell)
const STATIC_FILES = [
  '/',
  '/static/css/tailwind.css',
  '/static/js/app.js',
  '/static/manifest.json',
  '/static/icons/icon-192x192.png',
  '/static/icons/icon-512x512.png',
  '/offline'
];

// API endpoints to cache with network-first strategy
const API_CACHE_PATTERNS = [
  /^\/api\/v1\/user$/,
  /^\/api\/v1\/user\/gists/,
  /^\/api\/v1\/gists\/\w+$/,
  /^\/api\/v1\/search/
];

// Static resources to cache with cache-first strategy
const STATIC_CACHE_PATTERNS = [
  /\.(?:png|jpg|jpeg|svg|gif|webp|ico)$/,
  /\.(?:css|js)$/,
  /\/static\//
];

// Install event - cache static files
self.addEventListener('install', event => {
  console.log('[SW] Installing Service Worker');
  
  event.waitUntil(
    caches.open(STATIC_CACHE_NAME)
      .then(cache => {
        console.log('[SW] Caching static files');
        return cache.addAll(STATIC_FILES);
      })
      .then(() => {
        console.log('[SW] Static files cached');
        return self.skipWaiting();
      })
      .catch(error => {
        console.error('[SW] Failed to cache static files:', error);
      })
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', event => {
  console.log('[SW] Activating Service Worker');
  
  event.waitUntil(
    caches.keys()
      .then(cacheNames => {
        return Promise.all(
          cacheNames
            .filter(cacheName => {
              return cacheName !== STATIC_CACHE_NAME && 
                     cacheName !== DYNAMIC_CACHE_NAME &&
                     cacheName.startsWith('casgists-');
            })
            .map(cacheName => {
              console.log('[SW] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            })
        );
      })
      .then(() => {
        console.log('[SW] Service Worker activated');
        return self.clients.claim();
      })
  );
});

// Fetch event - handle requests with appropriate strategy
self.addEventListener('fetch', event => {
  const { request } = event;
  const url = new URL(request.url);
  
  // Skip non-GET requests
  if (request.method !== 'GET') {
    return;
  }
  
  // Skip chrome-extension and other non-http requests
  if (!url.protocol.startsWith('http')) {
    return;
  }
  
  // Handle API requests with network-first strategy
  if (isApiRequest(request)) {
    event.respondWith(networkFirstStrategy(request));
    return;
  }
  
  // Handle static resources with cache-first strategy
  if (isStaticResource(request)) {
    event.respondWith(cacheFirstStrategy(request));
    return;
  }
  
  // Handle navigation requests with network-first, fallback to offline page
  if (isNavigationRequest(request)) {
    event.respondWith(navigationStrategy(request));
    return;
  }
  
  // Default strategy for other requests
  event.respondWith(networkFirstStrategy(request));
});

// Network-first strategy for API calls and dynamic content
async function networkFirstStrategy(request) {
  const url = new URL(request.url);
  
  try {
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(DYNAMIC_CACHE_NAME);
      cache.put(request, networkResponse.clone());
      
      // Also store in IndexedDB for structured data
      await cacheStructuredData(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('[SW] Network failed, trying cache:', request.url);
    
    // Try structured cache first (IndexedDB)
    const structuredData = await getStructuredCacheData(request);
    if (structuredData) {
      return new Response(JSON.stringify(structuredData), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    // Try HTTP cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Return offline response for API requests
    if (isApiRequest(request)) {
      return await generateOfflineApiResponse(request);
    }
    
    throw error;
  }
}

async function cacheStructuredData(request, response) {
  try {
    const url = new URL(request.url);
    const data = await response.json();
    const db = await initDB();
    
    // Cache gists data
    if (url.pathname.match(/^\/api\/v1\/gists\/[\w-]+$/)) {
      await db.saveGist(data);
    }
    // Cache user data
    else if (url.pathname.match(/^\/api\/v1\/users?\/[\w-]+$/)) {
      await db.saveUser(data);
    }
    // Cache API responses generically
    else {
      await db.cacheApiResponse(request.url, data, 300000); // 5 minutes TTL
    }
  } catch (error) {
    console.log('[SW] Failed to cache structured data:', error);
  }
}

async function getStructuredCacheData(request) {
  try {
    const url = new URL(request.url);
    const db = await initDB();
    
    // Get cached gist
    if (url.pathname.match(/^\/api\/v1\/gists\/([\w-]+)$/)) {
      const gistId = url.pathname.split('/').pop();
      return await db.getGist(gistId);
    }
    // Get cached user
    else if (url.pathname.match(/^\/api\/v1\/users\/([\w-]+)$/)) {
      const username = url.pathname.split('/').pop();
      return await db.getUserByUsername(username);
    }
    // Get generic cached API response
    else {
      return await db.getCachedApiResponse(request.url);
    }
  } catch (error) {
    console.log('[SW] Failed to get structured cache data:', error);
    return null;
  }
}

async function generateOfflineApiResponse(request) {
  const url = new URL(request.url);
  const db = await initDB();
  
  try {
    // Handle offline gist listing
    if (url.pathname === '/api/v1/gists' || url.pathname === '/api/v1/user/gists') {
      const searchParams = new URLSearchParams(url.search);
      const userId = searchParams.get('user');
      const limit = parseInt(searchParams.get('limit')) || 30;
      
      let gists = [];
      if (userId) {
        gists = await db.getUserGists(userId, limit);
      } else {
        gists = await db.getAll('gists');
        gists = gists.slice(0, limit);
      }
      
      return new Response(JSON.stringify({
        gists,
        offline: true,
        message: 'Showing cached gists - some data may be outdated'
      }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    // Handle offline search
    if (url.pathname === '/api/v1/search/gists') {
      const searchParams = new URLSearchParams(url.search);
      const query = searchParams.get('q') || '';
      const filters = {
        visibility: searchParams.get('visibility'),
        language: searchParams.get('language'),
        userId: searchParams.get('user')
      };
      
      const results = await db.searchGists(query, filters);
      
      return new Response(JSON.stringify({
        results,
        offline: true,
        message: 'Showing cached search results'
      }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  } catch (error) {
    console.error('[SW] Failed to generate offline API response:', error);
  }
  
  // Default offline response
  return new Response(
    JSON.stringify({
      error: 'This feature requires an internet connection',
      offline: true,
      available_offline: [
        'View cached gists',
        'Create new gists (will sync when online)',
        'Search cached content'
      ]
    }),
    {
      status: 503,
      statusText: 'Service Unavailable',
      headers: { 'Content-Type': 'application/json' }
    }
  );
}

// Cache-first strategy for static resources
async function cacheFirstStrategy(request) {
  const cachedResponse = await caches.match(request);
  
  if (cachedResponse) {
    // Update cache in background
    fetch(request)
      .then(response => {
        if (response.ok) {
          caches.open(STATIC_CACHE_NAME)
            .then(cache => cache.put(request, response));
        }
      })
      .catch(() => {}); // Ignore errors for background updates
    
    return cachedResponse;
  }
  
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(STATIC_CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.error('[SW] Failed to fetch static resource:', request.url);
    throw error;
  }
}

// Navigation strategy with offline fallback
async function navigationStrategy(request) {
  try {
    const networkResponse = await fetch(request);
    
    // Cache successful page responses
    if (networkResponse.ok) {
      const cache = await caches.open(DYNAMIC_CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('[SW] Network failed for navigation, trying cache:', request.url);
    
    // Try cache first
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Fallback to offline page
    const offlineResponse = await caches.match('/offline');
    if (offlineResponse) {
      return offlineResponse;
    }
    
    // Last resort - basic offline page
    return new Response(`
      <!DOCTYPE html>
      <html>
      <head>
        <title>Offline - CasGists</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <style>
          body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; text-align: center; padding: 50px; }
          .offline { color: #666; }
          .retry-btn { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; }
        </style>
      </head>
      <body>
        <div class="offline">
          <h1>You're Offline</h1>
          <p>Please check your internet connection and try again.</p>
          <button class="retry-btn" onclick="window.location.reload()">Retry</button>
        </div>
      </body>
      </html>
    `, {
      headers: { 'Content-Type': 'text/html' }
    });
  }
}

// Helper functions
function isApiRequest(request) {
  const url = new URL(request.url);
  return url.pathname.startsWith('/api/') || 
         API_CACHE_PATTERNS.some(pattern => pattern.test(url.pathname));
}

function isStaticResource(request) {
  const url = new URL(request.url);
  return STATIC_CACHE_PATTERNS.some(pattern => pattern.test(url.pathname));
}

function isNavigationRequest(request) {
  return request.mode === 'navigate' || 
         (request.method === 'GET' && request.headers.get('accept').includes('text/html'));
}

// Background sync for offline actions
self.addEventListener('sync', event => {
  console.log('[SW] Background sync triggered:', event.tag);
  
  if (event.tag === 'gist-sync') {
    event.waitUntil(syncOfflineGists());
  }
  
  if (event.tag === 'user-actions-sync') {
    event.waitUntil(syncOfflineUserActions());
  }
});

// Sync offline-created gists when online
async function syncOfflineGists() {
  try {
    const offlineGists = await getOfflineData('pending-gists');
    
    for (const gist of offlineGists) {
      try {
        const response = await fetch('/api/v1/gists', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(gist.data)
        });
        
        if (response.ok) {
          await removeOfflineData('pending-gists', gist.id);
          console.log('[SW] Synced offline gist:', gist.id);
          
          // Notify user of successful sync
          self.registration.showNotification('Gist Synchronized', {
            body: `Your offline gist "${gist.data.title || 'Untitled'}" has been synchronized.`,
            icon: '/static/icons/icon-192x192.png',
            badge: '/static/icons/badge-72x72.png',
            tag: 'gist-sync'
          });
        }
      } catch (error) {
        console.error('[SW] Failed to sync gist:', error);
      }
    }
  } catch (error) {
    console.error('[SW] Background sync failed:', error);
  }
}

// Sync other offline user actions
async function syncOfflineUserActions() {
  try {
    const actions = await getOfflineData('pending-actions');
    
    for (const action of actions) {
      try {
        const response = await fetch(action.url, {
          method: action.method,
          headers: action.headers,
          body: action.body
        });
        
        if (response.ok) {
          await removeOfflineData('pending-actions', action.id);
          console.log('[SW] Synced offline action:', action.type);
        }
      } catch (error) {
        console.error('[SW] Failed to sync action:', error);
      }
    }
  } catch (error) {
    console.error('[SW] Action sync failed:', error);
  }
}

// Handle push notifications
self.addEventListener('push', event => {
  if (!event.data) return;
  
  try {
    const data = event.data.json();
    
    const options = {
      body: data.body,
      icon: '/static/icons/icon-192x192.png',
      badge: '/static/icons/badge-72x72.png',
      vibrate: [200, 100, 200],
      data: data.data,
      actions: data.actions || [
        {
          action: 'view',
          title: 'View',
          icon: '/static/icons/view-icon.png'
        },
        {
          action: 'dismiss',
          title: 'Dismiss',
          icon: '/static/icons/dismiss-icon.png'
        }
      ]
    };
    
    event.waitUntil(
      self.registration.showNotification(data.title, options)
    );
  } catch (error) {
    console.error('[SW] Push notification error:', error);
  }
});

// Handle notification clicks
self.addEventListener('notificationclick', event => {
  event.notification.close();
  
  const action = event.action;
  const data = event.notification.data;
  
  if (action === 'dismiss') {
    return;
  }
  
  let url = '/';
  if (data && data.url) {
    url = data.url;
  } else if (action === 'view' && data && data.gistId) {
    url = `/gist/${data.gistId}`;
  }
  
  event.waitUntil(
    clients.matchAll({ type: 'window' }).then(clientList => {
      // Try to focus existing window
      for (const client of clientList) {
        if (client.url === url && 'focus' in client) {
          return client.focus();
        }
      }
      
      // Open new window
      if (clients.openWindow) {
        return clients.openWindow(url);
      }
    })
  );
});

// Import IndexedDB utilities
importScripts('/static/js/indexeddb.js');

// Initialize IndexedDB in service worker context
let dbInstance = null;

async function initDB() {
  if (!dbInstance) {
    dbInstance = new CasGistsDB();
    await dbInstance.init();
  }
  return dbInstance;
}

// Utility functions for offline data management
async function getOfflineData(store) {
  try {
    const db = await initDB();
    
    switch (store) {
      case 'pending-gists':
        return await db.getQueuedActions('gist-creation');
      case 'pending-actions':
        return await db.getQueuedActions();
      default:
        return [];
    }
  } catch (error) {
    console.error('[SW] Failed to get offline data:', error);
    return [];
  }
}

async function removeOfflineData(store, id) {
  try {
    const db = await initDB();
    await db.removeQueuedAction(id);
    console.log(`[SW] Removed ${id} from ${store}`);
  } catch (error) {
    console.error('[SW] Failed to remove offline data:', error);
  }
}

async function storeOfflineData(store, data) {
  try {
    const db = await initDB();
    await db.queueAction(data);
    console.log(`[SW] Stored offline data in ${store}`);
  } catch (error) {
    console.error('[SW] Failed to store offline data:', error);
  }
}

// Cache size management
async function limitCacheSize(cacheName, maxItems) {
  const cache = await caches.open(cacheName);
  const keys = await cache.keys();
  
  if (keys.length > maxItems) {
    const keysToDelete = keys.slice(0, keys.length - maxItems);
    await Promise.all(keysToDelete.map(key => cache.delete(key)));
    console.log(`[SW] Trimmed ${keysToDelete.length} items from ${cacheName}`);
  }
}

// Periodic cache cleanup
setInterval(() => {
  limitCacheSize(DYNAMIC_CACHE_NAME, 50);
  limitCacheSize(STATIC_CACHE_NAME, 100);
}, 60000 * 60); // Every hour

console.log('[SW] Service Worker script loaded');