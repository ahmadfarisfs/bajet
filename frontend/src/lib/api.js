// If VITE_API_URL is set at build time, use the real HTTP backend.
// Otherwise fall back to localStorage so the app works on GitHub Pages
// without any server.

import { getToken, signIn, signOut } from './auth.js'
import { startSync, endSync, sessionExpired } from './sync.js'

const API_URL = import.meta.env.VITE_API_URL

export const api = API_URL ? buildHttpApi(API_URL) : buildLocalApi()

// ── HTTP API (real backend) ──────────────────────────────────────────────────

function buildHttpApi(base) {
  async function req(method, path, body) {
    startSync()
    try {
      const token = getToken()
      const headers = {}
      if (body)  headers['Content-Type']  = 'application/json'
      if (token) headers['Authorization'] = `Bearer ${token}`
      const res = await fetch(`${base}${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      })
      if (res.status === 401) {
        signOut()
        sessionExpired.set(true)
        throw new Error('session_expired')
      }
      // Sliding session: backend sends a fresh token when the current one is
      // within 7 days of expiry. Store it silently — user stays signed in.
      const refreshed = res.headers.get('X-Refresh-Token')
      if (refreshed) signIn(refreshed)
      const data = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(data.error || `Request failed (${res.status})`)
      return data
    } finally {
      endSync()
    }
  }

  return {
    getCycles:   ()         => req('GET',    '/api/cycles'),
    getCycle:    (id)       => req('GET',    `/api/cycles/${id}`),
    createCycle: (data)     => req('POST',   '/api/cycles', data),
    deleteCycle: (id)       => req('DELETE', `/api/cycles/${id}`),
    checkIn:     (id, data) => req('POST',   `/api/periods/${id}/checkin`, data),
    undoCheckIn: (id)       => req('DELETE', `/api/periods/${id}/checkin`),
  }
}

// ── localStorage API (client-side fallback) ──────────────────────────────────

function buildLocalApi() {
  const KEY = 'bajet_v1'

  function load() {
    try {
      return JSON.parse(localStorage.getItem(KEY) || 'null') ||
        { cycles: [], seq: { cycle: 0, period: 0 } }
    } catch {
      return { cycles: [], seq: { cycle: 0, period: 0 } }
    }
  }

  function persist(data) {
    localStorage.setItem(KEY, JSON.stringify(data))
  }

  function calcDistribution(totalDays, n, mode) {
    const base = Math.floor(totalDays / n)
    const extra = totalDays % n
    const dist = Array(n).fill(base)
    for (let i = 0; i < extra; i++) dist[i]++
    if (mode === 'behavioral') {
      if (extra >= 2) { dist[0]++; dist[extra - 1]-- }
      else            { dist[0]++; dist[n - 1]-- }
    }
    return dist
  }

  function calcBudgets(totalBudget, n, mode) {
    const budgets = Array(n).fill(0)
    if (mode === 'progresif') {
      const tw = n * (n + 1) / 2
      let sum = 0
      for (let i = 0; i < n - 1; i++) {
        const b = Math.floor(totalBudget * (i + 1) / tw)
        budgets[i] = b; sum += b
      }
      budgets[n - 1] = totalBudget - sum
    } else if (mode === 'menurun') {
      const tw = n * (n + 1) / 2
      let sum = 0
      for (let i = 0; i < n - 1; i++) {
        const b = Math.floor(totalBudget * (n - i) / tw)
        budgets[i] = b; sum += b
      }
      budgets[n - 1] = totalBudget - sum
    } else {
      const base = Math.round(totalBudget / n)
      budgets.fill(base)
    }
    return budgets
  }

  function shiftDate(dateStr, n) {
    const [y, m, d] = dateStr.split('-').map(Number)
    const dt = new Date(y, m - 1, d + n)
    return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
  }

  function daysBetween(start, end) {
    const [sy, sm, sd] = start.split('-').map(Number)
    const [ey, em, ed] = end.split('-').map(Number)
    return Math.round((new Date(ey, em-1, ed) - new Date(sy, sm-1, sd)) / 86400000) + 1
  }

  function buildPeriods(cycleId, startDate, endDate, totalBudget, mode, seq, numPeriods) {
    const n = Math.min(Math.max(numPeriods || 4, 1), 12)
    const dist = calcDistribution(daysBetween(startDate, endDate), n, mode)
    const budgets = calcBudgets(totalBudget, n, mode)
    const now = new Date().toISOString()
    const periods = []
    let cur = startDate
    for (let i = 0; i < n; i++) {
      seq.period++
      const pEnd = shiftDate(cur, dist[i] - 1)
      periods.push({ id: seq.period, cycle_id: cycleId, period_number: i + 1,
        start_date: cur, end_date: pEnd, budget: budgets[i], status: 'open',
        result_type: '', result_amount: 0, created_at: now })
      cur = shiftDate(pEnd, 1)
    }
    return periods
  }

  const ok   = (v)   => Promise.resolve(v)
  const fail = (msg) => Promise.reject(new Error(msg))

  return {
    getCycles() {
      return ok([...load().cycles].reverse())
    },

    getCycle(id) {
      const c = load().cycles.find(c => c.id === Number(id))
      return c ? ok(c) : fail('cycle not found')
    },

    createCycle({ start_date, end_date, total_budget, division_mode, num_periods }) {
      const data = load()
      data.seq.cycle  = (data.seq.cycle  || 0) + 1
      data.seq.period = (data.seq.period || 0)
      const mode = division_mode || 'equal'
      const n    = Math.min(Math.max(num_periods || 4, 1), 12)
      const now  = new Date().toISOString()
      const cycle = {
        id: data.seq.cycle, start_date, end_date,
        total_budget: Number(total_budget), division_mode: mode,
        num_periods: n, created_at: now,
        periods: buildPeriods(data.seq.cycle, start_date, end_date, total_budget, mode, data.seq, n),
      }
      data.cycles.push(cycle)
      persist(data)
      return ok(cycle)
    },

    deleteCycle(id) {
      const data = load()
      data.cycles = data.cycles.filter(c => c.id !== Number(id))
      persist(data)
      return ok({ message: 'deleted' })
    },

    checkIn(periodId, { result_type, result_amount }) {
      const data = load()
      for (const cycle of data.cycles) {
        const p = cycle.periods.find(p => p.id === Number(periodId))
        if (p) {
          if (p.status === 'completed') return fail('period already completed')
          p.status = 'completed'; p.result_type = result_type
          p.result_amount = Number(result_amount)
          persist(data); return ok(p)
        }
      }
      return fail('period not found')
    },

    undoCheckIn(periodId) {
      const data = load()
      for (const cycle of data.cycles) {
        const p = cycle.periods.find(p => p.id === Number(periodId))
        if (p) {
          p.status = 'open'; p.result_type = ''; p.result_amount = 0
          persist(data); return ok(p)
        }
      }
      return fail('period not found')
    },
  }
}
