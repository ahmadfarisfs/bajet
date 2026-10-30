# Bajet — Period Budgeting App

A minimal web app for budget check-ins by period, not by transaction.

## Concept

Users receive a fixed budget split across 4 periods per cycle.  
At the end of each period, they check in once with either:
- **Sisa** — money remaining (savings)
- **Defisit** — overspending

No transaction tracking. No categories. Just consistency.

---

## How data is stored

Bajet runs entirely in the browser. Everything is saved in the browser's
`localStorage` on your device: no server, no database, no sign-in.

- **Instant and offline**: every read and write happens on the device.
- **Private**: your numbers never leave your phone.
- **Back it up**: because this device holds the only copy, clearing browser
  data or losing the phone loses it. Overview → *Your data* has **Backup**
  (downloads a JSON file), **Restore** (loads one, e.g. on a new phone) and a
  **CSV** export for spreadsheets. The app also asks the browser to keep its
  storage persistent so it isn't cleared under storage pressure.
- **Coming from the server version**: the old app kept a copy of your cycles in
  the browser. On first start, Bajet adopts that copy automatically, so
  existing data carries over on any device where you used it.

## Stack

- Svelte 5 + Vite, installable PWA
- Hosted on GitHub Pages; every push to `main` builds and deploys

## Development

```bash
cd frontend
npm install
npm run dev      # http://localhost:5173
npm run build    # production build in frontend/dist
```

## Deployment

GitHub Actions (`.github/workflows/deploy.yml`) builds the frontend and
publishes `frontend/dist` to GitHub Pages on every push to `main`. In the repo,
**Settings → Pages → Source** must be *GitHub Actions*. No secrets are needed.

## Period Division Modes

| Mode | 30 days | 31 days |
|------|---------|---------|
| Equal | 8/8/7/7 | 8/8/8/7 |
| Behavioral | 9/7/7/7 | 9/8/7/7 |

**Behavioral** front-loads one extra day to P1 — reflecting real spending patterns where discipline is highest early in a cycle.

---
