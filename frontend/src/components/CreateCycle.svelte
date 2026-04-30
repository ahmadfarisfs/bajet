<script>
  import { api } from '../lib/api.js'
  import { todayStr, addDays } from '../lib/utils.js'

  let { onCreated, onCancel } = $props()

  let startDate = $state(todayStr())
  let endDate = $state(addDays(todayStr(), 29))
  let totalBudget = $state(2000000)
  let divisionMode = $state('equal')
  let numPeriods = $state(4)
  let loading = $state(false)
  let error = $state('')

  // ── Preview helpers (mirrors backend logic) ───────────────────────────────
  function shiftDate(dateStr, n) {
    const [y, m, d] = dateStr.split('-').map(Number)
    const dt = new Date(y, m - 1, d + n)
    return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
  }

  function daysBetween(start, end) {
    const [sy, sm, sd] = start.split('-').map(Number)
    const [ey, em, ed] = end.split('-').map(Number)
    return Math.round((new Date(ey, em - 1, ed) - new Date(sy, sm - 1, sd)) / 86400000) + 1
  }

  function calcDist(totalDays, n, mode) {
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

  function calcBudgets(total, n, mode) {
    const budgets = Array(n).fill(0)
    if (mode === 'progresif') {
      const tw = n * (n + 1) / 2
      let sum = 0
      for (let i = 0; i < n - 1; i++) {
        const b = Math.floor(total * (i + 1) / tw)
        budgets[i] = b; sum += b
      }
      budgets[n - 1] = total - sum
    } else if (mode === 'menurun') {
      const tw = n * (n + 1) / 2
      let sum = 0
      for (let i = 0; i < n - 1; i++) {
        const b = Math.floor(total * (n - i) / tw)
        budgets[i] = b; sum += b
      }
      budgets[n - 1] = total - sum
    } else {
      const base = Math.round(total / n)
      budgets.fill(base)
    }
    return budgets
  }

  let previewPeriods = $derived.by(() => {
    const n = Math.min(Math.max(Math.round(Number(numPeriods)) || 1, 1), 12)
    if (!startDate || !endDate) return []
    const totalDays = daysBetween(startDate, endDate)
    if (totalDays < n) return []
    const dist = calcDist(totalDays, n, divisionMode)
    const budgets = calcBudgets(Number(totalBudget) || 0, n, divisionMode)
    const periods = []
    let cur = startDate
    for (let i = 0; i < n; i++) {
      const pEnd = shiftDate(cur, dist[i] - 1)
      periods.push({ num: i + 1, start: cur, end: pEnd, budget: budgets[i], days: dist[i] })
      cur = shiftDate(pEnd, 1)
    }
    return periods
  })

  function fmtDate(d) {
    const [, m, day] = d.split('-')
    return `${day}/${m}`
  }

  function fmtIDR(n) {
    return new Intl.NumberFormat('id-ID').format(Math.round(n))
  }

  async function submit() {
    if (!startDate || !endDate) { error = 'Tanggal harus diisi'; return }
    if (totalBudget <= 0) { error = 'Budget harus lebih dari 0'; return }
    const n = Math.min(Math.max(Math.round(Number(numPeriods)) || 4, 1), 12)
    loading = true; error = ''
    try {
      const cycle = await api.createCycle({
        start_date: startDate,
        end_date: endDate,
        total_budget: Number(totalBudget),
        division_mode: divisionMode,
        num_periods: n,
      })
      onCreated(cycle)
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }
</script>

<div class="view">
  <div class="topbar">
    <button class="back" onclick={onCancel}>← Batal</button>
  </div>

  <div class="card">
    <h2>Buat Cycle Baru</h2>

    <label>
      Tanggal Mulai
      <input type="date" bind:value={startDate} />
    </label>

    <label>
      Tanggal Selesai
      <input type="date" bind:value={endDate} />
    </label>

    <label>
      Total Budget (IDR)
      <div class="input-row">
        <span class="prefix">Rp</span>
        <input type="number" min="1" bind:value={totalBudget} />
      </div>
    </label>

    <label>
      Jumlah Periode
      <input type="number" min="1" max="12" step="1" bind:value={numPeriods} />
    </label>

    <div class="field">
      <span class="field-label">Mode Pembagian</span>
      <div class="toggle">
        <button
          class="toggle-btn"
          class:active={divisionMode === 'equal'}
          onclick={() => divisionMode = 'equal'}
        >
          <strong>Equal</strong>
          <small>Merata</small>
        </button>
        <button
          class="toggle-btn"
          class:active={divisionMode === 'behavioral'}
          onclick={() => divisionMode = 'behavioral'}
        >
          <strong>Behavioral</strong>
          <small>Awal panjang</small>
        </button>
        <button
          class="toggle-btn"
          class:active={divisionMode === 'menurun'}
          onclick={() => divisionMode = 'menurun'}
        >
          <strong>Menurun</strong>
          <small>Budget turun</small>
        </button>
        <button
          class="toggle-btn"
          class:active={divisionMode === 'progresif'}
          onclick={() => divisionMode = 'progresif'}
        >
          <strong>Progresif</strong>
          <small>Budget naik</small>
        </button>
      </div>
    </div>

    {#if previewPeriods.length > 0}
      <div class="preview">
        <span class="preview-label">Preview Periode</span>
        <table>
          <thead>
            <tr><th>#</th><th>Tanggal</th><th>Hari</th><th>Budget</th></tr>
          </thead>
          <tbody>
            {#each previewPeriods as p}
              <tr>
                <td class="p-num">{p.num}</td>
                <td class="p-date">{fmtDate(p.start)} – {fmtDate(p.end)}</td>
                <td class="p-days">{p.days}h</td>
                <td class="p-budget">Rp {fmtIDR(p.budget)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}

    {#if error}<p class="err">{error}</p>{/if}

    <button class="btn-primary" onclick={submit} disabled={loading}>
      {loading ? 'Membuat...' : 'Buat Cycle'}
    </button>
  </div>

  <p class="hint">
    Cycle akan dibagi menjadi <strong>{numPeriods} periode</strong> berdasarkan rentang tanggal.
  </p>
</div>

<style>
  .view {
    max-width: 480px;
    margin: 0 auto;
    padding: 16px;
    padding-bottom: 48px;
  }
  .topbar { margin-bottom: 20px; }
  .back {
    background: none;
    font-family: var(--font-heading);
    font-size: 14px;
    font-weight: 700;
    color: var(--primary);
    padding: 6px 0;
  }

  .card {
    background: var(--surface);
    border-radius: var(--radius);
    padding: 24px 20px;
    box-shadow: var(--shadow-sm);
    display: flex;
    flex-direction: column;
    gap: 18px;
    border-top: 4px solid var(--sapphire-dark);
  }
  h2 {
    font-family: var(--font-heading);
    font-size: 20px;
    font-weight: 800;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-muted);
  }
  label input[type="date"],
  label input[type="number"] {
    border: 1.5px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0 12px;
    height: 44px;
    font-size: 15px;
    color: var(--text);
    background: var(--surface);
    outline: none;
    width: 100%;
    transition: border-color 0.15s;
  }
  label input:focus { border-color: var(--primary); }

  .input-row {
    display: flex;
    align-items: center;
    border: 1.5px solid var(--border);
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: var(--surface);
    transition: border-color 0.15s;
  }
  .input-row:focus-within { border-color: var(--primary); }
  .prefix {
    padding: 0 12px;
    color: var(--text-muted);
    font-size: 13px;
    font-weight: 600;
    border-right: 1.5px solid var(--border);
    background: var(--surface-2);
    height: 44px;
    display: flex;
    align-items: center;
  }
  .input-row input {
    flex: 1;
    border: none !important;
    padding: 0 12px;
    height: 44px;
    font-family: var(--font-heading);
    font-size: 15px;
    font-weight: 700;
    color: var(--text);
    background: transparent;
    outline: none;
  }

  .field { display: flex; flex-direction: column; gap: 6px; }
  .field-label { font-size: 13px; font-weight: 600; color: var(--text-muted); }

  .toggle { display: flex; gap: 6px; flex-wrap: wrap; }
  .toggle-btn {
    flex: 1;
    min-width: calc(50% - 4px);
    padding: 10px 8px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    background: var(--surface-2);
    color: var(--text-muted);
    border: 2px solid transparent;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    transition: all 0.15s;
  }
  .toggle-btn strong { font-size: 13px; color: var(--text); font-family: var(--font-heading); }
  .toggle-btn small  { font-size: 11px; }
  .toggle-btn.active {
    background: var(--banana-light);
    border-color: var(--banana-dark);
  }
  .toggle-btn.active strong { color: var(--sapphire-dark); }
  .toggle-btn.active small  { color: var(--sapphire-dark); }

  .err { font-size: 13px; color: var(--danger); margin-top: -8px; }

  .btn-primary {
    width: 100%;
    padding: 14px;
    border-radius: var(--radius-sm);
    font-family: var(--font-heading);
    font-size: 15px;
    font-weight: 800;
    background: var(--sapphire-dark);
    color: var(--banana);
    letter-spacing: 0.2px;
    transition: background 0.15s;
  }
  .btn-primary:hover:not(:disabled) { background: var(--primary); color: white; }
  .btn-primary:disabled { opacity: 0.6; }

  .hint {
    text-align: center;
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 16px;
    line-height: 1.5;
  }

  .preview {
    display: flex;
    flex-direction: column;
    gap: 8px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    padding: 12px;
  }
  .preview-label { font-size: 12px; font-weight: 600; color: var(--text-muted); }
  .preview table { width: 100%; border-collapse: collapse; font-size: 12px; }
  .preview th {
    text-align: left;
    font-size: 11px;
    color: var(--text-muted);
    font-weight: 600;
    padding: 4px 6px;
    border-bottom: 1px solid var(--border);
  }
  .preview td { padding: 5px 6px; color: var(--text); }
  .preview tr:not(:last-child) td { border-bottom: 1px solid var(--border); }
  .p-num { font-weight: 700; color: var(--primary); width: 20px; }
  .p-days { color: var(--text-muted); width: 32px; }
  .p-budget { font-weight: 600; text-align: right; }
</style>
