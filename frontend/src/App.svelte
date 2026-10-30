<script>
  import { untrack } from 'svelte'
  import { api, importLegacyCache } from './lib/api.js'
  import { i18n, lang } from './lib/i18n.js'
  import CycleList from './components/CycleList.svelte'
  import CycleDetail from './components/CycleDetail.svelte'
  import CreateCycle from './components/CreateCycle.svelte'
  import Overview from './components/Overview.svelte'
  import { nextCycleDraft, latestCycle } from './lib/utils.js'

  // Devices that used the old server-backed app keep their cycles: the copy it
  // left in this browser becomes the local data on first start.
  importLegacyCache()

  let view = $state('list')
  let selectedId = $state(null)
  let selectedCycle = $state(null)
  let cycles = $state([])
  let loading = $state(true)

  async function loadCycles() {
    try {
      cycles = await api.getCycles()
    } finally {
      loading = false
    }
  }

  // untrack: loadCycles writes `cycles`; tracking it would re-run this forever.
  $effect(() => { untrack(loadCycles) })

  // Create form state: `formInitial` prefills a new cycle, `formEditing` edits an existing one.
  let formInitial = $state(null)
  let formEditing = $state(null)

  function showCreate()   { formInitial = null; formEditing = null; view = 'create' }
  function startNext(cycle) { formInitial = nextCycleDraft(cycle); formEditing = null; view = 'create' }
  function editCycle(cycle) { formInitial = null; formEditing = cycle; view = 'create' }
  function showDetail(id) {
    selectedId = id
    selectedCycle = cycles.find(c => c.id === id) ?? null
    view = 'detail'
  }
  function showList()     { view = 'list'; loadCycles() }
  function showOverview()     { view = 'overview' }

  function openCycle(cycle) {
    selectedId = cycle.id
    selectedCycle = cycle
    view = 'detail'
    loadCycles()
  }
  function cancelForm() {
    if (formEditing) openCycle(formEditing)
    else showList()
  }

  let latestId = $derived(latestCycle(cycles)?.id ?? null)

  let showTabBar = $derived(view === 'list' || view === 'overview')
</script>

<div class="app">
  <header>
    <button class="logo" onclick={showList}>
      <svg width="30" height="30" viewBox="0 0 512 512" fill="none" aria-hidden="true">
        <rect width="512" height="512" rx="110" fill="#154374"/>
        <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
        <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
        <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
        <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
      </svg>
      <span class="logo-text">Bajet</span>
    </button>

    <div class="header-right">
      <button class="chip" onclick={() => lang.update(l => l === 'id' ? 'en' : 'id')} title="Switch language">
        {$lang === 'id' ? 'EN' : 'ID'}
      </button>
    </div>
  </header>

  <main>
    {#if view === 'list'}
      <CycleList {cycles} {loading} onSelect={showDetail} onNew={showCreate} onStartNext={startNext} />
    {:else if view === 'overview'}
      <Overview {cycles} onDataChanged={loadCycles} />
    {:else if view === 'create'}
      <CreateCycle initial={formInitial} editing={formEditing}
        onCreated={openCycle} onSaved={openCycle} onCancel={cancelForm} />
    {:else if view === 'detail'}
      <CycleDetail cycleId={selectedId} initialCycle={selectedCycle} isLatest={selectedId === latestId}
        onBack={showList} onEdit={editCycle} onStartNext={startNext} />
    {/if}
  </main>

  {#if showTabBar}
    <nav class="tab-bar">
      <div class="tabs">
        <button class="tab" class:active={view === 'list'} onclick={showList} aria-current={view === 'list' ? 'page' : undefined}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="3" width="7" height="7" rx="1.5"/>
            <rect x="14" y="3" width="7" height="7" rx="1.5"/>
            <rect x="3" y="14" width="7" height="7" rx="1.5"/>
            <rect x="14" y="14" width="7" height="7" rx="1.5"/>
          </svg>
          <span>{$i18n.cyclesTab}</span>
        </button>
        <button class="tab" class:active={view === 'overview'} onclick={showOverview} aria-current={view === 'overview' ? 'page' : undefined}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="20" x2="18" y2="10"/>
            <line x1="12" y1="20" x2="12" y2="4"/>
            <line x1="6"  y1="20" x2="6"  y2="14"/>
          </svg>
          <span>{$i18n.overviewTab}</span>
        </button>
      </div>
    </nav>
  {/if}
</div>

<style>
  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  /* ── Header ── */
  header {
    position: sticky;
    top: 0;
    z-index: 10;
    height: 60px;
    padding: 0 16px;
    padding-top: env(safe-area-inset-top, 0px);
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: rgba(245,244,240,0.82);
    backdrop-filter: saturate(160%) blur(14px);
    -webkit-backdrop-filter: saturate(160%) blur(14px);
    border-bottom: 1px solid rgba(214,211,202,0.6);
  }
  .logo {
    display: flex;
    align-items: center;
    gap: 9px;
    background: none;
    padding: 0;
  }
  .logo-text {
    font-family: var(--font-heading);
    font-size: 21px;
    font-weight: 800;
    color: var(--sapphire-dark);
    letter-spacing: -0.6px;
  }
  .header-right { display: flex; align-items: center; gap: 6px; }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 34px;
    padding: 0 12px;
    border-radius: 999px;
    background: var(--surface);
    border: 1px solid var(--border);
    font-family: var(--font-heading);
    font-size: 12px;
    font-weight: 700;
    color: var(--text-muted);
    transition: border-color 0.15s, color 0.15s;
  }
  .chip:hover { border-color: var(--border-strong); color: var(--text); }
  main { flex: 1; }

  /* ── Tab bar ── */
  .tab-bar {
    position: fixed;
    left: 0; right: 0;
    bottom: 0;
    z-index: 20;
    padding: 0 16px calc(10px + var(--safe-bottom));
    pointer-events: none;
  }
  .tabs {
    pointer-events: auto;
    max-width: 320px;
    height: var(--tabbar-h);
    margin: 0 auto;
    padding: 6px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
    border-radius: 999px;
    background: rgba(17,26,36,0.92);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: var(--shadow-lg);
  }
  .tab {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    border-radius: 999px;
    background: transparent;
    color: rgba(255,255,255,0.6);
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
    transition: background 0.2s, color 0.2s;
  }
  .tab.active { background: var(--banana); color: var(--sapphire-deep); }
  .tab:not(.active):hover { color: #fff; }

</style>
