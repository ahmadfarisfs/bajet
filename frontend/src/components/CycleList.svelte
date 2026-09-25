<script>
  import {
    fmtShort, fmtIDR, cycleSummary, isActive, activePeriod, daysLeft,
    adjustedBudgets, overduePeriods, needsNextCycle, latestCycle, nextCycleDraft,
  } from '../lib/utils.js'
  import { i18n } from '../lib/i18n.js'
  import PeriodStrip from './PeriodStrip.svelte'

  let { cycles, loading = false, onSelect, onNew, onStartNext } = $props()

  let activeCycle = $derived(cycles.find(c => isActive(c.start_date, c.end_date)) ?? null)
  let hero = $derived.by(() => {
    if (!activeCycle) return null
    const cp = activePeriod(activeCycle.periods ?? [])
    if (!cp) return null
    const adj = adjustedBudgets(activeCycle.periods).byId.get(cp.id)
    return { cycle: activeCycle, period: cp, budget: adj ?? cp.budget, adjusted: adj != null && adj !== cp.budget }
  })

  let overdue = $derived(overduePeriods(cycles))
  let nextDraft = $derived(needsNextCycle(cycles) ? nextCycleDraft(latestCycle(cycles)) : null)

  function completedCount(cycle) { return (cycle.periods ?? []).filter(p => p.status === 'completed').length }
</script>

<div class="page">
  {#if hero}
    {@const left = daysLeft(hero.period.end_date)}
    <button class="hero" onclick={() => onSelect(hero.cycle.id)}>
      <div class="hero-top">
        <span class="hero-chip">P{hero.period.period_number}</span>
        <span class="hero-days" class:urgent={left <= 1}>{$i18n.daysLeft(left)}</span>
      </div>
      {#if hero.period.status === 'completed'}
        <div class="hero-label">{$i18n.periodCheckedIn}</div>
        <div class="hero-amount num">
          {hero.period.result_type === 'sisa' ? '+' : '−'}Rp {fmtIDR(hero.period.result_amount)}
        </div>
      {:else}
        <div class="hero-label">{$i18n.spendThisPeriod}</div>
        <div class="hero-amount num">Rp {fmtIDR(hero.budget)}</div>
        {#if hero.adjusted}
          <div class="hero-planned num">{$i18n.planned} Rp {fmtIDR(hero.period.budget)}</div>
        {/if}
      {/if}
      <PeriodStrip periods={hero.cycle.periods ?? []} tone="dark" />
      <div class="hero-foot">
        <span>{fmtShort(hero.cycle.start_date)} – {fmtShort(hero.cycle.end_date)}</span>
        <span class="num">{completedCount(hero.cycle)}/{(hero.cycle.periods ?? []).length} {$i18n.periods}</span>
      </div>
    </button>
  {/if}

  {#if overdue.length}
    {@const first = overdue[0]}
    <button class="notice warn" onclick={() => onSelect(first.cycle.id)}>
      <span class="notice-icon" aria-hidden="true">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>
      </span>
      <span class="notice-text">
        <strong>{$i18n.overdueTitle(overdue.length)}</strong>
        <small>{$i18n.overdueSub(first.period.period_number, $i18n.endedAgo(-daysLeft(first.period.end_date)))}</small>
      </span>
      <span class="chev" aria-hidden="true">›</span>
    </button>
  {/if}

  {#if nextDraft}
    <div class="notice next">
      <span class="notice-text">
        <strong>{$i18n.noActiveCycle}</strong>
        <small>{$i18n.nextCycleDates(`${fmtShort(nextDraft.start_date)} – ${fmtShort(nextDraft.end_date)}`)} · Rp {fmtIDR(nextDraft.total_budget)}</small>
      </span>
      <button class="btn-next" onclick={() => onStartNext(latestCycle(cycles))}>{$i18n.startNextCycle}</button>
    </div>
  {/if}

  <div class="section-head">
    <h2>{$i18n.myCycles}</h2>
    <button class="btn-new" onclick={onNew}>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><path d="M12 5v14M5 12h14"/></svg>
      {$i18n.newBtn}
    </button>
  </div>

  {#if cycles.length === 0}
    {#if loading}
      <div class="skeleton"></div>
      <div class="skeleton"></div>
    {:else}
      <div class="empty">
        <svg width="56" height="56" viewBox="0 0 512 512" fill="none" aria-hidden="true">
          <rect width="512" height="512" rx="110" fill="#154374"/>
          <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
          <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
          <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
        </svg>
        <p>{$i18n.noCycles}</p>
        <button class="btn-start" onclick={onNew}>{$i18n.createFirst}</button>
      </div>
    {/if}
  {:else}
    <div class="cards">
      {#each cycles as cycle (cycle.id)}
        {@const s = cycleSummary(cycle.periods ?? [])}
        {@const done = completedCount(cycle)}
        {@const total = (cycle.periods ?? []).length}
        {@const active = isActive(cycle.start_date, cycle.end_date)}
        {@const late = (cycle.periods ?? []).some(p => overdue.some(o => o.period.id === p.id))}
        <button class="card" class:active onclick={() => onSelect(cycle.id)}>
          <div class="card-top">
            <span class="dates">{fmtShort(cycle.start_date)} – {fmtShort(cycle.end_date)}</span>
            {#if active}<span class="badge badge-active">{$i18n.active}</span>{/if}
            {#if late}<span class="badge badge-late">{$i18n.overdueBadge}</span>{/if}
            <span class="count num">{done}/{total}</span>
          </div>
          <PeriodStrip periods={cycle.periods ?? []} />
          <div class="card-foot">
            <span class="budget num">Rp {fmtIDR(cycle.total_budget)}</span>
            {#if done > 0}
              <span class="pill num" class:green={s.net >= 0} class:red={s.net < 0}>
                {$i18n.net} {s.net >= 0 ? '+' : '−'}Rp {fmtIDR(Math.abs(s.net))}
              </span>
            {/if}
          </div>
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    padding: 16px 16px calc(var(--tabbar-h) + var(--safe-bottom) + 32px);
    max-width: var(--page-max);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  /* ── Hero ── */
  .hero {
    text-align: left;
    color: #fff;
    background:
      radial-gradient(120% 90% at 100% 0%, rgba(242,233,66,0.16), transparent 55%),
      linear-gradient(160deg, var(--sapphire-dark), var(--sapphire-deep));
    border-radius: var(--radius);
    padding: 18px 18px 16px;
    box-shadow: var(--shadow-lg);
    display: flex;
    flex-direction: column;
    gap: 6px;
    transition: transform 0.12s;
  }
  .hero:active { transform: scale(0.99); }
  .hero-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
  .hero-chip {
    font-family: var(--font-heading);
    font-weight: 800; font-size: 12px;
    background: var(--banana); color: var(--sapphire-deep);
    padding: 3px 10px; border-radius: 999px;
  }
  .hero-days { font-size: 13px; font-weight: 600; color: rgba(255,255,255,0.8); }
  .hero-days.urgent { color: var(--banana); }
  .hero-label { font-size: 13px; color: rgba(255,255,255,0.7); }
  .hero-amount {
    font-family: var(--font-heading);
    font-size: 36px; font-weight: 800;
    letter-spacing: -1px; line-height: 1.05;
  }
  .hero-planned { font-size: 12px; color: rgba(255,255,255,0.6); text-decoration: line-through; margin-bottom: 4px; }
  .hero :global(.strip) { margin-top: 10px; }
  .hero-foot {
    display: flex; justify-content: space-between;
    font-size: 12px; color: rgba(255,255,255,0.65);
    margin-top: 4px;
  }

  /* ── Notices ── */
  .notice {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    text-align: left;
    padding: 12px 14px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: var(--surface);
  }
  .notice.warn { background: var(--pumpkin-light); border-color: #f7cfa6; color: #8a3f00; }
  .notice.next { background: var(--banana-light); border-color: #eee7a1; }
  .notice-icon {
    flex-shrink: 0;
    width: 32px; height: 32px; border-radius: 50%;
    display: grid; place-items: center;
    background: var(--pumpkin); color: #fff;
  }
  .notice-text { flex: 1; display: flex; flex-direction: column; gap: 1px; min-width: 0; }
  .notice-text strong { font-family: var(--font-heading); font-size: 14px; font-weight: 700; color: var(--text); }
  .notice.warn .notice-text strong { color: #7a3700; }
  .notice-text small { font-size: 12px; color: var(--text-muted); }
  .notice.warn .notice-text small { color: #9a4a08; }
  .chev { font-size: 22px; line-height: 1; opacity: 0.6; }
  .btn-next {
    flex-shrink: 0;
    font-family: var(--font-heading);
    font-size: 13px; font-weight: 700;
    padding: 9px 12px;
    border-radius: var(--radius-xs);
    background: var(--sapphire-dark); color: var(--banana);
  }

  /* ── Section ── */
  .section-head {
    display: flex; justify-content: space-between; align-items: center;
    margin-top: 8px;
  }
  h2 { font-family: var(--font-heading); font-size: 20px; font-weight: 800; letter-spacing: -0.3px; }
  .btn-new {
    display: inline-flex; align-items: center; gap: 6px;
    font-family: var(--font-heading); font-size: 13px; font-weight: 700;
    padding: 8px 14px; border-radius: 999px;
    background: var(--surface); border: 1px solid var(--border-strong);
    color: var(--sapphire-dark);
    transition: background 0.15s;
  }
  .btn-new:hover { background: var(--sapphire-light); }

  /* ── Cards ── */
  .cards { display: flex; flex-direction: column; gap: 10px; }
  .card {
    width: 100%;
    text-align: left;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 14px 16px;
    display: flex; flex-direction: column; gap: 12px;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .card:hover { border-color: var(--border-strong); box-shadow: var(--shadow); }
  .card.active { border-color: var(--sapphire); box-shadow: 0 0 0 3px var(--sapphire-light); }
  .card-top { display: flex; align-items: center; gap: 8px; }
  .dates { font-family: var(--font-heading); font-size: 15px; font-weight: 700; }
  .count { margin-left: auto; font-size: 13px; font-weight: 600; color: var(--text-muted); }
  .badge {
    font-size: 10px; font-weight: 800; letter-spacing: 0.6px; text-transform: uppercase;
    padding: 3px 8px; border-radius: 999px;
  }
  .badge-active { background: var(--sapphire-dark); color: var(--banana); }
  .badge-late   { background: var(--pumpkin-light); color: var(--pumpkin); }
  .card-foot { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
  .budget { font-size: 13px; color: var(--text-muted); font-weight: 600; }
  .pill { font-size: 12px; font-weight: 700; padding: 3px 9px; border-radius: 999px; }
  .pill.green { background: var(--success-light); color: var(--success); }
  .pill.red   { background: var(--danger-light);  color: var(--danger); }

  /* ── Empty / loading ── */
  .skeleton {
    height: 96px;
    border-radius: var(--radius-sm);
    background: linear-gradient(90deg, var(--surface-2), var(--surface), var(--surface-2));
    background-size: 200% 100%;
    animation: shimmer 1.2s ease-in-out infinite;
  }
  @keyframes shimmer { to { background-position: -200% 0; } }
  .empty {
    display: flex; flex-direction: column; align-items: center; gap: 14px;
    text-align: center;
    padding: 48px 20px;
    background: var(--surface);
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius);
    color: var(--text-muted);
  }
  .empty p { font-size: 15px; font-weight: 500; }
  .btn-start {
    font-family: var(--font-heading); font-size: 15px; font-weight: 700;
    padding: 13px 24px; border-radius: var(--radius-sm);
    background: var(--sapphire-dark); color: var(--banana);
  }
</style>
