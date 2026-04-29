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
