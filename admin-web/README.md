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

## Build

```bash
npm run build
```

Build output is `admin-web/dist` and can be served by Go backend at `/admin`.

## Notes

- All API requests include `Authorization: Bearer <token>` via axios interceptor.
- Backend CORS should allow admin origin in development.
