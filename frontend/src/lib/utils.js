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

// ── Period math ──────────────────────────────────────────────────────────────

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
  // Integer weights and total * w / tw, so the split never drifts from rounding.
  const tw = mode === 'progresif' || mode === 'menurun' ? n * (n + 1) / 2 : n
  const weight =
    mode === 'progresif' ? (i) => i + 1 :
    mode === 'menurun'   ? (i) => n - i :
                           ()  => 1
  const budgets = Array(n).fill(0)
  let sum = 0
  for (let i = 0; i < n - 1; i++) {
    budgets[i] = Math.floor(total * weight(i) / tw)
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

export const CARRY_MODES = ['spread', 'save', 'none']

// Works out each period's actual budget (its "target"): the planned budget plus
// its share of what earlier periods left over or overspent, following the
// cycle's carry mode:
//   spread – surplus and deficit are both passed on to later periods
//   save   – surplus is set aside as savings; only deficit is passed on
//   none   – every period keeps its planned budget
// A check-in stores the period's target at that moment (`target`) and its result
// is measured against it. Older check-ins without one get the target the old
// version displayed when periods were checked in in order. Shares
// are in proportion to the later periods' planned budgets.
// Returns {
//   targets: Map(periodId → actual budget) for every period,
//   byId:    Map(periodId → adjusted budget) for open periods that changed,
//   carry:   total passed on to the open periods,
//   saved:   surplus set aside (save mode),
// }
export function adjustedBudgets(periods, carryMode = 'spread') {
  const list = [...(periods ?? [])].sort((x, y) => x.period_number - y.period_number)
  const targets = new Map(), byId = new Map()
  let pending = 0, saved = 0, carry = 0
  // Older versions showed an open period its planned budget plus a share of
  // the earlier results only; check-ins made then were measured against that.
  let legacy = 0
  let rest = list.reduce((s, p) => s + p.budget, 0)

  list.forEach((p, i) => {
    const done = p.status === 'completed'
    const spreadOf = (amt) => i === list.length - 1 || rest <= 0 ? amt : Math.round(amt * p.budget / rest)
    let share = 0
    if (done && p.target > 0) share = p.target - p.budget      // budget it had at check-in
    else if (done)            share = legacy ? spreadOf(legacy) : 0
    else if (pending)         share = spreadOf(pending)
    pending -= share
    rest -= p.budget
    const target = Math.max(0, p.budget + share)
    targets.set(p.id, target)

    if (!done) {
      carry += share
      if (share) byId.set(p.id, target)
      return
    }
    if (carryMode === 'none') return
    let counted = -p.result_amount
    if (p.result_type === 'sisa') {
      if (carryMode === 'save') { saved += p.result_amount; counted = 0 }
      else counted = p.result_amount
    }
    pending += counted
    legacy += counted
  })
  return { carry, saved, byId, targets }
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
    carry_mode: cycle.carry_mode || 'spread',
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

export function cycleSummary(periods, carryMode) {
  const { targets } = adjustedBudgets(periods, carryMode)
  let totalSaved = 0, totalDeficit = 0, totalSpent = 0
  for (const p of periods ?? []) {
    if (p.status !== 'completed') continue
    const target = targets.get(p.id) ?? p.budget
    if (p.result_type === 'sisa') {
      totalSaved += p.result_amount
      totalSpent += target - p.result_amount
    } else if (p.result_type === 'defisit') {
      totalDeficit += p.result_amount
      totalSpent += target + p.result_amount
    }
  }
  return { totalSaved, totalDeficit, net: totalSaved - totalDeficit, totalSpent }
}
