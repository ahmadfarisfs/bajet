// Bajet keeps all data in this browser's localStorage — no server, no account.
// The methods return promises so the screens don't care where data lives.
//
// Because the browser is the only copy, Overview offers a JSON backup/restore,
// and storage is marked persistent where the browser allows it.

import { buildPeriods, calcBudgets, daysBetween, daysUntil, CARRY_MODES } from './utils.js'

const KEY = 'bajet_v1'
// Copy of the cycle list the old server-backed versions kept on each device.
const LEGACY_CACHE_KEY = 'bajet_cycles_cache'

export const api = buildLocalApi()

// Ask the browser not to clear our data under storage pressure (Safari otherwise
// may evict sites that aren't opened for a while). Best effort; ignored if unsupported.
try { navigator.storage?.persist?.() } catch { /* not available */ }

// ── Backup format & normalisation ────────────────────────────────────────────

const day = (v) => String(v ?? '').substring(0, 10)
const isDay = (v) => /^\d{4}-\d{2}-\d{2}$/.test(v)

// Cleans cycles from a backup file or the old server cache into the local
// shape: plain YYYY-MM-DD dates, numeric amounts, sequential ownership.
// Throws if the data isn't a list of Bajet cycles.
export function normalizeCycles(list) {
  if (!Array.isArray(list)) throw new Error('not a Bajet backup')
  return list.map((c) => {
    const periods = Array.isArray(c?.periods) ? c.periods : null
    if (!periods || !isDay(day(c.start_date)) || !isDay(day(c.end_date)) || !(Number(c.total_budget) > 0)) {
      throw new Error('not a Bajet backup')
    }
    const id = Number(c.id)
    return {
      id,
      start_date: day(c.start_date),
      end_date: day(c.end_date),
      total_budget: Number(c.total_budget),
      division_mode: ['equal', 'behavioral', 'menurun', 'progresif'].includes(c.division_mode) ? c.division_mode : 'equal',
      carry_mode: CARRY_MODES.includes(c.carry_mode) ? c.carry_mode : 'spread',
      num_periods: Number(c.num_periods) || periods.length || 4,
      created_at: c.created_at ?? new Date().toISOString(),
      periods: periods
        .map((p) => {
          const done = p.status === 'completed'
          return {
            id: Number(p.id),
            cycle_id: id,
            period_number: Number(p.period_number),
            start_date: day(p.start_date),
            end_date: day(p.end_date),
            budget: Number(p.budget),
            status: done ? 'completed' : 'open',
            result_type: done && p.result_type === 'defisit' ? 'defisit' : done ? 'sisa' : '',
            result_amount: done ? Number(p.result_amount) || 0 : 0,
            created_at: p.created_at ?? new Date().toISOString(),
          }
        })
        .sort((a, b) => a.period_number - b.period_number),
    }
  }).sort((a, b) => a.id - b.id)
}

function storeFrom(cycles) {
  const maxId = (xs) => xs.reduce((m, x) => Math.max(m, x.id || 0), 0)
  return {
    cycles,
    seq: { cycle: maxId(cycles), period: maxId(cycles.flatMap(c => c.periods)) },
  }
}

export function exportBackup() {
  const { cycles } = JSON.parse(localStorage.getItem(KEY) || 'null') ?? { cycles: [] }
  return JSON.stringify({ app: 'bajet', version: 1, exported_at: new Date().toISOString(), cycles }, null, 2)
}

// Replaces everything with the cycles in a backup file. Returns the cycle count.
export function restoreBackup(text) {
  let parsed
  try { parsed = JSON.parse(text) } catch { throw new Error('not a Bajet backup') }
  const cycles = normalizeCycles(Array.isArray(parsed) ? parsed : parsed?.cycles)
  localStorage.setItem(KEY, JSON.stringify(storeFrom(cycles)))
  return cycles.length
}

// One-time move for devices that used the server-backed app: if there's no
// local data yet but the old cycle cache exists, adopt it. Returns the count.
export function importLegacyCache() {
  try {
    if (localStorage.getItem(KEY)) return 0
    const cached = JSON.parse(localStorage.getItem(LEGACY_CACHE_KEY) || 'null')
    if (!Array.isArray(cached) || !cached.length) return 0
    const cycles = normalizeCycles(cached)
    localStorage.setItem(KEY, JSON.stringify(storeFrom(cycles)))
    localStorage.removeItem(LEGACY_CACHE_KEY)
    return cycles.length
  } catch {
    return 0
  }
}

// ── Local store ──────────────────────────────────────────────────────────────

function buildLocalApi() {
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

  // Validates and fills defaults. Returns [values, errorMessage].
  function validate({ start_date, end_date, total_budget, division_mode, carry_mode, num_periods }) {
    if (!start_date || !end_date) return [null, 'invalid dates']
    if (daysBetween(start_date, end_date) < 2) return [null, 'end_date must be after start_date']
    if (!(Number(total_budget) > 0)) return [null, 'total_budget must be greater than 0']
    const n = num_periods >= 1 && num_periods <= 12 ? Math.round(num_periods) : 4
    if (daysBetween(start_date, end_date) < n) return [null, 'date range is shorter than the number of periods']
    return [{ start_date, end_date, total_budget: Number(total_budget), division_mode: division_mode || 'equal',
      carry_mode: CARRY_MODES.includes(carry_mode) ? carry_mode : 'spread', num_periods: n }, '']
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
