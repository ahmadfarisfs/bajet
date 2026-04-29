# Bajet — Period Budgeting App

A minimal web app for budget check-ins by period, not by transaction.

## Concept

Users receive a fixed budget split across 4 periods per cycle.  
At the end of each period, they check in once with either:
- **Sisa** — money remaining (savings)
- **Defisit** — overspending

No transaction tracking. No categories. Just consistency.

## Stack

- **Backend**: Go + Echo + GORM + SQLite
- **Frontend**: Svelte 5 + Vite

## Development

### Backend

```bash
cd backend
go mod tidy
go run .          # starts on :8080
```

### Frontend

```bash
cd frontend
npm install
npm run dev       # dev server with API proxy to :8080
```

For production, run `npm run build` — the backend serves `frontend/dist/` as static files.

## Period Division Modes

| Mode | 30 days | 31 days |
|------|---------|---------|
| Equal | 8/8/7/7 | 8/8/8/7 |
| Behavioral | 9/7/7/7 | 9/8/7/7 |

**Behavioral** front-loads one extra day to P1 — reflecting real spending patterns where discipline is highest early in a cycle.

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/cycles` | List all cycles |
| POST | `/api/cycles` | Create cycle |
| GET | `/api/cycles/:id` | Get cycle with periods |
| DELETE | `/api/cycles/:id` | Delete cycle |
| POST | `/api/periods/:id/checkin` | Submit check-in |
| DELETE | `/api/periods/:id/checkin` | Undo check-in |
