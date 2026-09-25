export function parseDay(isoStr) {
  if (!isoStr) return null
  const [y, m, d] = isoStr.substring(0, 10).split('-').map(Number)
  return new Date(y, m - 1, d)
}

export function todayDate() {
  const d = new Date()
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

export function isActive(startStr, endStr) {
  const t = todayDate()
  return t >= parseDay(startStr) && t <= parseDay(endStr)
}

// Days remaining until end of period. 0 = today is last day. Negative = past.
export function daysLeft(endStr) {
  return Math.round((parseDay(endStr) - todayDate()) / 86400000)
}

// Days until period starts. Positive = future, 0 = starts today, negative = already started.
export function daysUntil(startStr) {
  return Math.round((parseDay(startStr) - todayDate()) / 86400000)
}

export function activePeriod(periods) {
  return (periods ?? []).find(p => isActive(p.start_date, p.end_date)) ?? null
}

export function fmtDate(isoStr) {
  if (!isoStr) return ''
  const s = isoStr.substring(0, 10)
  const [y, m, d] = s.split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString('id-ID', {
    day: 'numeric', month: 'short', year: 'numeric',
  })
}

export function fmtShort(isoStr) {
  if (!isoStr) return ''
  const s = isoStr.substring(0, 10)
  const [y, m, d] = s.split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString('id-ID', {
    day: 'numeric', month: 'short',
  })
}

export function fmtIDR(amount) {
  return new Intl.NumberFormat('id-ID').format(Math.round(amount))
}

export function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function addDays(dateStr, n) {
  const [y, m, d] = dateStr.split('-').map(Number)
  const dt = new Date(y, m - 1, d + n)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

// Inclusive day count between two YYYY-MM-DD strings.
export function daysBetween(start, end) {
  return Math.round((parseDay(end) - parseDay(start)) / 86400000) + 1
}

// ── Period math (mirrors backend/handlers/cycles.go) ─────────────────────────

// Splits totalDays into n periods; callers guarantee totalDays >= n.
export function calcDistribution(totalDays, n, mode) {
  const base = Math.floor(totalDays / n)
  const extra = totalDays % n
  const dist = Array(n).fill(base)
  for (let i = 0; i < extra; i++) dist[i]++
  if (mode === 'behavioral' && n > 1) {
    const donor = extra >= 2 ? extra - 1 : n - 1
    if (dist[donor] > 1) { dist[0]++; dist[donor]-- }
  }
  return dist
}

// Budget per period; the last period takes the rounding remainder so the sum is exact.
export function calcBudgets(total, n, mode) {
  const tw = n * (n + 1) / 2
  const weight =
    mode === 'progresif' ? (i) => (i + 1) / tw :
    mode === 'menurun'   ? (i) => (n - i) / tw :
                           ()  => 1 / n
  const budgets = Array(n).fill(0)
  let sum = 0
  for (let i = 0; i < n - 1; i++) {
    budgets[i] = Math.floor(total * weight(i))
    sum += budgets[i]
  }
  budgets[n - 1] = total - sum
  return budgets
}

export function buildPeriods(startDate, endDate, total, mode, n) {
  const dist = calcDistribution(daysBetween(startDate, endDate), n, mode)
  const budgets = calcBudgets(total, n, mode)
  const periods = []
  let cur = startDate
  for (let i = 0; i < n; i++) {
    const end = addDays(cur, dist[i] - 1)
    periods.push({ period_number: i + 1, start_date: cur, end_date: end, budget: budgets[i], days: dist[i] })
    cur = addDays(end, 1)
  }
  return periods
}

// ── Adjusted budget ──────────────────────────────────────────────────────────

// Spreads the running surplus/deficit of completed periods over the periods
// still open, in proportion to their planned budgets, so the cycle still lands
// on its total. Returns { carry, byId: Map(periodId → adjusted budget) }.
export function adjustedBudgets(periods) {
  const list = periods ?? []
  let carry = 0
  for (const p of list) {
    if (p.status !== 'completed') continue
    carry += p.result_type === 'sisa' ? p.result_amount : -p.result_amount
  }
  const open = list.filter(p => p.status !== 'completed')
  const openTotal = open.reduce((s, p) => s + p.budget, 0)
  const byId = new Map()
  if (!carry || !open.length || openTotal <= 0) return { carry, byId }

  let given = 0
  open.forEach((p, i) => {
    const share = i === open.length - 1 ? carry - given : Math.round(carry * p.budget / openTotal)
    given += share
    byId.set(p.id, Math.max(0, p.budget + share))
  })
  return { carry, byId }
}

// ── Cycle lifecycle ──────────────────────────────────────────────────────────

export function isOverdue(period) {
  return period.status === 'open' && daysLeft(period.end_date) < 0
}

// Overdue periods across all cycles, oldest first.
export function overduePeriods(cycles) {
  const out = []
  for (const c of cycles ?? []) {
    for (const p of c.periods ?? []) if (isOverdue(p)) out.push({ cycle: c, period: p })
  }
  return out.sort((a, b) => parseDay(a.period.end_date) - parseDay(b.period.end_date))
}

function addMonths(dateStr, n) {
  const [y, m, d] = dateStr.substring(0, 10).split('-').map(Number)
  const dt = new Date(y, m - 1 + n, 1)
  const lastDay = new Date(dt.getFullYear(), dt.getMonth() + 1, 0).getDate()
  dt.setDate(Math.min(d, lastDay))
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

// Settings for the cycle that follows `cycle`. Monthly cycles (e.g. 25 Jan – 24 Feb)
// stay aligned to the same day of the month; anything else keeps its length.
// After a gap, it skips ahead to the first cycle that hasn't ended yet.
export function nextCycleDraft(cycle) {
  const start = cycle.start_date.substring(0, 10)
  const end = cycle.end_date.substring(0, 10)
  const monthly = addDays(addMonths(start, 1), -1) === end
  const length = daysBetween(start, end)
  let nextStart = addDays(end, 1)
  let nextEnd
  for (let i = 0; i < 120; i++) {
    nextEnd = monthly ? addDays(addMonths(nextStart, 1), -1) : addDays(nextStart, length - 1)
    if (daysLeft(nextEnd) >= 0) break
    nextStart = addDays(nextEnd, 1)
  }
  return {
    start_date: nextStart,
    end_date: nextEnd,
    total_budget: cycle.total_budget,
    division_mode: cycle.division_mode,
    num_periods: cycle.num_periods || (cycle.periods ?? []).length || 4,
  }
}

// True when no cycle covers today or starts later — time to plan the next one.
export function needsNextCycle(cycles) {
  if (!cycles?.length) return false
  return !cycles.some(c => daysLeft(c.end_date) >= 0)
}

export function latestCycle(cycles) {
  return [...(cycles ?? [])].sort((a, b) => parseDay(b.end_date) - parseDay(a.end_date))[0] ?? null
}

export function cycleSummary(periods) {
  let totalSaved = 0, totalDeficit = 0, totalSpent = 0
  for (const p of periods) {
    if (p.status !== 'completed') continue
    if (p.result_type === 'sisa') {
      totalSaved += p.result_amount
      totalSpent += p.budget - p.result_amount
    } else if (p.result_type === 'defisit') {
      totalDeficit += p.result_amount
      totalSpent += p.budget + p.result_amount
    }
  }
  return { totalSaved, totalDeficit, net: totalSaved - totalDeficit, totalSpent }
}
