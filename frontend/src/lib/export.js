import { addDays, parseDay } from './utils.js'

const day = (iso) => (iso ?? '').substring(0, 10)

export function downloadFile(filename, content, mime) {
  const url = URL.createObjectURL(new Blob([content], { type: mime }))
  const a = Object.assign(document.createElement('a'), { href: url, download: filename })
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}

// ── CSV backup ───────────────────────────────────────────────────────────────

function csvCell(v) {
  const s = String(v ?? '')
  return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s
}

// One row per period. Starts with a BOM so Excel reads it as UTF-8.
export function cyclesToCsv(cycles) {
  const header = [
    'cycle_start', 'cycle_end', 'cycle_budget', 'division_mode',
    'period', 'period_start', 'period_end', 'period_budget',
    'status', 'result', 'result_amount', 'net',
  ]
  const rows = [header]
  const ordered = [...(cycles ?? [])].sort((a, b) => parseDay(a.start_date) - parseDay(b.start_date))
  for (const c of ordered) {
    for (const p of c.periods ?? []) {
      const done = p.status === 'completed'
      const net = !done ? '' : p.result_type === 'sisa' ? p.result_amount : -p.result_amount
      rows.push([
        day(c.start_date), day(c.end_date), c.total_budget, c.division_mode,
        p.period_number, day(p.start_date), day(p.end_date), p.budget,
        p.status, done ? p.result_type : '', done ? p.result_amount : '', net,
      ])
    }
  }
  return '﻿' + rows.map(r => r.map(csvCell).join(',')).join('\r\n') + '\r\n'
}

// ── Calendar reminders (.ics) ────────────────────────────────────────────────

function icsText(s) {
  return String(s).replace(/\\/g, '\\\\').replace(/[,;]/g, m => '\\' + m).replace(/\n/g, '\\n')
}

// All-day event on the last day of every open period, with an alarm at 19:00
// that day, so the phone's own calendar reminds the user to check in.
export function cycleToIcs(cycle, { title, body }) {
  const compact = (iso) => day(iso).replace(/-/g, '')
  const stamp = new Date().toISOString().replace(/[-:]/g, '').replace(/\.\d+/, '')
  const lines = [
    'BEGIN:VCALENDAR', 'VERSION:2.0', 'PRODID:-//Bajet//Check-in reminders//EN', 'CALSCALE:GREGORIAN',
  ]
  for (const p of cycle.periods ?? []) {
    if (p.status === 'completed') continue
    lines.push(
      'BEGIN:VEVENT',
      `UID:bajet-period-${p.id}@bajet`,
      `DTSTAMP:${stamp}`,
      `DTSTART;VALUE=DATE:${compact(p.end_date)}`,
      `DTEND;VALUE=DATE:${compact(addDays(day(p.end_date), 1))}`,
      `SUMMARY:${icsText(title(p))}`,
      `DESCRIPTION:${icsText(body(p))}`,
      'BEGIN:VALARM', 'ACTION:DISPLAY', `DESCRIPTION:${icsText(title(p))}`, 'TRIGGER:PT19H', 'END:VALARM',
      'END:VEVENT',
    )
  }
  lines.push('END:VCALENDAR')
  return lines.join('\r\n') + '\r\n'
}
