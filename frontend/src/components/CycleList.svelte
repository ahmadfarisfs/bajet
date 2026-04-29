<script>
  import { fmtShort, fmtIDR, cycleSummary } from '../lib/utils.js'

  let { cycles, onSelect, onNew } = $props()

  function summary(cycle) {
    return cycleSummary(cycle.periods ?? [])
  }

  function completedCount(cycle) {
    return (cycle.periods ?? []).filter(p => p.status === 'completed').length
  }
</script>

<div class="list">
  <div class="list-header">
    <h2>Cycle Saya</h2>
    <button class="btn-new" onclick={onNew}>+ Baru</button>
  </div>

  {#if cycles.length === 0}
    <div class="empty">
      <p>Belum ada cycle.</p>
      <button class="btn-start" onclick={onNew}>Buat Cycle Pertama</button>
    </div>
  {:else}
    {#each cycles as cycle (cycle.id)}
      {@const s = summary(cycle)}
      {@const done = completedCount(cycle)}
      {@const total = (cycle.periods ?? []).length}
      <button class="cycle-card" onclick={() => onSelect(cycle.id)}>
        <div class="top">
          <div>
            <div class="date-range">{fmtShort(cycle.start_date)} – {fmtShort(cycle.end_date)}</div>
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

        {#if done > 0}
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
    border: none;
  }
  .cycle-card:hover { box-shadow: var(--shadow); transform: translateY(-1px); }

  .top {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }
  .date-range { font-size: 15px; font-weight: 700; color: var(--text); margin-bottom: 3px; }
  .budget-text { font-size: 12px; color: var(--text-muted); }
  .right { text-align: right; }
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

  .stats { display: flex; gap: 8px; }
  .stat {
    font-size: 12px;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 20px;
  }
  .stat.green { background: var(--success-light); color: var(--success); }
  .stat.red { background: var(--danger-light); color: var(--danger); }

  .empty {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-muted);
  }
  .empty p { margin-bottom: 16px; font-size: 15px; }
  .btn-start {
    background: var(--primary);
    color: white;
    font-size: 14px;
    font-weight: 700;
    padding: 12px 24px;
    border-radius: var(--radius-sm);
  }
</style>
