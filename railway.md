# Deploying PatientOS to Railway

This guide covers deploying both services — the Go API and the Next.js
frontend — to [Railway](https://railway.app), plus the managed Postgres
instance they share.

## Overview

A PatientOS deployment is three Railway services in one project:

| Service     | Source                | Notes                                   |
|-------------|------------------------|------------------------------------------|
| `postgres`  | Railway's Postgres plugin | Managed, provides `DATABASE_URL`     |
| `backend`   | `backend/` (Dockerfile)   | Go API + WebSocket server            |
| `frontend`  | `frontend/` (Nixpacks)    | Next.js app                          |

Railway can deploy straight from this GitHub repo — each service points
at the same repo with a different **root directory**.

## 1. Create the project

1. Go to [railway.app/new](https://railway.app/new) → **Deploy from GitHub repo**.
2. Select the `patient-os` repository.
3. Railway will create one service from the repo root — delete it once
   the project exists; you'll add the three services below explicitly so
   each gets the correct root directory.

## 2. Add Postgres

1. In the project canvas: **New** → **Database** → **Add PostgreSQL**.
2. Railway provisions the instance and exposes a `DATABASE_URL` variable
   on that service. You'll reference it from the backend service in step 3.

## 3. Deploy the backend (Go API)

1. **New** → **GitHub Repo** → pick `patient-os` again.
2. Open the new service's **Settings**:
   - **Root Directory**: `backend`
   - **Builder**: Dockerfile (Railway will detect `backend/Dockerfile` —
     `backend/railway.json` already pins this explicitly, so no manual
     config should be needed)
3. Open **Variables** and add:
   ```
   DATABASE_URL=${{Postgres.DATABASE_URL}}
   APP_ENV=production
   ```
   The `${{Postgres.DATABASE_URL}}` reference syntax pulls the value
   live from the Postgres service — use Railway's variable picker
   (click "Add Variable Reference") rather than typing it by hand.
4. Railway sets `PORT` automatically; `cmd/server/main.go` already reads
   it via `internal/config.Load()`, so no `PORT` variable is needed here.
5. Deploy. Once it's live, open **Settings → Networking** and generate a
   public domain (e.g. `patientos-api.up.railway.app`). The Go server
   runs GORM `AutoMigrate` on startup, so the schema is created
   automatically on first boot — no separate migration step.
6. Verify with:
   ```
   curl https://<your-backend-domain>/health
   # {"status":"ok","database":"ok"}
   ```

## 4. Deploy the frontend (Next.js)

1. **New** → **GitHub Repo** → pick `patient-os` again.
2. Open the new service's **Settings**:
   - **Root Directory**: `frontend`
   - **Builder**: Nixpacks (default — no Dockerfile needed for the frontend)
3. Open **Variables** and add:
   ```
   NEXT_PUBLIC_API_URL=https://<your-backend-domain>
   NEXT_PUBLIC_WS_URL=wss://<your-backend-domain>
   ```
   Use the public backend domain from step 3.5. `NEXT_PUBLIC_*` variables
   are baked in at build time, so set them **before** the first deploy —
   changing them later requires a redeploy, not just a restart.
4. Deploy, then generate a public domain under **Settings → Networking**.

## 5. Verify the end-to-end flow

```bash
# seed a hospital/department/doctor against the live backend
curl -X POST https://<backend-domain>/api/v1/hospitals \
  -H "Content-Type: application/json" \
  -d '{"name":"City General Hospital"}'
```

Then visit `https://<frontend-domain>/h/1` and confirm the hospital name
renders — that exercises the frontend → backend round trip in production.

## Environment variable reference

**Backend** (`backend/.env.example`)
| Variable       | Source                          | Required |
|----------------|----------------------------------|----------|
| `DATABASE_URL` | Reference to the Postgres service | Yes    |
| `PORT`         | Set automatically by Railway      | No     |
| `APP_ENV`      | `production`                      | Recommended |

**Frontend** (`frontend/.env.local.example`)
| Variable              | Value                                  | Required |
|------------------------|----------------------------------------|----------|
| `NEXT_PUBLIC_API_URL` | Public backend domain (`https://…`)    | Yes      |
| `NEXT_PUBLIC_WS_URL`  | Public backend domain (`wss://…`)      | Yes (once realtime/websocket lands) |

## Notes

- **Custom domains**: add them under each service's **Settings → Networking**
  once DNS is ready; Railway issues TLS automatically.
- **Redeploys**: pushes to the tracked branch (`main` by default) trigger
  automatic redeploys for both services.
- **Logs/metrics**: available per-service under the **Observability** tab —
  useful for watching GORM's migration log on first boot.
- **Rollbacks**: each deploy is kept as a snapshot under **Deployments** —
  use "Redeploy" on a prior one if a release regresses.
