// If VITE_API_URL is set at build time, use the real HTTP backend.
// Otherwise fall back to localStorage so the app works on GitHub Pages
// without any server.

import { getToken, signIn, signOut } from './auth.js'
import { startSync, endSync, sessionExpired } from './sync.js'
import { buildPeriods, calcBudgets, daysBetween, daysUntil } from './utils.js'

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
    updateCycle: (id, data) => req('PUT',    `/api/cycles/${id}`, data),
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

  function buildLocalPeriods(cycleId, startDate, endDate, totalBudget, mode, seq, n) {
    const now = new Date().toISOString()
    return buildPeriods(startDate, endDate, Number(totalBudget), mode, n).map(p => ({
      id: ++seq.period, cycle_id: cycleId, period_number: p.period_number,
      start_date: p.start_date, end_date: p.end_date, budget: p.budget, status: 'open',
      result_type: '', result_amount: 0, created_at: now,
    }))
  }

  // Same rules as the backend's parseCycleRequest. Returns [values, errorMessage].
  function validate({ start_date, end_date, total_budget, division_mode, num_periods }) {
    if (!start_date || !end_date) return [null, 'invalid dates']
    if (daysBetween(start_date, end_date) < 2) return [null, 'end_date must be after start_date']
    if (!(Number(total_budget) > 0)) return [null, 'total_budget must be greater than 0']
    const n = num_periods >= 1 && num_periods <= 12 ? Math.round(num_periods) : 4
    if (daysBetween(start_date, end_date) < n) return [null, 'date range is shorter than the number of periods']
    return [{ start_date, end_date, total_budget: Number(total_budget), division_mode: division_mode || 'equal', num_periods: n }, '']
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

    createCycle(input) {
      const [v, err] = validate(input)
      if (err) return fail(err)
      const data = load()
      data.seq.cycle  = (data.seq.cycle  || 0) + 1
      data.seq.period = (data.seq.period || 0)
      const cycle = {
        id: data.seq.cycle, ...v, created_at: new Date().toISOString(),
        periods: buildLocalPeriods(data.seq.cycle, v.start_date, v.end_date, v.total_budget, v.division_mode, data.seq, v.num_periods),
      }
      data.cycles.push(cycle)
      persist(data)
      return ok(cycle)
    },

    updateCycle(id, input) {
      const [v, err] = validate(input)
      if (err) return fail(err)
      const data = load()
      const cycle = data.cycles.find(c => c.id === Number(id))
      if (!cycle) return fail('cycle not found')
      const hasCheckIns = cycle.periods.some(p => p.status === 'completed')
      const reshape = v.start_date !== cycle.start_date || v.end_date !== cycle.end_date ||
        v.num_periods !== cycle.periods.length
      if (reshape && hasCheckIns) {
        return fail("dates and number of periods can't change after a check-in; undo the check-ins first")
      }
      Object.assign(cycle, v)
      if (hasCheckIns) {
        const budgets = calcBudgets(v.total_budget, cycle.periods.length, v.division_mode)
        cycle.periods.forEach((p, i) => { p.budget = budgets[i] })
      } else {
        cycle.periods = buildLocalPeriods(cycle.id, v.start_date, v.end_date, v.total_budget, v.division_mode, data.seq, v.num_periods)
      }
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
          if (daysUntil(p.start_date) > 0) return fail('period has not started yet')
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
