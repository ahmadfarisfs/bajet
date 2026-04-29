// Client-side data layer using localStorage — no backend required.
// Keeps the same async interface as the original HTTP API so components are unchanged.

const KEY = 'bajet_v1'

function load() {
  try {
    return JSON.parse(localStorage.getItem(KEY) || 'null') || { cycles: [], seq: { cycle: 0, period: 0 } }
  } catch {
    return { cycles: [], seq: { cycle: 0, period: 0 } }
  }
}

function persist(data) {
  localStorage.setItem(KEY, JSON.stringify(data))
}

function calcDistribution(totalDays, mode) {
  const base = Math.floor(totalDays / 4)
  const extra = totalDays % 4
  const dist = [base, base, base, base]
  for (let i = 0; i < extra; i++) dist[i]++
  if (mode === 'behavioral') {
    if (extra >= 2) {
      dist[0]++
      dist[extra - 1]--
    } else {
      // extra 0 or 1: give P1 one day from P4
      dist[0]++
      dist[3]--
    }
  }
  return dist
}

function shiftDate(dateStr, n) {
  const [y, m, d] = dateStr.split('-').map(Number)
  const dt = new Date(y, m - 1, d + n)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

function daysBetween(start, end) {
  const [sy, sm, sd] = start.split('-').map(Number)
  const [ey, em, ed] = end.split('-').map(Number)
  const ms = new Date(ey, em - 1, ed) - new Date(sy, sm - 1, sd)
  return Math.round(ms / 86400000) + 1
}

function buildPeriods(cycleId, startDate, endDate, totalBudget, mode, seqRef) {
  const totalDays = daysBetween(startDate, endDate)
  const dist = calcDistribution(totalDays, mode)
  const budget = Math.round(totalBudget / 4)
  const now = new Date().toISOString()
  const periods = []
  let cur = startDate

  for (let i = 0; i < 4; i++) {
    seqRef.period++
    const pEnd = shiftDate(cur, dist[i] - 1)
    periods.push({
      id: seqRef.period,
      cycle_id: cycleId,
      period_number: i + 1,
      start_date: cur,
      end_date: pEnd,
      budget,
      status: 'open',
      result_type: '',
      result_amount: 0,
      created_at: now,
    })
    cur = shiftDate(pEnd, 1)
  }
  return periods
}

const ok = (v) => Promise.resolve(v)
const fail = (msg) => Promise.reject(new Error(msg))

export const api = {
  getCycles() {
    const { cycles } = load()
    return ok([...cycles].reverse())
  },

  getCycle(id) {
    const { cycles } = load()
    const cycle = cycles.find((c) => c.id === Number(id))
    return cycle ? ok(cycle) : fail('cycle not found')
  },

  createCycle({ start_date, end_date, total_budget, division_mode }) {
    const data = load()
    data.seq.cycle = (data.seq.cycle || 0) + 1
    data.seq.period = data.seq.period || 0

    const mode = division_mode || 'equal'
    const id = data.seq.cycle
    const now = new Date().toISOString()

    const cycle = {
      id,
      start_date,
      end_date,
      total_budget: Number(total_budget),
      division_mode: mode,
      created_at: now,
      periods: buildPeriods(id, start_date, end_date, total_budget, mode, data.seq),
    }

    data.cycles.push(cycle)
    persist(data)
    return ok(cycle)
  },

  deleteCycle(id) {
    const data = load()
    data.cycles = data.cycles.filter((c) => c.id !== Number(id))
    persist(data)
    return ok({ message: 'deleted' })
  },

  checkIn(periodId, { result_type, result_amount }) {
    const data = load()
    for (const cycle of data.cycles) {
      const p = cycle.periods.find((p) => p.id === Number(periodId))
      if (p) {
        if (p.status === 'completed') return fail('period already completed')
        p.status = 'completed'
        p.result_type = result_type
        p.result_amount = Number(result_amount)
        persist(data)
        return ok(p)
      }
    }
    return fail('period not found')
  },

  undoCheckIn(periodId) {
    const data = load()
    for (const cycle of data.cycles) {
      const p = cycle.periods.find((p) => p.id === Number(periodId))
      if (p) {
        p.status = 'open'
        p.result_type = ''
        p.result_amount = 0
        persist(data)
        return ok(p)
      }
    }
    return fail('period not found')
  },
}
