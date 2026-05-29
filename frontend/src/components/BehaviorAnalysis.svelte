<script>
  import brain from 'brain.js'
  import { i18n } from '../lib/i18n.js'

  let { cycles } = $props()

  // ── Feature extraction ────────────────────────────────────────────────────

  function cycleRate(c) {
    const cp = (c.periods ?? []).filter(p => p.status === 'completed')
    if (!cp.length) return null
    return cp.filter(p => p.result_type === 'sisa').length / cp.length
  }

  function extractFeatures(doneCycles) {
    const allCompleted = []
    for (const c of doneCycles) {
      for (const p of (c.periods ?? [])) {
        if (p.status === 'completed') allCompleted.push(p)
      }
    }
    if (!allCompleted.length) return null

    const total      = allCompleted.length
    const sisaCount  = allCompleted.filter(p => p.result_type === 'sisa').length
    const savingsRate = sisaCount / total

    // Early vs late split based on period_number
    const maxPos  = Math.max(...allCompleted.map(p => p.period_number))
    const mid     = Math.ceil(maxPos / 2)
    const early   = allCompleted.filter(p => p.period_number <= mid)
    const late    = allCompleted.filter(p => p.period_number > mid)
    const earlySuccess = early.length ? early.filter(p => p.result_type === 'sisa').length / early.length : 0.5
    const lateSuccess  = late.length  ? late.filter(p => p.result_type === 'sisa').length  / late.length  : 0.5

    // Volatility: std-dev of per-cycle savings rates, normalised to [0,1]
    const rates = doneCycles.map(cycleRate).filter(r => r !== null)
    let volatility = 0
    if (rates.length > 1) {
      const mean = rates.reduce((a, b) => a + b, 0) / rates.length
      const variance = rates.reduce((s, r) => s + (r - mean) ** 2, 0) / rates.length
      volatility = Math.min(Math.sqrt(variance) * 2, 1)
    }

    // Trend slope over last 5 cycles, normalised to [0,1] (0.5 = flat)
    const recent = rates.slice(-5)
    let trendSlope = 0.5
    if (recent.length >= 2) {
      const n = recent.length
      const xm = (n - 1) / 2
      const ym = recent.reduce((a, b) => a + b, 0) / n
      let num = 0, den = 0
      for (let i = 0; i < n; i++) { num += (i - xm) * (recent[i] - ym); den += (i - xm) ** 2 }
      const slope = den ? num / den : 0
      trendSlope = Math.min(Math.max((slope + 0.5), 0), 1)
    }

    return { savingsRate, earlySuccess, lateSuccess, volatility, trendSlope }
  }

  // ── brain.js archetype classifier ─────────────────────────────────────────

  // Synthetic archetype prototypes used as training set.
  // The network learns a continuous manifold between them; user data maps to nearest.
  const TRAINING = [
    { input: { savingsRate:0.95, earlySuccess:0.95, lateSuccess:0.90, volatility:0.05, trendSlope:0.50 }, output: { consistent_saver:1, improving:0, late_rusher:0, volatile:0 } },
    { input: { savingsRate:0.85, earlySuccess:0.85, lateSuccess:0.82, volatility:0.08, trendSlope:0.52 }, output: { consistent_saver:1, improving:0, late_rusher:0, volatile:0 } },
    { input: { savingsRate:0.80, earlySuccess:0.80, lateSuccess:0.78, volatility:0.10, trendSlope:0.55 }, output: { consistent_saver:1, improving:0, late_rusher:0, volatile:0 } },
    { input: { savingsRate:0.55, earlySuccess:0.55, lateSuccess:0.50, volatility:0.20, trendSlope:0.80 }, output: { consistent_saver:0, improving:1, late_rusher:0, volatile:0 } },
    { input: { savingsRate:0.40, earlySuccess:0.38, lateSuccess:0.42, volatility:0.15, trendSlope:0.92 }, output: { consistent_saver:0, improving:1, late_rusher:0, volatile:0 } },
    { input: { savingsRate:0.65, earlySuccess:0.62, lateSuccess:0.60, volatility:0.18, trendSlope:0.78 }, output: { consistent_saver:0, improving:1, late_rusher:0, volatile:0 } },
    { input: { savingsRate:0.55, earlySuccess:0.92, lateSuccess:0.10, volatility:0.22, trendSlope:0.50 }, output: { consistent_saver:0, improving:0, late_rusher:1, volatile:0 } },
    { input: { savingsRate:0.50, earlySuccess:0.88, lateSuccess:0.12, volatility:0.25, trendSlope:0.48 }, output: { consistent_saver:0, improving:0, late_rusher:1, volatile:0 } },
    { input: { savingsRate:0.60, earlySuccess:0.95, lateSuccess:0.18, volatility:0.30, trendSlope:0.45 }, output: { consistent_saver:0, improving:0, late_rusher:1, volatile:0 } },
    { input: { savingsRate:0.50, earlySuccess:0.50, lateSuccess:0.50, volatility:0.80, trendSlope:0.50 }, output: { consistent_saver:0, improving:0, late_rusher:0, volatile:1 } },
    { input: { savingsRate:0.45, earlySuccess:0.40, lateSuccess:0.55, volatility:0.85, trendSlope:0.42 }, output: { consistent_saver:0, improving:0, late_rusher:0, volatile:1 } },
    { input: { savingsRate:0.60, earlySuccess:0.65, lateSuccess:0.52, volatility:0.78, trendSlope:0.45 }, output: { consistent_saver:0, improving:0, late_rusher:0, volatile:1 } },
  ]

  // Lazy-initialised trained network (shared across renders)
  let _net = null
  function getNet() {
    if (_net) return _net
    _net = new brain.NeuralNetwork({ hiddenLayers: [8, 4], activation: 'sigmoid' })
    _net.train(TRAINING, { iterations: 2000, errorThresh: 0.005, log: false })
    return _net
  }

  function classifyArchetype(features) {
    try {
      const scores = getNet().run(features)
      let best = 'consistent_saver', bestScore = 0
      for (const [k, v] of Object.entries(scores)) {
        if (v > bestScore) { bestScore = v; best = k }
      }
      return { key: best, confidence: Math.round(bestScore * 100), scores }
    } catch { return null }
  }

  // ── Period vulnerability ──────────────────────────────────────────────────

  function periodVulnerability(doneCycles) {
    const byPos = {}
    for (const c of doneCycles) {
      for (const p of (c.periods ?? [])) {
        if (p.status !== 'completed') continue
        const n = p.period_number
        if (!byPos[n]) byPos[n] = { total: 0, deficits: 0 }
        byPos[n].total++
        if (p.result_type === 'defisit') byPos[n].deficits++
      }
    }
    return Object.entries(byPos)
      .map(([pos, d]) => ({
        pos: Number(pos),
        label: `P${pos}`,
        total: d.total,
        deficits: d.deficits,
        rate: d.deficits / d.total,
      }))
      .sort((a, b) => a.pos - b.pos)
  }

  // ── Sliding-window pattern scan (CNN-inspired 1D convolution) ────────────

  function scanPatterns(doneCycles) {
    // Build chronological sequence: +1 = sisa, -1 = defisit
    const seq = []
    for (const c of doneCycles) {
      const sorted = [...(c.periods ?? [])].sort((a, b) => a.period_number - b.period_number)
      for (const p of sorted) {
        if (p.status === 'completed') seq.push(p.result_type === 'sisa' ? 1 : -1)
      }
    }
    if (seq.length < 3) return null

    // Kernels (like 1D conv filters)
    const kernels = {
      streak_surplus: [1, 1, 1],
      streak_deficit: [-1, -1, -1],
      recovery:       [-1, -1, 1],
      late_slide:     [1, -1, -1],
    }
    const counts = { streak_surplus: 0, streak_deficit: 0, recovery: 0, late_slide: 0 }
    for (let i = 0; i <= seq.length - 3; i++) {
      for (const [name, k] of Object.entries(kernels)) {
        if (k.every((v, j) => v === seq[i + j])) counts[name]++
      }
    }
    const total = Object.values(counts).reduce((a, b) => a + b, 0) || 1
    return Object.entries(counts)
      .map(([name, count]) => ({ name, count, pct: Math.round(count / total * 100) }))
      .sort((a, b) => b.count - a.count)
  }

  // ── Momentum ─────────────────────────────────────────────────────────────

  function calcMomentum(doneCycles) {
    const rates = doneCycles.map(cycleRate).filter(r => r !== null).slice(-5)
    if (rates.length < 2) return null
    const n = rates.length, xm = (n - 1) / 2
    const ym = rates.reduce((a, b) => a + b, 0) / n
    let num = 0, den = 0
    for (let i = 0; i < n; i++) { num += (i - xm) * (rates[i] - ym); den += (i - xm) ** 2 }
    const slope = den ? num / den : 0
    return {
      rates,
      direction: slope > 0.05 ? 'up' : slope < -0.05 ? 'down' : 'stable',
      slope,
    }
  }

  // ── Main analysis (all derived together) ─────────────────────────────────

  let analysis = $derived.by(() => {
    const done = cycles.filter(c => (c.periods ?? []).some(p => p.status === 'completed'))
    if (done.length < 2) return null

    const features    = extractFeatures(done)
    if (!features) return null

    const vuln        = periodVulnerability(done)
    const patterns    = scanPatterns(done)
    const momentum    = calcMomentum(done)
    const archetype   = done.length >= 3 ? classifyArchetype(features) : null

    // Build insight list
    const insights = []
    const worst = vuln.length ? vuln.reduce((a, b) => a.rate > b.rate ? a : b) : null
    if (worst && worst.rate >= 0.5 && worst.total >= 2) {
      insights.push({ key: 'weak_period', p: worst.label, r: Math.round(worst.rate * 100) })
    }
    if (momentum?.direction === 'up')   insights.push({ key: 'improving' })
    if (momentum?.direction === 'down') insights.push({ key: 'declining' })
    if (features.earlySuccess - features.lateSuccess > 0.3)  insights.push({ key: 'late_rusher' })
    if (features.lateSuccess  - features.earlySuccess > 0.3) insights.push({ key: 'early_rusher' })

    return { features, archetype, vuln, patterns, momentum, insights, done }
  })

  // ── Archetype meta ────────────────────────────────────────────────────────

  const ARCHETYPE_META = {
    consistent_saver: { icon: '⭐', colorClass: 'green' },
    improving:        { icon: '📈', colorClass: 'blue'  },
    late_rusher:      { icon: '⚡', colorClass: 'orange'},
    volatile:         { icon: '🎲', colorClass: 'red'   },
  }

  function archetypeName(key, t) {
    return { consistent_saver: t.archetypeConsistent, improving: t.archetypeImproving,
             late_rusher: t.archetypeLate, volatile: t.archetypeVolatile }[key] ?? key
  }
  function archetypeDesc(key, t) {
    return { consistent_saver: t.archetypeConsistentDesc, improving: t.archetypeImprovingDesc,
             late_rusher: t.archetypeLateDesc, volatile: t.archetypeVolatileDesc }[key] ?? ''
  }
  function patternLabel(name, t) {
    return { streak_surplus: t.patternStreakSurplus, streak_deficit: t.patternStreakDeficit,
             recovery: t.patternRecovery, late_slide: t.patternLateSlide }[name] ?? name
  }
  function insightText(item, t) {
    if (item.key === 'weak_period') return t.insightWeakPeriod(item.p, item.r)
    if (item.key === 'improving')   return t.insightImproving
    if (item.key === 'declining')   return t.insightDeclining
    if (item.key === 'late_rusher') return t.insightLateRusher
    if (item.key === 'early_rusher')return t.insightEarlyRusher
    return ''
  }

  // Sparkline SVG path from momentum rates
  function sparklinePath(rates) {
    if (!rates?.length) return ''
    const w = 80, h = 28, pad = 2
    const min = Math.min(...rates), max = Math.max(...rates)
    const range = max - min || 0.01
    const pts = rates.map((r, i) => {
      const x = pad + (i / (rates.length - 1)) * (w - pad * 2)
      const y = h - pad - ((r - min) / range) * (h - pad * 2)
      return `${x},${y}`
    })
    return 'M' + pts.join('L')
  }
</script>

{#if analysis === null}
  {@const done = cycles.filter(c => (c.periods ?? []).some(p => p.status === 'completed'))}
  {#if done.length < 2 && cycles.length > 0}
    <div class="ba-card ba-locked">
      <div class="ba-header">
        <span class="ba-title">{$i18n.behaviorTitle}</span>
        <span class="ai-chip">AI</span>
      </div>
      <p class="ba-min-data">{$i18n.behaviorMinData}</p>
    </div>
  {/if}
{:else}
  {@const t = $i18n}
  <div class="ba-card">
    <div class="ba-header">
      <span class="ba-title">{t.behaviorTitle}</span>
      <span class="ai-chip">AI</span>
    </div>
    <p class="ba-subtitle">{t.behaviorSubtitle}</p>

    <!-- Archetype (brain.js, 3+ cycles) -->
    {#if analysis.archetype}
      {@const meta = ARCHETYPE_META[analysis.archetype.key] ?? ARCHETYPE_META.volatile}
      <div class="archetype-row {meta.colorClass}">
        <div class="arch-icon">{meta.icon}</div>
        <div class="arch-body">
          <div class="arch-label">{t.archetypeLabel}</div>
          <div class="arch-name">{archetypeName(analysis.archetype.key, t)}</div>
          <div class="arch-desc">{archetypeDesc(analysis.archetype.key, t)}</div>
          <div class="arch-conf-wrap">
            <div class="arch-conf-bar" style="width:{analysis.archetype.confidence}%"></div>
          </div>
          <div class="arch-conf-label">{t.archetypeConfidence(analysis.archetype.confidence)}</div>
        </div>
      </div>
    {/if}

    <!-- Period vulnerability -->
    {#if analysis.vuln.length > 0}
      <div class="section">
        <div class="section-title">{t.periodVulnTitle}</div>
        <div class="vuln-bars">
          {#each analysis.vuln as v}
            {@const isWorst = analysis.vuln.length > 1 && v.rate === Math.max(...analysis.vuln.map(x => x.rate)) && v.rate > 0}
            <div class="vuln-row">
              <div class="vuln-label" class:worst={isWorst}>{v.label}</div>
              <div class="vuln-track">
                <div class="vuln-fill" class:danger={v.rate >= 0.5} style="width:{Math.round(v.rate * 100)}%"></div>
              </div>
              <div class="vuln-pct" class:danger={v.rate >= 0.5}>{Math.round(v.rate * 100)}%</div>
              {#if isWorst}<span class="worst-badge">{t.periodVulnHardest}</span>{/if}
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Momentum sparkline -->
    {#if analysis.momentum}
      <div class="section momentum-row">
        <div class="section-title">{t.momentumLabel}</div>
        <div class="momentum-body">
          <svg class="sparkline" width="80" height="28" viewBox="0 0 80 28">
            <path d={sparklinePath(analysis.momentum.rates)}
              fill="none"
              stroke={analysis.momentum.direction === 'up' ? 'var(--success)' : analysis.momentum.direction === 'down' ? 'var(--danger)' : 'var(--primary)'}
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span class="momentum-dir"
            class:up={analysis.momentum.direction === 'up'}
            class:down={analysis.momentum.direction === 'down'}
            class:stable={analysis.momentum.direction === 'stable'}>
            {analysis.momentum.direction === 'up' ? t.momentumUp : analysis.momentum.direction === 'down' ? t.momentumDown : t.momentumStable}
          </span>
        </div>
      </div>
    {/if}

    <!-- Pattern scan -->
    {#if analysis.patterns}
      <div class="section">
        <div class="section-title">{t.patternTitle}</div>
        <div class="patterns">
          {#each analysis.patterns.filter(p => p.count > 0) as p}
            <div class="pattern-row">
              <span class="pattern-name">{patternLabel(p.name, t)}</span>
              <div class="pattern-track">
                <div class="pattern-fill
                  {p.name === 'streak_surplus' || p.name === 'recovery' ? 'pos' : 'neg'}"
                  style="width:{p.pct}%"></div>
              </div>
              <span class="pattern-pct">{p.pct}%</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Insights -->
    {#if analysis.insights.length > 0}
      <div class="section">
        <div class="section-title">{t.insightsLabel}</div>
        <ul class="insights">
          {#each analysis.insights as item}
            <li>{insightText(item, t)}</li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}

<style>
  .ba-card {
    background: var(--surface);
    border-radius: var(--radius-sm);
    padding: 18px 16px;
    box-shadow: var(--shadow-sm);
    border-top: 3px solid var(--sapphire-dark);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .ba-locked { border-top-color: var(--border); }

  .ba-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .ba-title {
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
    color: var(--text);
  }
  .ai-chip {
    font-family: var(--font-heading);
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.5px;
    padding: 2px 6px;
    border-radius: 20px;
    background: var(--sapphire-dark);
    color: var(--banana);
  }
  .ba-subtitle {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: -8px;
  }
  .ba-min-data { font-size: 12px; color: var(--text-muted); }

  /* ── Archetype ── */
  .archetype-row {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    padding: 14px;
    border-radius: var(--radius-xs);
    background: var(--surface-2);
  }
  .archetype-row.green  { background: var(--success-light); }
  .archetype-row.blue   { background: var(--sapphire-light); }
  .archetype-row.orange { background: var(--pumpkin-light); }
  .archetype-row.red    { background: var(--danger-light); }

  .arch-icon {
    font-size: 28px;
    line-height: 1;
    flex-shrink: 0;
  }
  .arch-body { flex: 1; min-width: 0; }
  .arch-label {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.6px;
    color: var(--text-muted);
    margin-bottom: 2px;
  }
  .arch-name {
    font-family: var(--font-heading);
    font-size: 16px;
    font-weight: 800;
    color: var(--text);
    margin-bottom: 4px;
  }
  .arch-desc { font-size: 12px; color: var(--text-muted); margin-bottom: 10px; line-height: 1.4; }
  .arch-conf-wrap {
    height: 5px;
    background: rgba(0,0,0,0.08);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 4px;
  }
  .arch-conf-bar {
    height: 100%;
    background: var(--sapphire-dark);
    border-radius: 3px;
    transition: width 0.6s ease;
  }
  .archetype-row.green  .arch-conf-bar { background: var(--success); }
  .archetype-row.blue   .arch-conf-bar { background: var(--primary); }
  .archetype-row.orange .arch-conf-bar { background: var(--pumpkin); }
  .archetype-row.red    .arch-conf-bar { background: var(--danger); }
  .arch-conf-label { font-size: 11px; color: var(--text-muted); }

  /* ── Sections ── */
  .section { display: flex; flex-direction: column; gap: 8px; }
  .section-title {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }

  /* ── Vulnerability ── */
  .vuln-bars { display: flex; flex-direction: column; gap: 6px; }
  .vuln-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .vuln-label {
    font-size: 11px;
    font-weight: 700;
    color: var(--text-muted);
    width: 24px;
    flex-shrink: 0;
  }
  .vuln-label.worst { color: var(--danger); }
  .vuln-track {
    flex: 1;
    height: 8px;
    background: var(--surface-2);
    border-radius: 4px;
    overflow: hidden;
  }
  .vuln-fill {
    height: 100%;
    background: var(--sapphire-light);
    border-radius: 4px;
    transition: width 0.5s ease;
  }
  .vuln-fill.danger { background: var(--danger-light); }
  .vuln-pct {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    width: 30px;
    text-align: right;
    flex-shrink: 0;
  }
  .vuln-pct.danger { color: var(--danger); }
  .worst-badge {
    font-size: 10px;
    font-weight: 700;
    padding: 1px 6px;
    border-radius: 20px;
    background: var(--danger-light);
    color: var(--danger);
    flex-shrink: 0;
  }

  /* ── Momentum ── */
  .momentum-row { flex-direction: row; align-items: center; justify-content: space-between; }
  .momentum-body { display: flex; align-items: center; gap: 10px; }
  .sparkline { display: block; }
  .momentum-dir {
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
  }
  .momentum-dir.up     { color: var(--success); }
  .momentum-dir.down   { color: var(--danger); }
  .momentum-dir.stable { color: var(--primary); }

  /* ── Patterns ── */
  .patterns { display: flex; flex-direction: column; gap: 5px; }
  .pattern-row { display: flex; align-items: center; gap: 6px; }
  .pattern-name {
    font-size: 11px;
    color: var(--text-muted);
    width: 110px;
    flex-shrink: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .pattern-track {
    flex: 1;
    height: 6px;
    background: var(--surface-2);
    border-radius: 3px;
    overflow: hidden;
  }
  .pattern-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.5s ease;
  }
  .pattern-fill.pos { background: var(--primary); }
  .pattern-fill.neg { background: var(--danger); }
  .pattern-pct { font-size: 11px; color: var(--text-muted); width: 30px; text-align: right; flex-shrink: 0; }

  /* ── Insights ── */
  .insights {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 0;
    margin: 0;
  }
  .insights li {
    font-size: 12px;
    color: var(--text-muted);
    line-height: 1.5;
    padding-left: 14px;
    position: relative;
  }
  .insights li::before {
    content: '→';
    position: absolute;
    left: 0;
    color: var(--primary);
    font-weight: 700;
  }
</style>
