# AGENTS.md

## Scope
- This file applies to `admin-web/` only.

## Commands
- Install: `npm install`
- Dev: `npm run dev`
- Build: `npm run build`

## Env
- `VITE_API_BASE` default: `/api/v1`

## Backend Contract
- Auth: `Authorization: Bearer <token>` required for admin APIs.
- API response envelope: `{ code, msg, data }`.
- Admin pages require authenticated user with role `admin`.

## Verification
- Prefer `npm run build` as required check for frontend changes.
