<script>
  import { api } from '../lib/api.js'
  import { fmtDate, fmtIDR, cycleSummary, isActive, activePeriod, daysLeft } from '../lib/utils.js'
  import PeriodCard from './PeriodCard.svelte'

  let { cycleId, onBack } = $props()

  let cycle = $state(null)
  let loading = $state(true)
  let error = $state('')

  async function load() {
    loading = true; error = ''
    try {
      cycle = await api.getCycle(cycleId)
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }

  async function deleteCycle() {
    if (!confirm('Hapus cycle ini beserta semua datanya?')) return
    try {
      await api.deleteCycle(cycleId)
      onBack()
    } catch (e) {
      alert(e.message)
    }
  }

  $effect(() => { load() })

  let summary   = $derived(cycle ? cycleSummary(cycle.periods ?? []) : null)
  let completed = $derived(cycle ? (cycle.periods ?? []).filter(p => p.status === 'completed').length : 0)
  let total     = $derived(cycle ? (cycle.periods ?? []).length : 0)
  let isCurrent = $derived(cycle ? isActive(cycle.start_date, cycle.end_date) : false)
  let cp        = $derived(cycle ? activePeriod(cycle.periods ?? []) : null)
</script>

<div class="view">
  <div class="topbar">
    <button class="back" onclick={onBack}>← Kembali</button>
    <button class="del-btn" onclick={deleteCycle} title="Hapus cycle">🗑</button>
  </div>

  {#if loading}
    <div class="center">Memuat...</div>
  {:else if error}
    <div class="center err">{error}</div>
  {:else if cycle}
    <div class="cycle-header">
      <div class="cycle-title">
        <h1>{fmtDate(cycle.start_date)} – {fmtDate(cycle.end_date)}</h1>
        <span class="mode-badge">{cycle.division_mode === 'behavioral' ? 'Behavioral' : 'Equal'}</span>
      </div>
      <p class="budget-label">Budget: <strong>Rp {fmtIDR(cycle.total_budget)}</strong></p>
      <div class="progress-bar-wrap">
        <div class="progress-bar" style="width: {total ? (completed/total*100) : 0}%"></div>
      </div>
      <p class="progress-text">{completed}/{total} periode selesai</p>

      {#if isCurrent && cp && cp.status === 'open'}
        {@const left = daysLeft(cp.end_date)}
        <div class="active-period-banner">
          <div class="ap-left">
            <span class="ap-label">Periode aktif</span>
            <span class="ap-name">P{cp.period_number}</span>
          </div>
          <span class="ap-countdown" class:urgent={left <= 1}>
            {left <= 0 ? 'Hari ini terakhir!' : `${left} hari lagi`}
          </span>
        </div>
      {/if}
    </div>

    {#if summary && completed > 0}
      <div class="summary-grid">
        <div class="summary-card green">
          <div class="s-label">Total Sisa</div>
          <div class="s-val">Rp {fmtIDR(summary.totalSaved)}</div>
        </div>
        <div class="summary-card red">
          <div class="s-label">Total Defisit</div>
          <div class="s-val">Rp {fmtIDR(summary.totalDeficit)}</div>
        </div>
        <div class="summary-card" class:green={summary.net >= 0} class:red={summary.net < 0}>
          <div class="s-label">Net</div>
          <div class="s-val">{summary.net >= 0 ? '+' : ''}Rp {fmtIDR(Math.abs(summary.net))}</div>
        </div>
        <div class="summary-card">
          <div class="s-label">Terpakai</div>
          <div class="s-val">Rp {fmtIDR(summary.totalSpent)}</div>
        </div>
      </div>
    {/if}

    <div class="periods">
      {#each (cycle.periods ?? []) as period (period.id)}
        <PeriodCard {period} onUpdate={load} />
      {/each}
    </div>
  {/if}
</div>

<style>
  .view {
    max-width: 480px;
    margin: 0 auto;
    padding: 16px;
    padding-bottom: 40px;
  }
  .topbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }
  .back {
    background: none;
    font-size: 14px;
    font-weight: 600;
    color: var(--primary);
    padding: 6px 0;
  }
  .del-btn {
    background: none;
    font-size: 18px;
    padding: 6px;
    border-radius: 6px;
    color: var(--text-muted);
  }
  .del-btn:hover { background: var(--danger-light); color: var(--danger); }

  .cycle-header {
    background: var(--surface);
    border-radius: var(--radius);
    padding: 20px;
    margin-bottom: 16px;
    box-shadow: var(--shadow-sm);
  }
  .cycle-title {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 6px;
  }
  h1 {
    font-size: 16px;
    font-weight: 700;
    color: var(--text);
  }
  .mode-badge {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 20px;
    background: var(--primary-light);
    color: var(--primary);
  }
  .budget-label {
    font-size: 13px;
    color: var(--text-muted);
    margin-bottom: 12px;
  }
  .progress-bar-wrap {
    height: 6px;
    background: var(--border);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 6px;
  }
  .progress-bar {
    height: 100%;
    background: var(--primary);
    border-radius: 3px;
    transition: width 0.4s ease;
  }
  .progress-text {
    font-size: 12px;
    color: var(--text-muted);
  }

  .active-period-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 12px;
    padding: 10px 14px;
    background: var(--warning-light);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--warning);
  }
  .ap-left { display: flex; flex-direction: column; gap: 1px; }
  .ap-label { font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: var(--warning); }
  .ap-name  { font-size: 16px; font-weight: 800; color: var(--warning); }
  .ap-countdown { font-size: 14px; font-weight: 700; color: var(--warning); }
  .ap-countdown.urgent { color: var(--danger); }

  .summary-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    margin-bottom: 16px;
  }
  .summary-card {
    background: var(--surface);
    border-radius: var(--radius-sm);
    padding: 14px;
    box-shadow: var(--shadow-sm);
  }
  .summary-card.green { background: var(--success-light); }
  .summary-card.red { background: var(--danger-light); }

  .s-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 4px;
  }
  .summary-card.green .s-label { color: var(--success); }
  .summary-card.red .s-label { color: var(--danger); }

  .s-val {
    font-size: 15px;
    font-weight: 700;
    color: var(--text);
  }
  .summary-card.green .s-val { color: var(--success); }
  .summary-card.red .s-val { color: var(--danger); }

  .periods {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .center {
    text-align: center;
    padding: 40px;
    color: var(--text-muted);
  }
  .err { color: var(--danger); }
</style>
