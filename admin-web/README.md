# Admin Web

## Install

```bash
npm install
```

## Dev

```bash
npm run dev
```

Vite dev server proxies `/api/*` to `http://localhost:8080`.
Default proxy target can be overridden with `VITE_API_PROXY_TARGET` (for example `http://192.168.1.10:8080`).
Vite dev server listens on `0.0.0.0:5173`, so LAN devices can access `http://<server-ip>:5173`.

## Build

```bash
npm run build
```

Build output is `admin-web/dist` and can be served by Go backend at `/admin`.

## Notes

- All API requests include `Authorization: Bearer <token>` via axios interceptor.
- Backend CORS should allow admin origin in development.
