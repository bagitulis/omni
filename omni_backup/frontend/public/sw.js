const CACHE_VERSION = "v2";
const CACHE_NAME = `ecommerce-${CACHE_VERSION}`;

// Assets yang selalu di-cache (optional - won't fail if missing)
const STATIC_ASSETS = ["/", "/index.html"];

// Install event
self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(STATIC_ASSETS).catch(() => {
          console.log("Some optional assets could not be cached");
        });
      })
      .then(() => self.skipWaiting())
  );
});

// Activate event
self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((name) => {
            if (name !== CACHE_NAME) {
              return caches.delete(name);
            }
          })
        );
      })
      .then(() => self.clients.claim())
  );
});

// Fetch event
self.addEventListener("fetch", (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests
  if (request.method !== "GET") {
    return;
  }

  // Skip external URLs (only handle same-origin requests)
  if (url.origin !== self.location.origin) {
    return;
  }

  // Skip browser extension requests
  if (
    url.protocol === "chrome-extension:" ||
    url.protocol === "moz-extension:"
  ) {
    return;
  }

  // API requests - network first with cache fallback
  if (url.pathname.startsWith("/api/")) {
    event.respondWith(
      fetch(request)
        .then((response) => {
          if (response.ok) {
            const cloned = response.clone();
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(request, cloned);
            });
          }
          return response;
        })
        .catch(() => {
          return caches.match(request);
        })
    );
    return;
  }

  // Static assets - cache first
  if (url.pathname.match(/\.(js|css|png|jpg|jpeg|svg|gif|webp|woff|woff2)$/i)) {
    event.respondWith(
      caches.match(request).then((cached) => {
        return cached || fetch(request);
      })
    );
    return;
  }

  // Default - network first
  event.respondWith(
    fetch(request).catch(() => {
      return caches.match(request);
    })
  );
});
