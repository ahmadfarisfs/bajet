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
┌─────────────────────────────────────────────────────────┐
│  User's browser                                         │
│                                                         │
│  GitHub Pages  ──── static SPA (Svelte 5 + Vite PWA)   │
│       │                                                 │
│       │  1. Google Sign-In popup → ID token             │
│       │  2. API calls + "Authorization: Bearer <token>" │
│       ▼                                                 │
│  Koyeb (Docker)  ──── Go + Echo API server              │
│       │                                                 │
│       │  verifies token against Google tokeninfo        │
│       │  DATABASE_URL (postgres)                        │
│       ▼                                                 │
│  Neon  ──── serverless PostgreSQL                       │
│                                                         │
│  GCP  ──── Google OAuth 2.0 (issues ID tokens)         │
└─────────────────────────────────────────────────────────┘
```

### On every push to `main`
1. **GitHub Actions** builds the Svelte frontend with secrets baked in as `VITE_*` env vars
2. Built `dist/` is published to **GitHub Pages**
3. **Koyeb** detects the push and rebuilds the Docker image from `backend/Dockerfile`

---

## Stack

- **Backend**: Go + Echo + GORM + PostgreSQL (Neon)
- **Frontend**: Svelte 5 + Vite + PWA
- **Auth**: Google Sign-In (ID token, verified server-side)
- **Hosting**: GitHub Pages (frontend) + Koyeb free tier (backend)
- **Database**: Neon serverless PostgreSQL (free tier)

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
