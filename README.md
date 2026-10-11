# Bajet — Period Budgeting App

A minimal web app for budget check-ins by period, not by transaction.

## Concept

Users receive a fixed budget split across 4 periods per cycle.  
At the end of each period, they check in once with either:
- **Sisa** — money remaining (savings)
- **Defisit** — overspending

No transaction tracking. No categories. Just consistency.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  User's browser                                              │
│                                                              │
│  GitHub Pages ──── static SPA (Svelte 5 + Vite PWA)          │
│       │                                                      │
│       │ 1. Google Sign-In button → ID token                  │
│       │ 2. supabase.auth.signInWithIdToken → Supabase session│
│       │ 3. reads (RLS) + writes (SQL functions) via supabase-js
│       ▼                                                      │
│  Supabase "Bajet" (ap-southeast-1) ──── Postgres + Auth      │
│     - row level security: each user reads only their data    │
│     - create_cycle / update_cycle / check_in / undo_check_in │
│       enforce the business rules in one transaction          │
└──────────────────────────────────────────────────────────────┘
```

Same approach as the Perut repo: no server of our own, so nothing sleeps and
there is no cold start. The schema, security rules and functions live in
[`supabase/schema.sql`](supabase/schema.sql).

The old Go backend (`backend/`, Koyeb + Neon) still works and is used until the
Supabase variables below are set; see **Switching to Supabase**.

### On every push to `main`
1. **GitHub Actions** builds the Svelte frontend (Supabase URL/key from repo variables)
2. Built `dist/` is published to **GitHub Pages**

## Switching to Supabase

The Supabase project and schema already exist. To move over:

1. **Enable Google sign-in in Supabase.** Dashboard → Authentication → Sign In /
   Providers → Google → enable, then paste the existing Google OAuth **Client
   ID** (the `VITE_GOOGLE_CLIENT_ID` secret) and its **Client secret** from
   Google Cloud Console. Leave "Skip nonce check" off.
2. **Copy the data from Neon.** Add two repository secrets: `NEON_DATABASE_URL`
   (the old `DATABASE_URL`) and `SUPABASE_DB_URL` (Supabase → Connect →
   *Session pooler* string with the DB password; not the direct host, which is
   IPv6-only). Then Actions → **Migrate Neon → Supabase** → Run workflow. It is
   safe to run again. Each person's cycles attach to them the first time they
   sign in to the new app.
3. **Point the app at Supabase.** Add repository *variables* (not secrets):
   `SUPABASE_URL` = `https://ycjbdfdxuqcfyhiwcadn.supabase.co` and
   `SUPABASE_ANON_KEY` = the project's publishable key
   (`sb_publishable_…`, public by design). Re-run the Pages deploy.
4. When everything looks right, delete the Koyeb service and the Neon project.

## Stack

- **Frontend**: Svelte 5 + Vite + PWA, talking to Supabase with supabase-js
- **Database**: Supabase Postgres (free tier), schema in `supabase/schema.sql`
- **Auth**: Google Sign-In → Supabase Auth (`signInWithIdToken`, with nonce)
- **Hosting**: GitHub Pages
- **Legacy** (until the switch): Go + Echo + GORM backend on Koyeb with Neon Postgres

## Development

### Option A — Docker (recommended)

Requires [Docker Desktop](https://www.docker.com/products/docker-desktop/).

```bash
# 1. Copy and fill the env file (same Client ID as production)
cp .env.example .env

# 2. Start backend + frontend together
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend (Svelte dev server) | http://localhost:5173 |
| Backend API | http://localhost:8080 |

Data is persisted in a Docker volume (`bajet-data`) using SQLite — no database setup needed locally.

> Make sure `http://localhost:5173` is added to **Authorized JavaScript origins** in your GCP OAuth client.

---

### Option B — Without Docker

**Backend**
```bash
cd backend
go mod tidy
go run .          # starts on :8080, uses bajet.db (SQLite) when DATABASE_URL is unset
```

**Frontend**
```bash
cd frontend
npm install
npm run dev       # no VITE_API_URL set → uses localStorage only
```

To connect the frontend to your local backend, create `frontend/.env.local`:
```
VITE_API_URL=http://localhost:8080
VITE_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
```

---

## Deployment Setup

### 1 — GCP (Google Cloud Console)

1. Go to [console.cloud.google.com](https://console.cloud.google.com) → **APIs & Services → Credentials**
2. **Create Credentials → OAuth 2.0 Client ID** → Application type: **Web application**
3. Under **Authorized JavaScript origins** add:
   - `https://<your-github-username>.github.io`
   - Your Koyeb service URL (e.g. `https://bajet-xxxx.koyeb.app`)
   - `http://localhost:5173` (local dev)
4. Copy the **Client ID** — used in both Koyeb env vars and GitHub secrets

---

### 2 — Neon (PostgreSQL)

1. Go to [neon.tech](https://neon.tech) → **New project**
2. Copy the **Connection string** from **Connection Details → URI**:
   ```
   postgresql://user:password@ep-xxx.region.aws.neon.tech/dbname?sslmode=require
   ```
3. Used as `DATABASE_URL` in Koyeb

---

### 3 — Koyeb (backend)

1. Go to [app.koyeb.com](https://app.koyeb.com) → **Create Service → GitHub**
2. Select your repo, branch `main`
3. **Builder**: Dockerfile
   - **Dockerfile location** (Override ON): `Dockerfile`
   - **Work directory** (Override ON): `backend`
4. **Exposed port**: `8080`
5. **Environment variables**:

| Key | Value | Where to get it |
|---|---|---|
| `DATABASE_URL` | Neon connection string | Neon dashboard |
| `GOOGLE_CLIENT_ID` | OAuth Client ID | GCP Credentials |
| `PORT` | `8080` | literal value |

6. Deploy → copy the service URL once healthy (e.g. `https://bajet-xxxx.koyeb.app`)

---

### 4 — GitHub

#### Repository secrets
Go to **Settings → Secrets and variables → Actions → New repository secret**:

| Secret name | Value | Where to get it |
|---|---|---|
| `BACKEND_URL` | Koyeb service URL (no trailing slash) | Koyeb service page |
| `VITE_GOOGLE_CLIENT_ID` | OAuth Client ID | GCP Credentials |

#### GitHub Pages
Go to **Settings → Pages**:
- Source: **GitHub Actions**

That's it — on every push to `main` the workflow builds and deploys automatically.

---

## Environment variable reference

### Backend (Koyeb)

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes (production) | Postgres connection string. If unset, falls back to SQLite at `DB_PATH` |
| `GOOGLE_CLIENT_ID` | Yes | GCP OAuth 2.0 Client ID — used to verify ID tokens |
| `PORT` | No | HTTP port (default: `8080`) |
| `DB_PATH` | No | SQLite file path when `DATABASE_URL` is unset (default: `bajet.db`) |

### Frontend build (GitHub Actions secrets)

| Variable | Required | Description |
|---|---|---|
| `BACKEND_URL` | No | If set, frontend calls real backend. If empty, uses localStorage |
| `VITE_GOOGLE_CLIENT_ID` | Yes (if `BACKEND_URL` set) | GCP OAuth 2.0 Client ID baked into the frontend build |

---

## Period Division Modes

| Mode | 30 days | 31 days |
|------|---------|---------|
| Equal | 8/8/7/7 | 8/8/8/7 |
| Behavioral | 9/7/7/7 | 9/8/7/7 |

**Behavioral** front-loads one extra day to P1 — reflecting real spending patterns where discipline is highest early in a cycle.

---

## API

All endpoints require `Authorization: Bearer <google-id-token>`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/cycles` | List cycles for authenticated user |
| POST | `/api/cycles` | Create cycle |
| GET | `/api/cycles/:id` | Get cycle with periods |
| DELETE | `/api/cycles/:id` | Delete cycle |
| POST | `/api/periods/:id/checkin` | Submit check-in |
| DELETE | `/api/periods/:id/checkin` | Undo check-in |
