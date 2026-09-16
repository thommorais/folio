Object.assign(process.env, {
	NODE_ENV: 'test',
	PUBLIC_WEBAPP_URL: 'https://localhost:3000',
	PUBLIC_SITE_NAME: 'thom',
	PUBLIC_VAPID_PUBLIC_KEY: 'test-public-key',
	PUBLIC_API_URL: 'http://127.0.0.1:8090',
	WEBAPP_URL: 'https://localhost:3000',
	VAPID_PRIVATE_KEY: 'test-private-key',
	API_URL: 'http://127.0.0.1:8090',
})
