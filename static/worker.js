const CacheName = "widgets-cache";

self.addEventListener("install", () => {
	self.skipWaiting();
});

self.addEventListener("activate", event => {
	event.waitUntil(clients.claim());
});

self.addEventListener("fetch", event => {
	if (event.request.method !== "GET" || new URL(event.request.url).origin !== self.location.origin) {
		return;
	}

	event.respondWith(
		caches.open(CacheName).then(async cache => {
			const cachedResponse = await cache.match(event.request);

			try {
				const networkResponse = await fetch(event.request);

				if (!networkResponse.ok) {
					return cachedResponse || networkResponse;
				}

				event.waitUntil(cache.put(event.request, networkResponse.clone()));

				return networkResponse;
			} catch {
				return cachedResponse || Response.error();
			}
		})
	);
});
