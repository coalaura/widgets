const CacheName = "widgets-cache";

self.addEventListener("install", () => {
	self.skipWaiting();
});

self.addEventListener("activate", event => {
	event.waitUntil(clients.claim());
});

self.addEventListener("fetch", event => {
	event.respondWith(
		caches.open(CacheName).then(async cache => {
			let fetchedResponse;

			const cachedResponse = await cache.match(event.request);

			try {
				const networkResponse = await fetch(event.request);

				if (!networkResponse.ok) {
					throw new Error(networkResponse.statusText);
				}

				fetchedResponse = networkResponse;

				cache.put(event.request, networkResponse.clone());
			} catch {
				fetchedResponse = cachedResponse;
			}

			return fetchedResponse;
		})
	);
});
