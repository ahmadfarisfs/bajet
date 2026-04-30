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
        <div class="empty-icon">
          <svg width="52" height="52" viewBox="0 0 512 512" fill="none">
            <rect width="512" height="512" rx="110" fill="#154374"/>
            <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
            <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
            <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
            <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
          </svg>
        </div>
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
    padding: 20px 16px;
    max-width: 480px;
    margin: 0 auto;
    padding-bottom: 48px;
  }
  .list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 18px;
  }
  h2 {
    font-family: var(--font-heading);
    font-size: 20px;
    font-weight: 800;
    color: var(--text);
  }

  .btn-new {
    background: var(--sapphire-dark);
    color: var(--banana);
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
    padding: 8px 16px;
    border-radius: var(--radius-sm);
    letter-spacing: 0.2px;
    transition: background 0.15s;
  }
  .btn-new:hover { background: var(--primary); color: white; }

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
  .cycle-card.active {
    border-color: var(--banana-dark);
    background: var(--banana-light);
  }

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
  .date-range {
    font-family: var(--font-heading);
    font-size: 15px;
    font-weight: 700;
    color: var(--text);
  }
  .aktif-badge {
    font-family: var(--font-heading);
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.8px;
    padding: 3px 8px;
    border-radius: 20px;
    background: var(--sapphire-dark);
    color: var(--banana);
  }
  .budget-text { font-size: 12px; color: var(--text-muted); }

  .right { text-align: right; flex-shrink: 0; margin-left: 8px; }
  .period-count {
    font-family: var(--font-heading);
    font-size: 22px;
    font-weight: 800;
    color: var(--primary);
    line-height: 1;
  }
  .cycle-card.active .period-count { color: var(--sapphire-dark); }
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
  .cycle-card.active .progress-bar { background: var(--sapphire-dark); }

  .current-period-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--pumpkin-light);
    border-radius: var(--radius-xs);
    padding: 7px 10px;
    border-left: 3px solid var(--pumpkin);
  }
  .cp-label { font-size: 12px; font-weight: 600; color: var(--pumpkin); }
  .cp-days  { font-size: 12px; font-weight: 700; color: var(--pumpkin); }
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
  .empty-icon { margin-bottom: 20px; display: flex; justify-content: center; }
  .empty p { margin-bottom: 20px; font-size: 15px; font-weight: 500; }

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
    background: var(--sapphire-dark);
    color: var(--banana);
    font-family: var(--font-heading);
    font-size: 15px;
    font-weight: 700;
    padding: 14px 28px;
    border-radius: var(--radius-sm);
    transition: background 0.15s;
  }
  .btn-start:hover { background: var(--primary); color: white; }
</style>
