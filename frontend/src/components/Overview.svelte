<script>
  import { fmtIDR, fmtShort, cycleSummary } from '../lib/utils.js'
  import { i18n } from '../lib/i18n.js'

  let { cycles } = $props()

  let stats = $derived.by(() => {
    let totalSaved = 0, totalDeficit = 0, totalBudget = 0, periodsCompleted = 0, sisaCount = 0
    for (const c of cycles) {
      for (const p of c.periods ?? []) {
        totalBudget += p.budget
        if (p.status !== 'completed') continue
        periodsCompleted++
        if (p.result_type === 'sisa') { totalSaved += p.result_amount; sisaCount++ }
        else                          { totalDeficit += p.result_amount }
      }
    }
    const net = totalSaved - totalDeficit
    const savingsRate = periodsCompleted > 0 ? Math.round((sisaCount / periodsCompleted) * 100) : 0
    return { totalSaved, totalDeficit, net, totalBudget, periodsCompleted, savingsRate, sisaCount }
  })

  let chartData = $derived.by(() => {
    return cycles
      .map(c => {
        const s = cycleSummary(c.periods ?? [])
        const done = (c.periods ?? []).filter(p => p.status === 'completed').length
        return { cycle: c, summary: s, done }
      })
      .filter(d => d.done > 0)
      .slice(-8)
  })

  let maxAbs = $derived(
    chartData.length
      ? Math.max(...chartData.map(d => Math.abs(d.summary.net)), 1)
      : 1
  )

  let streak = $derived.by(() => {
    const allCompleted = []
    for (const c of [...cycles].reverse()) {
      for (const p of [...(c.periods ?? [])].reverse()) {
        if (p.status === 'completed') allCompleted.push(p)
      }
    }
    let s = 0
    for (const p of allCompleted) {
      if (p.result_type === 'sisa') s++
      else break
    }
    return s
  })

  function cycleLabel(c) {
    return fmtShort(c.start_date)
  }
</script>

<div class="overview">

  <!-- Hero -->
  <div class="hero" class:hero-positive={stats.net >= 0} class:hero-negative={stats.net < 0}>
    <div class="hero-inner">
      <div class="hero-label">
        {stats.net >= 0 ? $i18n.heroSaved : $i18n.heroBalance}
      </div>
      <div class="hero-amount">
        {stats.net >= 0 ? '' : '-'}Rp {fmtIDR(Math.abs(stats.net))}
      </div>
      <div class="hero-sub">
        {$i18n.fromCycles(cycles.length, stats.periodsCompleted)}
      </div>
    </div>
    {#if streak >= 2}
      <div class="streak-badge">{$i18n.streak(streak)}</div>
    {/if}
  </div>

  <!-- Stats grid -->
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-icon green">↑</div>
      <div class="stat-body">
        <div class="stat-label">{$i18n.statSurplus}</div>
        <div class="stat-val green">Rp {fmtIDR(stats.totalSaved)}</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon red">↓</div>
      <div class="stat-body">
        <div class="stat-label">{$i18n.statDeficit}</div>
        <div class="stat-val red">Rp {fmtIDR(stats.totalDeficit)}</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon blue">%</div>
      <div class="stat-body">
        <div class="stat-label">{$i18n.savingsRate}</div>
        <div class="stat-val blue">{stats.savingsRate}%</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon navy">Σ</div>
      <div class="stat-body">
        <div class="stat-label">{$i18n.budgetMgd}</div>
        <div class="stat-val navy">Rp {fmtIDR(stats.totalBudget)}</div>
      </div>
    </div>
  </div>

  <!-- Savings rate visual -->
  {#if stats.periodsCompleted > 0}
    <div class="rate-card">
      <div class="rate-header">
        <span class="rate-title">{$i18n.consistency}</span>
        <span class="rate-pct" class:good={stats.savingsRate >= 75} class:ok={stats.savingsRate >= 50 && stats.savingsRate < 75} class:bad={stats.savingsRate < 50}>
          {$i18n.perCount(stats.sisaCount, stats.periodsCompleted)}
        </span>
      </div>
      <div class="rate-bar-wrap">
        <div class="rate-bar" style="width:{stats.savingsRate}%"
          class:good={stats.savingsRate >= 75}
          class:ok={stats.savingsRate >= 50 && stats.savingsRate < 75}
          class:bad={stats.savingsRate < 50}
        ></div>
      </div>
      <div class="rate-caption">
        {#if stats.savingsRate >= 75}
          {$i18n.msgGood}
        {:else if stats.savingsRate >= 50}
          {$i18n.msgOk}
        {:else if stats.savingsRate > 0}
          {$i18n.msgBad}
        {/if}
      </div>
    </div>
  {/if}

  <!-- Per-cycle chart -->
  {#if chartData.length > 0}
    <div class="chart-card">
      <div class="chart-title">{$i18n.historyTitle}</div>
      <div class="chart-rows">
        {#each chartData as d}
          {@const pct = Math.abs(d.summary.net) / maxAbs * 100}
          {@const isPos = d.summary.net >= 0}
          <div class="chart-row">
            <div class="chart-label">{cycleLabel(d.cycle)}</div>
            <div class="chart-bar-wrap">
              <div
                class="chart-bar"
                class:pos={isPos}
                class:neg={!isPos}
                style="width: {Math.max(pct, 4)}%"
              ></div>
            </div>
            <div class="chart-amount" class:pos={isPos} class:neg={!isPos}>
              {isPos ? '+' : '-'}Rp {fmtIDR(Math.abs(d.summary.net))}
            </div>
          </div>
        {/each}
      </div>
      <div class="chart-legend">
        <span class="legend-dot pos"></span><span>{$i18n.legendSurplus}</span>
        <span class="legend-dot neg"></span><span>{$i18n.legendDeficit}</span>
      </div>
    </div>
  {/if}

  <!-- Empty state -->
  {#if cycles.length === 0 || stats.periodsCompleted === 0}
    <div class="empty-state">
      <div class="empty-icon">
        <svg width="56" height="56" viewBox="0 0 512 512" fill="none">
          <rect width="512" height="512" rx="110" fill="#154374"/>
          <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
          <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
          <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
          <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
        </svg>
      </div>
      <p class="empty-title">{$i18n.noData}</p>
      <p class="empty-sub">{$i18n.noDataSub}</p>
    </div>
  {/if}

</div>

<style>
  .overview {
    max-width: 480px;
    margin: 0 auto;
    padding: 16px 16px 80px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  /* ── Hero ── */
  .hero {
    border-radius: var(--radius);
    padding: 24px 20px 20px;
    position: relative;
    overflow: hidden;
  }
  .hero-positive {
    background: linear-gradient(135deg, var(--sapphire-dark) 0%, var(--primary) 100%);
  }
  .hero-negative {
    background: linear-gradient(135deg, #7f1d1d 0%, var(--danger) 100%);
  }
  .hero::before {
    content: '';
    position: absolute;
    top: -40px; right: -40px;
    width: 140px; height: 140px;
    border-radius: 50%;
    background: rgba(255,255,255,0.06);
  }
  .hero::after {
    content: '';
    position: absolute;
    bottom: -30px; left: 30px;
    width: 100px; height: 100px;
    border-radius: 50%;
    background: rgba(255,255,255,0.04);
  }
  .hero-inner { position: relative; z-index: 1; }
  .hero-label {
    font-size: 13px;
    font-weight: 600;
    color: rgba(255,255,255,0.75);
    margin-bottom: 6px;
  }
  .hero-amount {
    font-family: var(--font-heading);
    font-size: 32px;
    font-weight: 800;
    line-height: 1.1;
    margin-bottom: 6px;
    letter-spacing: -1px;
  }
  .hero-positive .hero-amount { color: var(--banana); }
  .hero-negative .hero-amount { color: #fca5a5; }
  .hero-sub {
    font-size: 12px;
    color: rgba(255,255,255,0.6);
  }
  .streak-badge {
    position: relative;
    z-index: 1;
    display: inline-block;
    margin-top: 14px;
    background: rgba(242,233,66,0.18);
    border: 1px solid rgba(242,233,66,0.35);
    color: var(--banana);
    font-size: 12px;
    font-weight: 700;
    padding: 5px 12px;
    border-radius: 20px;
  }

  /* ── Stats grid ── */
  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }
  .stat-card {
    background: var(--surface);
    border-radius: var(--radius-sm);
    padding: 14px 12px;
    display: flex;
    align-items: center;
    gap: 10px;
    box-shadow: var(--shadow-sm);
  }
  .stat-icon {
    width: 34px; height: 34px;
    border-radius: var(--radius-xs);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 15px;
    font-weight: 800;
    flex-shrink: 0;
  }
  .stat-icon.green { background: var(--success-light); color: var(--success); }
  .stat-icon.red   { background: var(--danger-light);  color: var(--danger);  }
  .stat-icon.blue  { background: var(--sapphire-light); color: var(--primary); }
  .stat-icon.navy  { background: var(--sapphire-light); color: var(--sapphire-dark); }
  .stat-body { min-width: 0; }
  .stat-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.4px;
    margin-bottom: 2px;
  }
  .stat-val {
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 800;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .stat-val.green { color: var(--success); }
  .stat-val.red   { color: var(--danger); }
  .stat-val.blue  { color: var(--primary); font-size: 18px; }
  .stat-val.navy  { color: var(--sapphire-dark); }

  /* ── Savings rate ── */
  .rate-card {
    background: var(--surface);
    border-radius: var(--radius-sm);
    padding: 16px;
    box-shadow: var(--shadow-sm);
  }
  .rate-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }
  .rate-title {
    font-size: 13px;
    font-weight: 700;
    color: var(--text);
  }
  .rate-pct {
    font-size: 12px;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 20px;
  }
  .rate-pct.good { background: var(--success-light); color: var(--success); }
  .rate-pct.ok   { background: var(--pumpkin-light); color: var(--pumpkin); }
  .rate-pct.bad  { background: var(--danger-light);  color: var(--danger);  }

  .rate-bar-wrap {
    height: 10px;
    background: var(--border);
    border-radius: 5px;
    overflow: hidden;
    margin-bottom: 8px;
  }
  .rate-bar {
    height: 100%;
    border-radius: 5px;
    transition: width 0.6s ease;
  }
  .rate-bar.good { background: linear-gradient(90deg, var(--success), #22c55e); }
  .rate-bar.ok   { background: linear-gradient(90deg, var(--pumpkin), #f59e0b); }
  .rate-bar.bad  { background: linear-gradient(90deg, var(--danger),  #ef4444); }

  .rate-caption {
    font-size: 12px;
    color: var(--text-muted);
  }

  /* ── Chart ── */
  .chart-card {
    background: var(--surface);
    border-radius: var(--radius-sm);
    padding: 16px;
    box-shadow: var(--shadow-sm);
  }
  .chart-title {
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
    color: var(--text);
    margin-bottom: 14px;
  }
  .chart-rows {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .chart-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .chart-label {
    font-size: 11px;
    color: var(--text-muted);
    font-weight: 600;
    width: 52px;
    flex-shrink: 0;
    text-align: right;
  }
  .chart-bar-wrap {
    flex: 1;
    height: 22px;
    background: var(--surface-2);
    border-radius: var(--radius-xs);
    overflow: hidden;
  }
  .chart-bar {
    height: 100%;
    border-radius: var(--radius-xs);
    transition: width 0.5s ease;
    min-width: 4px;
  }
  .chart-bar.pos { background: linear-gradient(90deg, var(--primary), #4fa0f0); }
  .chart-bar.neg { background: linear-gradient(90deg, var(--danger),  #f87171); }

  .chart-amount {
    font-family: var(--font-heading);
    font-size: 11px;
    font-weight: 700;
    width: 70px;
    flex-shrink: 0;
    text-align: right;
    white-space: nowrap;
  }
  .chart-amount.pos { color: var(--primary); }
  .chart-amount.neg { color: var(--danger);  }

  .chart-legend {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 14px;
    font-size: 11px;
    color: var(--text-muted);
  }
  .legend-dot {
    width: 10px; height: 10px;
    border-radius: 2px;
    display: inline-block;
    flex-shrink: 0;
  }
  .legend-dot.pos { background: var(--primary); }
  .legend-dot.neg { background: var(--danger);  }

  /* ── Empty ── */
  .empty-state {
    text-align: center;
    padding: 40px 20px 20px;
    color: var(--text-muted);
  }
  .empty-icon { display: flex; justify-content: center; margin-bottom: 16px; }
  .empty-title {
    font-family: var(--font-heading);
    font-size: 16px;
    font-weight: 700;
    color: var(--text);
    margin-bottom: 8px;
  }
  .empty-sub {
    font-size: 13px;
    line-height: 1.5;
    max-width: 260px;
    margin: 0 auto;
  }
</style>
