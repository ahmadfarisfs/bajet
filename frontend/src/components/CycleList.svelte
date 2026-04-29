<script>
  import { fmtShort, fmtIDR, cycleSummary, isActive, activePeriod, daysLeft } from '../lib/utils.js'

  let { cycles, loading = false, onSelect, onNew } = $props()

  function summary(cycle) { return cycleSummary(cycle.periods ?? []) }
  function completedCount(cycle) { return (cycle.periods ?? []).filter(p => p.status === 'completed').length }
  function currentCycle(cycle) { return isActive(cycle.start_date, cycle.end_date) }
  function currentPeriod(cycle) { return activePeriod(cycle.periods ?? []) }
</script>

<div class="list">
  <div class="list-header">
    <h2>Cycle Saya</h2>
    <button class="btn-new" onclick={onNew}>+ Baru</button>
  </div>

  {#if cycles.length === 0}
    {#if loading}
      <div class="empty"><span class="spinner"></span></div>
    {:else}
      <div class="empty">
        <p>Belum ada cycle.</p>
        <button class="btn-start" onclick={onNew}>Buat Cycle Pertama</button>
      </div>
    {/if}
  {:else}
    {#each cycles as cycle (cycle.id)}
      {@const s = summary(cycle)}
      {@const done = completedCount(cycle)}
      {@const total = (cycle.periods ?? []).length}
      {@const active = currentCycle(cycle)}
      {@const cp = currentPeriod(cycle)}
      <button class="cycle-card" class:active onclick={() => onSelect(cycle.id)}>
        <div class="top">
          <div class="top-left">
            <div class="date-row">
              <span class="date-range">{fmtShort(cycle.start_date)} – {fmtShort(cycle.end_date)}</span>
              {#if active}
                <span class="aktif-badge">AKTIF</span>
              {/if}
            </div>
            <div class="budget-text">Budget: Rp {fmtIDR(cycle.total_budget)}</div>
          </div>
          <div class="right">
            <div class="period-count">{done}/{total}</div>
            <div class="period-label">periode</div>
          </div>
        </div>

        <div class="progress-bar-wrap">
          <div class="progress-bar" style="width: {total ? (done/total*100) : 0}%"></div>
        </div>

        {#if active && cp && cp.status === 'open'}
          {@const left = daysLeft(cp.end_date)}
          <div class="current-period-row">
            <span class="cp-label">P{cp.period_number} sedang berjalan</span>
            <span class="cp-days" class:urgent={left <= 1}>
              {left <= 0 ? 'Hari ini terakhir!' : `${left} hari lagi`}
            </span>
          </div>
        {:else if done > 0}
          <div class="stats">
            {#if s.totalSaved > 0}
              <span class="stat green">↑ Rp {fmtIDR(s.totalSaved)}</span>
            {/if}
            {#if s.totalDeficit > 0}
              <span class="stat red">↓ Rp {fmtIDR(s.totalDeficit)}</span>
            {/if}
          </div>
        {/if}
      </button>
    {/each}
  {/if}
</div>

<style>
  .list {
    padding: 16px;
    max-width: 480px;
    margin: 0 auto;
    padding-bottom: 40px;
  }
  .list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  h2 { font-size: 18px; font-weight: 700; }

  .btn-new {
    background: var(--primary);
    color: white;
    font-size: 13px;
    font-weight: 700;
    padding: 8px 16px;
    border-radius: var(--radius-sm);
  }
  .btn-new:hover { background: var(--primary-dark); }

  .cycle-card {
    width: 100%;
    background: var(--surface);
    border-radius: var(--radius);
    padding: 16px;
    margin-bottom: 10px;
    text-align: left;
    box-shadow: var(--shadow-sm);
    transition: box-shadow 0.15s, transform 0.1s;
    cursor: pointer;
    border: 2px solid transparent;
  }
  .cycle-card:hover { box-shadow: var(--shadow); transform: translateY(-1px); }
  .cycle-card.active { border-color: var(--warning); background: #fffdf5; }

  .top {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }
  .top-left { flex: 1; min-width: 0; }
  .date-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 3px;
    flex-wrap: wrap;
  }
  .date-range { font-size: 15px; font-weight: 700; color: var(--text); }
  .aktif-badge {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.5px;
    padding: 2px 7px;
    border-radius: 20px;
    background: var(--warning);
    color: white;
  }
  .budget-text { font-size: 12px; color: var(--text-muted); }

  .right { text-align: right; flex-shrink: 0; margin-left: 8px; }
  .period-count { font-size: 20px; font-weight: 700; color: var(--primary); line-height: 1; }
  .period-label { font-size: 11px; color: var(--text-muted); }

  .progress-bar-wrap {
    height: 5px;
    background: var(--border);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 10px;
  }
  .progress-bar {
    height: 100%;
    background: var(--primary);
    border-radius: 3px;
    transition: width 0.4s;
  }
  .cycle-card.active .progress-bar { background: var(--warning); }

  .current-period-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--warning-light);
    border-radius: var(--radius-sm);
    padding: 7px 10px;
  }
  .cp-label { font-size: 12px; font-weight: 600; color: var(--warning); }
  .cp-days  { font-size: 12px; font-weight: 700; color: var(--warning); }
  .cp-days.urgent { color: var(--danger); }

  .stats { display: flex; gap: 8px; }
  .stat {
    font-size: 12px;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 20px;
  }
  .stat.green { background: var(--success-light); color: var(--success); }
  .stat.red   { background: var(--danger-light);  color: var(--danger);  }

  .empty {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-muted);
  }
  .empty p { margin-bottom: 16px; font-size: 15px; }

  .spinner {
    display: inline-block;
    width: 28px;
    height: 28px;
    border: 3px solid var(--border);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  .btn-start {
    background: var(--primary);
    color: white;
    font-size: 14px;
    font-weight: 700;
    padding: 12px 24px;
    border-radius: var(--radius-sm);
  }
</style>
