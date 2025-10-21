// CasGists PWA Integration
// Handles service worker registration, installation prompts, and offline functionality

class CasGistsPWA {
    constructor() {
        this.deferredPrompt = null;
        this.isOnline = navigator.onLine;
        this.installButton = null;
        this.offlineQueue = [];
        
        this.init();
    }
    
    async init() {
        // Register service worker
        if ('serviceWorker' in navigator) {
            try {
                const registration = await navigator.serviceWorker.register('/static/sw.js');
                console.log('Service Worker registered:', registration);
                
                // Listen for updates
                registration.addEventListener('updatefound', () => {
                    this.handleServiceWorkerUpdate(registration);
                });
            } catch (error) {
                console.error('Service Worker registration failed:', error);
            }
        }
        
        // Set up PWA install prompt
        this.setupInstallPrompt();
        
        // Set up offline/online detection
        this.setupOfflineDetection();
        
        // Set up background sync
        this.setupBackgroundSync();
        
        // Set up push notifications
        this.setupPushNotifications();
        
        // Initialize offline UI
        this.initializeOfflineUI();
    }
    
    setupInstallPrompt() {
        // Listen for install prompt
        window.addEventListener('beforeinstallprompt', (e) => {
            e.preventDefault();
            this.deferredPrompt = e;
            this.showInstallButton();
        });
        
        // Listen for app installed
        window.addEventListener('appinstalled', () => {
            console.log('PWA installed');
            this.hideInstallButton();
            this.showToast('App installed successfully!', 'success');
        });
        
        // Create install button
        this.createInstallButton();
    }
    
    createInstallButton() {
        // Check if already running as PWA
        if (window.matchMedia('(display-mode: standalone)').matches) {
            return;
        }
        
        // Create install button
        this.installButton = document.createElement('button');
        this.installButton.innerHTML = `
            <i class="fas fa-download mr-2"></i>
            Install App
        `;
        this.installButton.className = 'btn btn-primary btn-sm fixed bottom-4 right-4 z-50 shadow-lg';
        this.installButton.style.display = 'none';
        this.installButton.id = 'pwa-install-btn';
        
        this.installButton.addEventListener('click', () => {
            this.promptInstall();
        });
        
        document.body.appendChild(this.installButton);
    }
    
    showInstallButton() {
        if (this.installButton && !this.isRunningAsPWA()) {
            this.installButton.style.display = 'block';
            
            // Auto-hide after 10 seconds
            setTimeout(() => {
                if (this.installButton) {
                    this.installButton.style.display = 'none';
                }
            }, 10000);
        }
    }
    
    hideInstallButton() {
        if (this.installButton) {
            this.installButton.style.display = 'none';
        }
    }
    
    async promptInstall() {
        if (!this.deferredPrompt) {
            this.showToast('Installation not available', 'warning');
            return;
        }
        
        try {
            this.deferredPrompt.prompt();
            const { outcome } = await this.deferredPrompt.userChoice;
            
            if (outcome === 'accepted') {
                console.log('User accepted install prompt');
            } else {
                console.log('User dismissed install prompt');
            }
            
            this.deferredPrompt = null;
            this.hideInstallButton();
        } catch (error) {
            console.error('Install prompt failed:', error);
        }
    }
    
    setupOfflineDetection() {
        // Listen for online/offline events
        window.addEventListener('online', () => {
            this.isOnline = true;
            this.handleOnlineStatus();
        });
        
        window.addEventListener('offline', () => {
            this.isOnline = false;
            this.handleOfflineStatus();
        });
        
        // Initial status
        this.updateOnlineStatus();
    }
    
    handleOnlineStatus() {
        console.log('App is online');
        this.hideOfflineIndicator();
        this.syncOfflineData();
        this.showToast('You\'re back online!', 'success');
    }
    
    handleOfflineStatus() {
        console.log('App is offline');
        this.showOfflineIndicator();
        this.showToast('You\'re offline. Some features may be limited.', 'warning');
    }
    
    updateOnlineStatus() {
        if (this.isOnline) {
            this.hideOfflineIndicator();
        } else {
            this.showOfflineIndicator();
        }
    }
    
    setupBackgroundSync() {
        // Register for background sync when service worker is ready
        if ('serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype) {
            navigator.serviceWorker.ready.then(registration => {
                this.syncRegistration = registration;
            });
        }
    }
    
    async requestBackgroundSync(tag) {
        if (this.syncRegistration) {
            try {
                await this.syncRegistration.sync.register(tag);
                console.log('Background sync registered:', tag);
            } catch (error) {
                console.error('Background sync registration failed:', error);
            }
        }
    }
    
    setupPushNotifications() {
        // Request notification permission
        if ('Notification' in window && 'serviceWorker' in navigator) {
            if (Notification.permission === 'default') {
                this.showNotificationPermissionPrompt();
            }
        }
    }
    
    async showNotificationPermissionPrompt() {
        // Show a user-friendly prompt before requesting permission
        const promptElement = document.createElement('div');
        promptElement.className = 'alert alert-info fixed top-4 right-4 max-w-sm z-50 shadow-lg';
        promptElement.innerHTML = `
            <div>
                <h3 class="font-bold">Stay Updated</h3>
                <div class="text-sm">Get notified about important updates and activity.</div>
            </div>
            <div class="flex-none">
                <button class="btn btn-sm" onclick="this.parentElement.parentElement.remove()">Later</button>
                <button class="btn btn-sm btn-primary ml-2" onclick="window.casgistsPWA.requestNotificationPermission(); this.parentElement.parentElement.remove()">Enable</button>
            </div>
        `;
        
        document.body.appendChild(promptElement);
        
        // Auto-remove after 10 seconds
        setTimeout(() => {
            if (promptElement.parentElement) {
                promptElement.remove();
            }
        }, 10000);
    }
    
    async requestNotificationPermission() {
        try {
            const permission = await Notification.requestPermission();
            
            if (permission === 'granted') {
                console.log('Notification permission granted');
                this.showToast('Notifications enabled!', 'success');
                
                // Subscribe to push notifications
                this.subscribeToPushNotifications();
            } else {
                console.log('Notification permission denied');
            }
        } catch (error) {
            console.error('Notification permission request failed:', error);
        }
    }
    
    async subscribeToPushNotifications() {
        if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
            console.log('Push notifications not supported');
            return;
        }
        
        try {
            const registration = await navigator.serviceWorker.ready;
            
            // Check if already subscribed
            const existingSubscription = await registration.pushManager.getSubscription();
            if (existingSubscription) {
                console.log('Already subscribed to push notifications');
                return;
            }
            
            // Subscribe to push notifications
            const subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: this.urlBase64ToUint8Array('YOUR_VAPID_PUBLIC_KEY') // Replace with actual VAPID key
            });
            
            // Send subscription to server
            await this.sendSubscriptionToServer(subscription);
            
            console.log('Subscribed to push notifications');
        } catch (error) {
            console.error('Push subscription failed:', error);
        }
    }
    
    async sendSubscriptionToServer(subscription) {
        try {
            await fetch('/api/v1/push/subscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': window.CasGists?.csrf
                },
                body: JSON.stringify(subscription)
            });
        } catch (error) {
            console.error('Failed to send subscription to server:', error);
        }
    }
    
    initializeOfflineUI() {
        // Create offline indicator
        this.createOfflineIndicator();
        
        // Create offline queue indicator
        this.createOfflineQueueIndicator();
        
        // Intercept form submissions for offline queueing
        this.interceptOfflineActions();
    }
    
    createOfflineIndicator() {
        this.offlineIndicator = document.createElement('div');
        this.offlineIndicator.className = 'alert alert-warning fixed top-20 left-1/2 transform -translate-x-1/2 z-50 shadow-lg max-w-sm';
        this.offlineIndicator.style.display = 'none';
        this.offlineIndicator.id = 'offline-indicator';
        this.offlineIndicator.innerHTML = `
            <i class="fas fa-wifi-slash"></i>
            <span>You're offline</span>
        `;
        
        document.body.appendChild(this.offlineIndicator);
    }
    
    createOfflineQueueIndicator() {
        this.queueIndicator = document.createElement('div');
        this.queueIndicator.className = 'fixed bottom-20 right-4 z-50';
        this.queueIndicator.style.display = 'none';
        this.queueIndicator.id = 'offline-queue-indicator';
        
        document.body.appendChild(this.queueIndicator);
    }
    
    showOfflineIndicator() {
        if (this.offlineIndicator) {
            this.offlineIndicator.style.display = 'flex';
        }
    }
    
    hideOfflineIndicator() {
        if (this.offlineIndicator) {
            this.offlineIndicator.style.display = 'none';
        }
    }
    
    updateOfflineQueueIndicator() {
        if (!this.queueIndicator) return;
        
        if (this.offlineQueue.length > 0) {
            this.queueIndicator.innerHTML = `
                <div class="badge badge-warning">
                    <i class="fas fa-clock mr-1"></i>
                    ${this.offlineQueue.length} pending
                </div>
            `;
            this.queueIndicator.style.display = 'block';
        } else {
            this.queueIndicator.style.display = 'none';
        }
    }
    
    interceptOfflineActions() {
        // Intercept gist creation forms
        document.addEventListener('submit', (event) => {
            if (!this.isOnline && event.target.id === 'gist-form') {
                event.preventDefault();
                this.queueGistCreation(event.target);
            }
        });
        
        // Intercept API calls
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            if (!this.isOnline && this.isWriteOperation(args[0], args[1])) {
                return this.queueOfflineAction(...args);
            }
            return originalFetch(...args);
        };
    }
    
    isWriteOperation(url, options) {
        const method = options?.method || 'GET';
        return ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method.toUpperCase());
    }
    
    async queueGistCreation(form) {
        const formData = new FormData(form);
        const gistData = Object.fromEntries(formData.entries());
        
        const queueItem = {
            id: Date.now().toString(),
            type: 'gist-creation',
            data: gistData,
            timestamp: new Date().toISOString()
        };
        
        this.offlineQueue.push(queueItem);
        this.updateOfflineQueueIndicator();
        
        // Store in localStorage for persistence
        localStorage.setItem('offline-queue', JSON.stringify(this.offlineQueue));
        
        this.showToast('Gist queued for creation when online', 'info');
        
        // Request background sync
        await this.requestBackgroundSync('gist-sync');
    }
    
    async queueOfflineAction(url, options) {
        const queueItem = {
            id: Date.now().toString(),
            type: 'api-call',
            url: url,
            options: options,
            timestamp: new Date().toISOString()
        };
        
        this.offlineQueue.push(queueItem);
        this.updateOfflineQueueIndicator();
        
        localStorage.setItem('offline-queue', JSON.stringify(this.offlineQueue));
        
        // Request background sync
        await this.requestBackgroundSync('user-actions-sync');
        
        // Return a response that indicates the action was queued
        return new Response(
            JSON.stringify({
                queued: true,
                message: 'Action queued for when you\'re back online'
            }),
            {
                status: 202,
                headers: { 'Content-Type': 'application/json' }
            }
        );
    }
    
    async syncOfflineData() {
        // Load queued actions from localStorage
        const stored = localStorage.getItem('offline-queue');
        if (stored) {
            try {
                this.offlineQueue = JSON.parse(stored);
            } catch (error) {
                console.error('Failed to load offline queue:', error);
                this.offlineQueue = [];
            }
        }
        
        // Process queued actions
        const failedActions = [];
        
        for (const action of this.offlineQueue) {
            try {
                if (action.type === 'gist-creation') {
                    await this.syncGistCreation(action);
                } else if (action.type === 'api-call') {
                    await this.syncApiCall(action);
                }
                
                console.log('Synced offline action:', action.id);
            } catch (error) {
                console.error('Failed to sync action:', action.id, error);
                failedActions.push(action);
            }
        }
        
        // Update queue with failed actions
        this.offlineQueue = failedActions;
        this.updateOfflineQueueIndicator();
        localStorage.setItem('offline-queue', JSON.stringify(this.offlineQueue));
        
        if (failedActions.length === 0) {
            this.showToast('All offline actions synchronized!', 'success');
        } else {
            this.showToast(`${failedActions.length} actions failed to sync`, 'warning');
        }
    }
    
    async syncGistCreation(action) {
        const response = await fetch('/api/v1/gists', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': window.CasGists?.csrf
            },
            body: JSON.stringify(action.data)
        });
        
        if (!response.ok) {
            throw new Error(`Gist creation failed: ${response.status}`);
        }
        
        return response.json();
    }
    
    async syncApiCall(action) {
        const response = await fetch(action.url, action.options);
        
        if (!response.ok) {
            throw new Error(`API call failed: ${response.status}`);
        }
        
        return response;
    }
    
    handleServiceWorkerUpdate(registration) {
        const newWorker = registration.installing;
        
        newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                // New version available
                this.showUpdateAvailable();
            }
        });
    }
    
    showUpdateAvailable() {
        const updateAlert = document.createElement('div');
        updateAlert.className = 'alert alert-info fixed top-4 left-1/2 transform -translate-x-1/2 z-50 shadow-lg max-w-sm';
        updateAlert.innerHTML = `
            <div>
                <h3 class="font-bold">Update Available</h3>
                <div class="text-sm">A new version of CasGists is ready.</div>
            </div>
            <div class="flex-none">
                <button class="btn btn-sm" onclick="this.parentElement.parentElement.remove()">Later</button>
                <button class="btn btn-sm btn-primary ml-2" onclick="window.casgistsPWA.applyUpdate(); this.parentElement.parentElement.remove()">Update</button>
            </div>
        `;
        
        document.body.appendChild(updateAlert);
    }
    
    applyUpdate() {
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.getRegistration().then(registration => {
                if (registration && registration.waiting) {
                    registration.waiting.postMessage({ type: 'SKIP_WAITING' });
                    window.location.reload();
                }
            });
        }
    }
    
    isRunningAsPWA() {
        return window.matchMedia('(display-mode: standalone)').matches ||
               window.navigator.standalone ||
               document.referrer.includes('android-app://');
    }
    
    // Utility functions
    urlBase64ToUint8Array(base64String) {
        const padding = '='.repeat((4 - base64String.length % 4) % 4);
        const base64 = (base64String + padding)
            .replace(/-/g, '+')
            .replace(/_/g, '/');
        
        const rawData = window.atob(base64);
        const outputArray = new Uint8Array(rawData.length);
        
        for (let i = 0; i < rawData.length; ++i) {
            outputArray[i] = rawData.charCodeAt(i);
        }
        
        return outputArray;
    }
    
    showToast(message, type = 'info') {
        // Create toast if toast container exists
        const container = document.getElementById('toast-container');
        if (container) {
            const toast = document.createElement('div');
            toast.className = `alert alert-${type} mb-2`;
            toast.innerHTML = `<span>${message}</span>`;
            
            container.appendChild(toast);
            
            setTimeout(() => {
                toast.remove();
            }, 5000);
        } else {
            // Fallback to console if no toast container
            console.log(`[PWA] ${type.toUpperCase()}: ${message}`);
        }
    }
}

// Initialize PWA when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.casgistsPWA = new CasGistsPWA();
});

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CasGistsPWA;
}