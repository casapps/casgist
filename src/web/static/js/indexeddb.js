// IndexedDB wrapper for offline data storage
class CasGistsDB {
    constructor() {
        this.dbName = 'CasGistsDB';
        this.version = 1;
        this.db = null;
        this.stores = {
            gists: 'gists',
            users: 'users',
            settings: 'settings',
            queue: 'offline-queue',
            cache: 'api-cache'
        };
    }

    async init() {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open(this.dbName, this.version);

            request.onerror = () => {
                console.error('IndexedDB error:', request.error);
                reject(request.error);
            };

            request.onsuccess = () => {
                this.db = request.result;
                console.log('IndexedDB initialized successfully');
                resolve(this.db);
            };

            request.onupgradeneeded = (event) => {
                const db = event.target.result;
                this.createStores(db);
            };
        });
    }

    createStores(db) {
        // Gists store - for offline gist data
        if (!db.objectStoreNames.contains(this.stores.gists)) {
            const gistStore = db.createObjectStore(this.stores.gists, { keyPath: 'id' });
            gistStore.createIndex('userId', 'userId', { unique: false });
            gistStore.createIndex('visibility', 'visibility', { unique: false });
            gistStore.createIndex('language', 'language', { unique: false });
            gistStore.createIndex('createdAt', 'createdAt', { unique: false });
            gistStore.createIndex('updatedAt', 'updatedAt', { unique: false });
        }

        // Users store - for user profile data
        if (!db.objectStoreNames.contains(this.stores.users)) {
            const userStore = db.createObjectStore(this.stores.users, { keyPath: 'id' });
            userStore.createIndex('username', 'username', { unique: true });
            userStore.createIndex('email', 'email', { unique: true });
        }

        // Settings store - for app settings and preferences
        if (!db.objectStoreNames.contains(this.stores.settings)) {
            const settingsStore = db.createObjectStore(this.stores.settings, { keyPath: 'key' });
        }

        // Queue store - for offline action queue
        if (!db.objectStoreNames.contains(this.stores.queue)) {
            const queueStore = db.createObjectStore(this.stores.queue, { keyPath: 'id' });
            queueStore.createIndex('type', 'type', { unique: false });
            queueStore.createIndex('timestamp', 'timestamp', { unique: false });
            queueStore.createIndex('priority', 'priority', { unique: false });
        }

        // Cache store - for API response caching
        if (!db.objectStoreNames.contains(this.stores.cache)) {
            const cacheStore = db.createObjectStore(this.stores.cache, { keyPath: 'url' });
            cacheStore.createIndex('timestamp', 'timestamp', { unique: false });
            cacheStore.createIndex('expiry', 'expiry', { unique: false });
        }
    }

    async transaction(storeNames, mode = 'readonly') {
        if (!this.db) {
            await this.init();
        }
        return this.db.transaction(storeNames, mode);
    }

    // Generic CRUD operations
    async get(storeName, key) {
        const tx = await this.transaction([storeName], 'readonly');
        const store = tx.objectStore(storeName);
        const request = store.get(key);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async getAll(storeName, query = null, count = null) {
        const tx = await this.transaction([storeName], 'readonly');
        const store = tx.objectStore(storeName);
        const request = store.getAll(query, count);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async put(storeName, data) {
        const tx = await this.transaction([storeName], 'readwrite');
        const store = tx.objectStore(storeName);
        const request = store.put(data);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async add(storeName, data) {
        const tx = await this.transaction([storeName], 'readwrite');
        const store = tx.objectStore(storeName);
        const request = store.add(data);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async delete(storeName, key) {
        const tx = await this.transaction([storeName], 'readwrite');
        const store = tx.objectStore(storeName);
        const request = store.delete(key);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async clear(storeName) {
        const tx = await this.transaction([storeName], 'readwrite');
        const store = tx.objectStore(storeName);
        const request = store.clear();

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    // Gist-specific operations
    async saveGist(gist) {
        gist.cachedAt = Date.now();
        return this.put(this.stores.gists, gist);
    }

    async getGist(gistId) {
        return this.get(this.stores.gists, gistId);
    }

    async getUserGists(userId, limit = null) {
        const tx = await this.transaction([this.stores.gists], 'readonly');
        const store = tx.objectStore(this.stores.gists);
        const index = store.index('userId');
        const request = index.getAll(userId, limit);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => {
                const results = request.result.sort((a, b) => 
                    new Date(b.updatedAt) - new Date(a.updatedAt)
                );
                resolve(results);
            };
            request.onerror = () => reject(request.error);
        });
    }

    async searchGists(query, filters = {}) {
        const allGists = await this.getAll(this.stores.gists);
        const lowerQuery = query.toLowerCase();

        return allGists.filter(gist => {
            // Text search in title, description, and content
            const searchText = `${gist.title} ${gist.description} ${gist.files?.map(f => f.content || '').join(' ')}`.toLowerCase();
            const matchesQuery = !query || searchText.includes(lowerQuery);

            // Apply filters
            const matchesVisibility = !filters.visibility || gist.visibility === filters.visibility;
            const matchesLanguage = !filters.language || gist.language === filters.language;
            const matchesUser = !filters.userId || gist.userId === filters.userId;

            return matchesQuery && matchesVisibility && matchesLanguage && matchesUser;
        });
    }

    // User operations
    async saveUser(user) {
        user.cachedAt = Date.now();
        return this.put(this.stores.users, user);
    }

    async getUser(userId) {
        return this.get(this.stores.users, userId);
    }

    async getUserByUsername(username) {
        const tx = await this.transaction([this.stores.users], 'readonly');
        const store = tx.objectStore(this.stores.users);
        const index = store.index('username');
        const request = index.get(username);

        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    // Settings operations
    async saveSetting(key, value) {
        return this.put(this.stores.settings, { key, value, timestamp: Date.now() });
    }

    async getSetting(key) {
        const result = await this.get(this.stores.settings, key);
        return result ? result.value : null;
    }

    async getAllSettings() {
        const settings = await this.getAll(this.stores.settings);
        const result = {};
        settings.forEach(setting => {
            result[setting.key] = setting.value;
        });
        return result;
    }

    // Queue operations for offline actions
    async queueAction(action) {
        action.id = action.id || `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        action.timestamp = action.timestamp || Date.now();
        action.priority = action.priority || 1;
        action.attempts = action.attempts || 0;
        action.status = 'pending';

        return this.put(this.stores.queue, action);
    }

    async getQueuedActions(type = null) {
        if (type) {
            const tx = await this.transaction([this.stores.queue], 'readonly');
            const store = tx.objectStore(this.stores.queue);
            const index = store.index('type');
            const request = index.getAll(type);

            return new Promise((resolve, reject) => {
                request.onsuccess = () => resolve(request.result);
                request.onerror = () => reject(request.error);
            });
        } else {
            return this.getAll(this.stores.queue);
        }
    }

    async removeQueuedAction(actionId) {
        return this.delete(this.stores.queue, actionId);
    }

    async updateQueuedAction(actionId, updates) {
        const action = await this.get(this.stores.queue, actionId);
        if (action) {
            Object.assign(action, updates);
            return this.put(this.stores.queue, action);
        }
        return null;
    }

    // Cache operations for API responses
    async cacheApiResponse(url, response, ttl = 300000) { // 5 minutes default TTL
        const cacheEntry = {
            url,
            data: response,
            timestamp: Date.now(),
            expiry: Date.now() + ttl
        };

        return this.put(this.stores.cache, cacheEntry);
    }

    async getCachedApiResponse(url) {
        const entry = await this.get(this.stores.cache, url);
        
        if (!entry) return null;

        // Check if expired
        if (Date.now() > entry.expiry) {
            await this.delete(this.stores.cache, url);
            return null;
        }

        return entry.data;
    }

    async clearExpiredCache() {
        const allCache = await this.getAll(this.stores.cache);
        const now = Date.now();
        const expired = allCache.filter(entry => now > entry.expiry);

        for (const entry of expired) {
            await this.delete(this.stores.cache, entry.url);
        }

        console.log(`[IndexedDB] Cleared ${expired.length} expired cache entries`);
    }

    // Cleanup operations
    async cleanup() {
        console.log('[IndexedDB] Running cleanup...');

        // Clear expired cache entries
        await this.clearExpiredCache();

        // Remove old gists (keep last 100)
        const gists = await this.getAll(this.stores.gists);
        if (gists.length > 100) {
            gists.sort((a, b) => (b.cachedAt || 0) - (a.cachedAt || 0));
            const toDelete = gists.slice(100);
            
            for (const gist of toDelete) {
                await this.delete(this.stores.gists, gist.id);
            }
            
            console.log(`[IndexedDB] Removed ${toDelete.length} old gists`);
        }

        // Remove failed queue actions older than 7 days
        const queue = await this.getAll(this.stores.queue);
        const weekAgo = Date.now() - (7 * 24 * 60 * 60 * 1000);
        const oldFailedActions = queue.filter(action => 
            action.status === 'failed' && action.timestamp < weekAgo
        );

        for (const action of oldFailedActions) {
            await this.delete(this.stores.queue, action.id);
        }

        if (oldFailedActions.length > 0) {
            console.log(`[IndexedDB] Removed ${oldFailedActions.length} old failed actions`);
        }
    }

    // Import/Export for backup
    async exportData() {
        const data = {};
        
        for (const [key, storeName] of Object.entries(this.stores)) {
            data[key] = await this.getAll(storeName);
        }

        return data;
    }

    async importData(data, clearFirst = false) {
        if (clearFirst) {
            for (const storeName of Object.values(this.stores)) {
                await this.clear(storeName);
            }
        }

        for (const [key, items] of Object.entries(data)) {
            if (this.stores[key] && Array.isArray(items)) {
                for (const item of items) {
                    await this.put(this.stores[key], item);
                }
            }
        }

        console.log('[IndexedDB] Data import completed');
    }

    // Get storage usage info
    async getStorageInfo() {
        const info = {};

        for (const [key, storeName] of Object.entries(this.stores)) {
            const items = await this.getAll(storeName);
            info[key] = {
                count: items.length,
                size: this.estimateSize(items)
            };
        }

        return info;
    }

    estimateSize(data) {
        return new Blob([JSON.stringify(data)]).size;
    }

    // Close database connection
    close() {
        if (this.db) {
            this.db.close();
            this.db = null;
        }
    }
}

// Create global instance
const casGistsDB = new CasGistsDB();

// Initialize when DOM is ready
if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', async () => {
        try {
            await casGistsDB.init();
            
            // Run periodic cleanup every hour
            setInterval(() => {
                casGistsDB.cleanup();
            }, 60 * 60 * 1000);
            
        } catch (error) {
            console.error('Failed to initialize IndexedDB:', error);
        }
    });
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { CasGistsDB, casGistsDB };
}

if (typeof window !== 'undefined') {
    window.casGistsDB = casGistsDB;
}