<script>
  // Approximate local prayer times [hour, minute]
  const PT    = [[4,30],[12,0],[15,15],[18,15],[19,30]]
  const NAMES = ['Subuh','Dzuhur','Ashar','Maghrib','Isya']
  const TOTAL = 5
  const KEY   = 'bajet_prayer'

  function todayStr() {
    const d = new Date()
    return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
  }
  function load() {
    try {
      const r = JSON.parse(localStorage.getItem(KEY) || 'null')
      return r?.date === todayStr() ? (r.count ?? 0) : 0
    } catch { return 0 }
  }
  function persist(n) {
    localStorage.setItem(KEY, JSON.stringify({ date: todayStr(), count: n }))
  }
  function calcExpected() {
    const d = new Date(), h = d.getHours(), m = d.getMinutes()
    return PT.filter(([ph, pm]) => h > ph || (h === ph && m >= pm)).length
  }
  function calcNext() {
    const d = new Date(), h = d.getHours(), m = d.getMinutes()
    return PT.findIndex(([ph, pm]) => h < ph || (h === ph && m < pm))
  }

  let count = $state(load())
  let exp   = $state(calcExpected())
  let nxt   = $state(calcNext())

  $effect(() => {
    const id = setInterval(() => { exp = calcExpected(); nxt = calcNext() }, 60_000)
    return () => clearInterval(id)
  })

  function add()  { if (count < TOTAL) { count++; persist(count) } }
  function undo() { if (count > 0)     { count--; persist(count) } }

  const R    = 26
  const CIRC = 2 * Math.PI * R

  let dash   = $derived(CIRC * (1 - count / TOTAL))
  let status = $derived(
    count >= TOTAL ? 'done'   :
    count >= exp   ? 'ok'     :
    count >= exp-1 ? 'warn'   : 'behind'
  )
  let nextName = $derived(nxt >= 0 && count < TOTAL ? NAMES[nxt] : null)

  let statusText = $derived(
    status === 'done'   ? 'Alhamdulillah 🤲' :
    status === 'ok'     ? 'Tepat waktu' :
    status === 'warn'   ? 'Hampir tercapai' :
                          `${exp - count} perlu dikejar`
  )
</script>

<div class="prayer-card" data-status={status}>
  <div class="prayer-top">
    <!-- Donut ring -->
    <div class="ring-wrap" role="img" aria-label="{count} dari {TOTAL} sholat">
      <svg width="68" height="68" viewBox="0 0 64 64" aria-hidden="true">
        <circle cx="32" cy="32" r={R} fill="none" stroke="var(--border)" stroke-width="7"/>
        <circle
          cx="32" cy="32" r={R}
          fill="none"
          class="arc"
          stroke-width="7"
          stroke-linecap="round"
          stroke-dasharray="{CIRC}"
          stroke-dashoffset="{dash}"
          transform="rotate(-90 32 32)"
        />
      </svg>
      <div class="ring-inner">
        <span class="ring-n">{count}</span><span class="ring-d">/{TOTAL}</span>
      </div>
    </div>

    <!-- Status text -->
    <div class="prayer-info">
      <span class="prayer-label">Sholat Hari Ini</span>
      <span class="prayer-status">{statusText}</span>
      {#if nextName}
        <span class="prayer-next">Selanjutnya: {nextName}</span>
      {/if}
    </div>
  </div>

  {#if count < TOTAL}
    <button class="prayer-btn" onclick={add}>
      Sholat selesai &nbsp;+1
    </button>
    {#if count > 0}
      <button class="undo-btn" onclick={undo}>Batalkan</button>
    {/if}
  {:else}
    <div class="prayer-complete">✓&nbsp; Semua 5 sholat hari ini selesai</div>
  {/if}
</div>

<style>
  .prayer-card {
    background: var(--surface);
    border-radius: var(--radius);
    padding: 16px 16px 14px;
    margin-bottom: 18px;
    box-shadow: var(--shadow-sm);
    border-left: 4px solid var(--border);
    transition: border-left-color 0.3s;
  }
  .prayer-card[data-status="done"]   { border-left-color: var(--sapphire-dark); }
  .prayer-card[data-status="ok"]     { border-left-color: var(--success); }
  .prayer-card[data-status="warn"]   { border-left-color: var(--pumpkin); }
  .prayer-card[data-status="behind"] { border-left-color: var(--danger); }

  .prayer-top {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 14px;
  }

  .ring-wrap {
    position: relative;
    flex-shrink: 0;
    width: 68px;
    height: 68px;
  }
  .ring-inner {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .ring-n {
    font-family: var(--font-heading);
    font-size: 22px;
    font-weight: 800;
    color: var(--text);
    line-height: 1;
  }
  .ring-d {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-muted);
    margin-top: 5px;
    line-height: 1;
  }
  .arc {
    transition: stroke-dashoffset 0.55s cubic-bezier(.4,0,.2,1), stroke 0.3s;
  }
  [data-status="done"]   .arc { stroke: var(--sapphire-dark); }
  [data-status="ok"]     .arc { stroke: var(--success); }
  [data-status="warn"]   .arc { stroke: var(--pumpkin); }
  [data-status="behind"] .arc { stroke: var(--danger); }

  .prayer-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .prayer-label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }
  .prayer-status {
    font-family: var(--font-heading);
    font-size: 17px;
    font-weight: 700;
    line-height: 1.2;
  }
  [data-status="done"]   .prayer-status { color: var(--sapphire-dark); }
  [data-status="ok"]     .prayer-status { color: var(--success); }
  [data-status="warn"]   .prayer-status { color: var(--pumpkin); }
  [data-status="behind"] .prayer-status { color: var(--danger); }
  .prayer-next {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 1px;
  }

  .prayer-btn {
    width: 100%;
    height: 54px;
    border-radius: var(--radius-sm);
    font-family: var(--font-heading);
    font-size: 17px;
    font-weight: 700;
    color: #fff;
    background: var(--success);
    letter-spacing: 0.2px;
    transition: filter 0.15s, transform 0.12s;
    box-shadow: 0 3px 10px rgba(22,163,74,0.3);
  }
  .prayer-btn:hover  { filter: brightness(1.08); }
  .prayer-btn:active { transform: scale(0.97); }
  [data-status="warn"]   .prayer-btn {
    background: var(--pumpkin);
    box-shadow: 0 3px 10px rgba(238,111,0,0.3);
  }
  [data-status="behind"] .prayer-btn {
    background: var(--danger);
    box-shadow: 0 3px 10px rgba(220,38,38,0.3);
  }

  .undo-btn {
    display: block;
    margin: 8px auto 0;
    background: none;
    color: var(--text-light);
    font-size: 12px;
    padding: 2px 6px;
    text-decoration: underline;
    text-underline-offset: 2px;
    transition: color 0.15s;
  }
  .undo-btn:hover { color: var(--text-muted); }

  .prayer-complete {
    width: 100%;
    text-align: center;
    padding: 14px;
    background: var(--sapphire-light);
    border-radius: var(--radius-sm);
    font-family: var(--font-heading);
    font-size: 15px;
    font-weight: 700;
    color: var(--sapphire-dark);
  }
</style>
