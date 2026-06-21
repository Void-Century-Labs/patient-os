# PatientOS

Hospital queue and patient navigation platform. Patients join a queue digitally,
track wait time in real time, and get notified when it's their turn — no
standing in a physical line.

## Structure

- `frontend/` — Next.js (TypeScript, TailwindCSS, shadcn/ui, TanStack Query/Table)
- `backend/` — Go (Gin, GORM, PostgreSQL)

## Local development

### Backend

```bash
cd backend
cp .env.example .env   # adjust DATABASE_URL if needed
go run ./cmd/server
```

Or run everything (Postgres + API) with Docker:

```bash
docker compose up --build
```

### Frontend

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

## Deployment

- Backend: `backend/Dockerfile`, `backend/fly.toml` (Fly.io), `backend/railway.json` (Railway)
- CI: `.github/workflows/ci.yml` runs build/lint/test on both apps
