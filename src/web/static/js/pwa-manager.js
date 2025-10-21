// PWA registration and management for CasGists
class PWAManager {
  constructor() {
    this.serviceWorker = null;
    this.updateAvailable = false;
    this.deferredPrompt = null;
    this.init();
  }

  async init() {
    if ('serviceWorker' in navigator) {
      try {
        await this.registerServiceWorker();
        this.setupUpdateHandling();
        this.setupInstallPrompt();
        this.setupBackgroundSync();
        this.setupNotifications();
      } catch (error) {
        console.error('PWA initialization failed:', error);
      }
    } else {
      console.warn('Service workers not supported in this browser');
    }
  }

  async registerServiceWorker() {
    try {
      console.log('Registering service worker...');
      
      const registration = await navigator.serviceWorker.register('/service-worker.js', {
        scope: '/'
      });

      this.serviceWorker = registration;
      console.log('Service worker registered successfully:', registration);

      // Listen for updates
      registration.addEventListener('updatefound', () => {
        console.log('Service worker update found');
        const newWorker = registration.installing;
        
        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            console.log('Service worker update available');
            this.updateAvailable = true;
            this.showUpdateNotification();
          }
        });
      });

    } catch (error) {
      console.error('Service worker registration failed:', error);
      throw error;
    }
  }

  setupUpdateHandling() {
    // Listen for messages from service worker
    navigator.serviceWorker.addEventListener('message', (event) => {
      console.log('Message from service worker:', event.data);
      
      if (event.data.type === 'UPDATE_AVAILABLE') {
        this.updateAvailable = true;
        this.showUpdateNotification();
      }
    });

    // Check for updates periodically
    setInterval(() => {
      if (this.serviceWorker) {
        this.serviceWorker.update();
      }
    }, 60000); // Check every minute
  }

  showUpdateNotification() {
    // Create update notification
    const notification = document.createElement('div');
    notification.className = 'pwa-update-notification';
    notification.innerHTML = `
      <div class="bg-blue-600 text-white p-4 rounded-lg shadow-lg fixed top-4 right-4 z-50 max-w-sm">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="font-semibold">Update Available</h4>
            <p class="text-sm">A new version of CasGists is available.</p>
          </div>
          <button id="pwa-update-btn" class="ml-4 bg-blue-500 hover:bg-blue-700 px-3 py-1 rounded text-sm">
            Update
          </button>
        </div>
      </div>
    `;

    document.body.appendChild(notification);

    // Handle update button click
    document.getElementById('pwa-update-btn').addEventListener('click', () => {
      this.applyUpdate();
      notification.remove();
    });

    // Auto-hide after 10 seconds
    setTimeout(() => {
      if (notification.parentNode) {
        notification.remove();
      }
    }, 10000);
  }

  async applyUpdate() {
    if (this.serviceWorker && this.serviceWorker.waiting) {
      console.log('Applying service worker update...');
      
      // Send message to service worker to skip waiting
      this.serviceWorker.waiting.postMessage({ type: 'SKIP_WAITING' });
      
      // Reload page to activate new service worker
      window.location.reload();
    }
  }

  setupInstallPrompt() {
    // Capture the install prompt
    window.addEventListener('beforeinstallprompt', (event) => {
      console.log('Install prompt event captured');
      
      // Prevent the default prompt
      event.preventDefault();
      
      // Save the event for later use
      this.deferredPrompt = event;
      
      // Show custom install button
      this.showInstallButton();
    });

    // Listen for app installed event
    window.addEventListener('appinstalled', () => {
      console.log('PWA was installed');
      this.hideInstallButton();
      this.deferredPrompt = null;
      
      // Track installation
      this.trackEvent('pwa_installed');
    });
  }

  showInstallButton() {
    const existingButton = document.getElementById('pwa-install-btn');
    if (existingButton) return;

    const installButton = document.createElement('button');
    installButton.id = 'pwa-install-btn';
    installButton.className = 'fixed bottom-4 right-4 bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg shadow-lg z-50 flex items-center space-x-2';
    installButton.innerHTML = `
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
      </svg>
      <span>Install App</span>
    `;

    installButton.addEventListener('click', () => {
      this.promptInstall();
    });

    document.body.appendChild(installButton);
  }

  hideInstallButton() {
    const installButton = document.getElementById('pwa-install-btn');
    if (installButton) {
      installButton.remove();
    }
  }

  async promptInstall() {
    if (!this.deferredPrompt) {
      console.warn('No install prompt available');
      return;
    }

    try {
      // Show the install prompt
      this.deferredPrompt.prompt();

      // Wait for user response
      const { outcome } = await this.deferredPrompt.userChoice;
      console.log('Install prompt outcome:', outcome);

      if (outcome === 'accepted') {
        console.log('User accepted install prompt');
        this.trackEvent('pwa_install_accepted');
      } else {
        console.log('User dismissed install prompt');
        this.trackEvent('pwa_install_dismissed');
      }

      // Clear the deferred prompt
      this.deferredPrompt = null;
      this.hideInstallButton();

    } catch (error) {
      console.error('Install prompt failed:', error);
    }
  }

  setupBackgroundSync() {
    // Register for background sync
    if ('serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype) {
      console.log('Background sync supported');
      
      // Listen for online/offline events
      window.addEventListener('online', () => {
        console.log('Back online - triggering sync');
        this.triggerBackgroundSync();
      });

      window.addEventListener('offline', () => {
        console.log('Gone offline - background sync will activate when online');
        this.showOfflineNotification();
      });
    }
  }

  async triggerBackgroundSync() {
    try {
      if (this.serviceWorker) {
        await this.serviceWorker.sync.register('background-sync');
        console.log('Background sync registered');
      }
    } catch (error) {
      console.error('Background sync registration failed:', error);
    }
  }

  showOfflineNotification() {
    // Show offline indicator
    const offlineIndicator = document.createElement('div');
    offlineIndicator.id = 'offline-indicator';
    offlineIndicator.className = 'fixed top-0 left-0 right-0 bg-yellow-600 text-white text-center py-2 z-50';
    offlineIndicator.innerHTML = `
      <div class="flex items-center justify-center space-x-2">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192L5.636 18.364M12 2.25a9.75 9.75 0 110 19.5 9.75 9.75 0 010-19.5z"/>
        </svg>
        <span>You are currently offline</span>
      </div>
    `;

    document.body.appendChild(offlineIndicator);

    // Remove when back online
    window.addEventListener('online', () => {
      const indicator = document.getElementById('offline-indicator');
      if (indicator) {
        indicator.remove();
      }
    });
  }

  async setupNotifications() {
    // Check notification support
    if ('Notification' in window && 'serviceWorker' in navigator) {
      console.log('Notifications supported');
      
      // Request permission if not already granted
      if (Notification.permission === 'default') {
        // Show notification permission prompt later, not immediately
        setTimeout(() => {
          this.showNotificationPermissionPrompt();
        }, 30000); // Wait 30 seconds before asking
      }
    }
  }

  showNotificationPermissionPrompt() {
    // Create custom notification permission prompt
    const promptModal = document.createElement('div');
    promptModal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
    promptModal.innerHTML = `
      <div class="bg-white dark:bg-gray-800 p-6 rounded-lg max-w-md mx-4">
        <h3 class="text-lg font-semibold mb-4">Enable Notifications</h3>
        <p class="text-gray-600 dark:text-gray-300 mb-4">
          Get notified about gist comments, stars, and other activity even when CasGists isn't open.
        </p>
        <div class="flex space-x-3">
          <button id="enable-notifications" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded">
            Enable
          </button>
          <button id="skip-notifications" class="bg-gray-300 hover:bg-gray-400 text-gray-700 px-4 py-2 rounded">
            Skip
          </button>
        </div>
      </div>
    `;

    document.body.appendChild(promptModal);

    // Handle enable button
    document.getElementById('enable-notifications').addEventListener('click', async () => {
      const permission = await Notification.requestPermission();
      console.log('Notification permission:', permission);
      
      if (permission === 'granted') {
        this.trackEvent('notifications_enabled');
      } else {
        this.trackEvent('notifications_denied');
      }
      
      promptModal.remove();
    });

    // Handle skip button
    document.getElementById('skip-notifications').addEventListener('click', () => {
      this.trackEvent('notifications_skipped');
      promptModal.remove();
    });
  }

  // Utility methods
  async getCacheStatus() {
    if (!this.serviceWorker) return null;

    return new Promise((resolve) => {
      const messageChannel = new MessageChannel();
      
      messageChannel.port1.onmessage = (event) => {
        resolve(event.data);
      };

      navigator.serviceWorker.controller?.postMessage(
        { type: 'GET_VERSION' },
        [messageChannel.port2]
      );
    });
  }

  async clearCache() {
    if (!this.serviceWorker) return false;

    return new Promise((resolve) => {
      const messageChannel = new MessageChannel();
      
      messageChannel.port1.onmessage = (event) => {
        resolve(event.data.success);
      };

      navigator.serviceWorker.controller?.postMessage(
        { type: 'CLEAR_CACHE' },
        [messageChannel.port2]
      );
    });
  }

  async cacheUrls(urls) {
    if (!this.serviceWorker) return false;

    return new Promise((resolve) => {
      const messageChannel = new MessageChannel();
      
      messageChannel.port1.onmessage = (event) => {
        resolve(event.data.success);
      };

      navigator.serviceWorker.controller?.postMessage(
        { type: 'CACHE_URLs', urls },
        [messageChannel.port2]
      );
    });
  }

  trackEvent(eventName, data = {}) {
    console.log('PWA Event:', eventName, data);
    
    // Send to analytics if available
    if (typeof gtag !== 'undefined') {
      gtag('event', eventName, {
        event_category: 'pwa',
        ...data
      });
    }
  }

  // Check if app is running as PWA
  isPWA() {
    return window.matchMedia('(display-mode: standalone)').matches ||
           window.navigator.standalone === true;
  }
}

// Initialize PWA when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    window.pwaManager = new PWAManager();
  });
} else {
  window.pwaManager = new PWAManager();
}

// Export for external use
window.PWAManager = PWAManager;