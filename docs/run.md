# Run Guide

## One-command start (Linux/macOS)

```bash
bash scripts/dev-up.sh
```

This script will:

1. start `postgres` and `redis` via `docker compose`
2. wait for Postgres readiness
3. apply all `migrations/*.up.sql`
4. start:
   - `go run main.go -mode server`
   - `go run main.go -mode worker`
5. write logs and PID files into `.run/`

You can also specify a custom env file path:

```bash
ENV_FILE=/absolute/path/to/.env bash scripts/dev-up.sh
```

## Stop all

```bash
bash scripts/dev-down.sh
```

This stops:

1. local `server` and `worker` processes (from `.run/*.pid`)
2. `postgres` and `redis` containers

## Logs

- server: `.run/server.log`
- worker: `.run/worker.log`

## Notes

- If `.env` does not exist, `dev-up.sh` copies from `.env.example`.
- Required commands: `docker` (with compose plugin) and `go`.
- `main.go` will prioritize `ENV_FILE` when loading environment variables.
