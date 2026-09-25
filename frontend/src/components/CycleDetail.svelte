<script>
  import { untrack } from 'svelte'
  import { api } from '../lib/api.js'
  import { fmtDate, fmtIDR, cycleSummary, isActive, activePeriod, daysLeft, adjustedBudgets } from '../lib/utils.js'
  import { cycleToIcs, downloadFile } from '../lib/export.js'
  import PeriodCard from './PeriodCard.svelte'
  import PeriodStrip from './PeriodStrip.svelte'
  import { i18n } from '../lib/i18n.js'

  let { cycleId, onBack, onEdit, onStartNext, isLatest = false, initialCycle = null } = $props()

  // Seed with the pre-loaded cycle from parent for instant display; refresh in background
  let cycle = $state(initialCycle ?? null)
  let loading = $state(initialCycle == null)
  let error = $state('')

  async function load() {
    // Don't blank the screen if we already have data — refresh silently
    if (!cycle) loading = true
    error = ''
    try {
      cycle = await api.getCycle(cycleId)
    } catch (e) {
      if (!cycle) error = e.message
    } finally {
      loading = false
    }
  }

  async function deleteCycle() {
    if (!confirm($i18n.deleteConfirm)) return
    try {
      await api.deleteCycle(cycleId)
      onBack()
    } catch (e) {
      alert(e.message)
    }
  }

  function exportReminders() {
    const ics = cycleToIcs(cycle, {
      title: (p) => $i18n.icsTitle(p.period_number),
      body:  (p) => $i18n.icsBody(p.period_number, fmtIDR(adjusted.byId.get(p.id) ?? p.budget)),
    })
    downloadFile(`bajet-${cycle.start_date.substring(0, 10)}.ics`, ics, 'text/calendar')
  }

  // Reload only when the id changes; load() reads `cycle`, and tracking it
  // would make every successful load trigger another one.
  $effect(() => { cycleId; untrack(load) })

  const MODE_KEY = { equal: 'modeEqual', behavioral: 'modeBehav', menurun: 'modeMenurun', progresif: 'modeProgresif' }

  let periods   = $derived(cycle?.periods ?? [])
  let summary   = $derived(cycleSummary(periods))
  let adjusted  = $derived(adjustedBudgets(periods))
  let completed = $derived(periods.filter(p => p.status === 'completed').length)
  let total     = $derived(periods.length)
  let isCurrent = $derived(cycle ? isActive(cycle.start_date, cycle.end_date) : false)
  let cp        = $derived(activePeriod(periods))
  let hasOpen   = $derived(periods.some(p => p.status !== 'completed'))
  let finished  = $derived(cycle ? (daysLeft(cycle.end_date) < 0 || (total > 0 && completed === total)) : false)
</script>

<div class="page">
  <div class="topbar">
    <button class="back" onclick={onBack}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
      {$i18n.back}
    </button>
    {#if cycle}
      <div class="tools">
        <button class="icon-btn" onclick={() => onEdit(cycle)} title={$i18n.editBtn} aria-label={$i18n.editBtn}>
          <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>
        </button>
        {#if hasOpen}
          <button class="icon-btn" onclick={exportReminders} title={$i18n.remindCalHint} aria-label={$i18n.remindCal}>
            <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/><path d="M12 14v4M10 16h4"/></svg>
          </button>
        {/if}
        <button class="icon-btn danger" onclick={deleteCycle} title={$i18n.deleteConfirm} aria-label={$i18n.deleteConfirm}>
          <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
        </button>
      </div>
    {/if}
  </div>

  {#if loading}
    <div class="center"><span class="spinner"></span></div>
  {:else if error}
    <div class="center err">{error}</div>
  {:else if cycle}
    <div class="head">
      <div class="head-row">
        <span class="mode">{$i18n[MODE_KEY[cycle.division_mode] ?? 'modeEqual']}</span>
        {#if isCurrent && cp && cp.status === 'open'}
          {@const left = daysLeft(cp.end_date)}
          <span class="now" class:urgent={left <= 1}>P{cp.period_number} · {left <= 0 ? $i18n.lastDay : $i18n.daysLeftN(left)}</span>
        {/if}
      </div>
      <h1>{fmtDate(cycle.start_date)} – {fmtDate(cycle.end_date)}</h1>
      <div class="head-budget num">Rp {fmtIDR(cycle.total_budget)}</div>
      <PeriodStrip {periods} tone="dark" />
      <p class="head-progress">{$i18n.periodsDone(completed, total)}</p>
    </div>

    {#if finished && isLatest}
      <div class="next-card">
        <div>
          <strong>{$i18n.cycleFinished}</strong>
          <small>{$i18n.nextCycleSub}</small>
        </div>
        <button class="btn-next" onclick={() => onStartNext(cycle)}>{$i18n.startNextCycle}</button>
      </div>
    {/if}

    {#if completed > 0}
      <div class="stats">
        <div class="stat">
          <span class="s-label">{$i18n.net}</span>
          <span class="s-val num" class:green={summary.net >= 0} class:red={summary.net < 0}>
            {summary.net >= 0 ? '+' : '−'}Rp {fmtIDR(Math.abs(summary.net))}
          </span>
        </div>
        <div class="stat">
          <span class="s-label">{$i18n.spent}</span>
          <span class="s-val num">Rp {fmtIDR(summary.totalSpent)}</span>
        </div>
        <div class="stat">
          <span class="s-label">{$i18n.totalSurplus}</span>
          <span class="s-val num green">Rp {fmtIDR(summary.totalSaved)}</span>
        </div>
        <div class="stat">
          <span class="s-label">{$i18n.totalDeficit}</span>
          <span class="s-val num red">Rp {fmtIDR(summary.totalDeficit)}</span>
        </div>
      </div>
    {/if}

    {#if adjusted.byId.size > 0}
      <div class="carry" class:red={adjusted.carry < 0}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 12h13M13 6l6 6-6 6"/></svg>
        {adjusted.carry > 0
          ? $i18n.carrySurplus(fmtIDR(adjusted.carry))
          : $i18n.carryDeficit(fmtIDR(-adjusted.carry))}
      </div>
    {/if}

    <div class="periods">
      {#each periods as period (period.id)}
        <PeriodCard {period} adjusted={adjusted.byId.get(period.id) ?? null} onUpdate={load} />
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    max-width: var(--page-max);
    margin: 0 auto;
    padding: 12px 16px calc(48px + var(--safe-bottom));
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .topbar { display: flex; justify-content: space-between; align-items: center; }
  .back {
    display: inline-flex; align-items: center; gap: 2px;
    background: none;
    font-family: var(--font-heading); font-size: 15px; font-weight: 700;
    color: var(--sapphire-dark);
    padding: 8px 8px 8px 0;
  }
  .tools { display: flex; gap: 6px; }
  .icon-btn {
    width: 38px; height: 38px;
    display: grid; place-items: center;
    border-radius: 12px;
    background: var(--surface); border: 1px solid var(--border);
    color: var(--text-muted);
    transition: background 0.15s, color 0.15s;
  }
  .icon-btn:hover { background: var(--sapphire-light); color: var(--sapphire-dark); }
  .icon-btn.danger:hover { background: var(--danger-light); color: var(--danger); }

  .head {
    color: #fff;
    background:
      radial-gradient(120% 90% at 100% 0%, rgba(242,233,66,0.14), transparent 55%),
      linear-gradient(160deg, var(--sapphire-dark), var(--sapphire-deep));
    border-radius: var(--radius);
    padding: 18px;
    box-shadow: var(--shadow-lg);
  }
  .head-row { display: flex; justify-content: space-between; align-items: center; gap: 8px; margin-bottom: 10px; }
  .mode {
    font-size: 11px; font-weight: 700; letter-spacing: 0.4px;
    padding: 3px 10px; border-radius: 999px;
    background: rgba(255,255,255,0.14); color: rgba(255,255,255,0.9);
  }
  .now { font-size: 12px; font-weight: 700; color: var(--banana); }
  .now.urgent { color: #ffb4a8; }
  h1 { font-family: var(--font-heading); font-size: 17px; font-weight: 700; color: rgba(255,255,255,0.85); }
  .head-budget {
    font-family: var(--font-heading);
    font-size: 32px; font-weight: 800; letter-spacing: -0.8px;
    margin: 2px 0 14px;
  }
  .head-progress { font-size: 12px; color: rgba(255,255,255,0.65); margin-top: 8px; }

  .next-card {
    display: flex; align-items: center; gap: 12px;
    padding: 14px;
    border-radius: var(--radius-sm);
    background: var(--banana-light); border: 1px solid #eee7a1;
  }
  .next-card div { flex: 1; display: flex; flex-direction: column; gap: 2px; }
  .next-card strong { font-family: var(--font-heading); font-size: 14px; }
  .next-card small { font-size: 12px; color: var(--text-muted); }
  .btn-next {
    flex-shrink: 0;
    font-family: var(--font-heading); font-size: 13px; font-weight: 700;
    padding: 10px 12px; border-radius: var(--radius-xs);
    background: var(--sapphire-dark); color: var(--banana);
  }

  .stats {
    display: grid; grid-template-columns: 1fr 1fr;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    overflow: hidden;
  }
  .stat { padding: 12px 14px; display: flex; flex-direction: column; gap: 2px; }
  .stat:nth-child(odd)  { border-right: 1px solid var(--border); }
  .stat:nth-child(-n+2) { border-bottom: 1px solid var(--border); }
  .s-label { font-size: 11px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; }
  .s-val { font-family: var(--font-heading); font-size: 16px; font-weight: 700; }
  .green { color: var(--success); }
  .red   { color: var(--danger); }

  .carry {
    display: flex; align-items: center; gap: 8px;
    font-size: 13px; font-weight: 600;
    padding: 10px 12px;
    border-radius: var(--radius-xs);
    background: var(--success-light); color: var(--success);
  }
  .carry.red { background: var(--danger-light); color: var(--danger); }
  .carry svg { flex-shrink: 0; }

  .periods { display: flex; flex-direction: column; gap: 8px; }

  .center { text-align: center; padding: 48px; color: var(--text-muted); }
  .spinner {
    display: inline-block;
    width: 28px; height: 28px;
    border: 3px solid var(--border);
    border-top-color: var(--sapphire-dark);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  .err { color: var(--danger); }
</style>
