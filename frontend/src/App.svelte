<script>
  import { api } from './lib/api.js'
  import CycleList from './components/CycleList.svelte'
  import CycleDetail from './components/CycleDetail.svelte'
  import CreateCycle from './components/CreateCycle.svelte'

  let view = $state('list')
  let selectedId = $state(null)
  let cycles = $state([])
  let loadError = $state('')

  async function loadCycles() {
    try {
      cycles = await api.getCycles()
      loadError = ''
    } catch (e) {
      loadError = e.message
    }
  }

  $effect(() => { loadCycles() })

  function showCreate() { view = 'create' }
  function showDetail(id) { selectedId = id; view = 'detail' }
  function showList() { view = 'list'; loadCycles() }

  function onCreated(cycle) {
    view = 'detail'
    selectedId = cycle.id
    loadCycles()
  }
</script>

<div class="app">
  <header>
    <button class="logo" onclick={showList}>
      <svg width="28" height="28" viewBox="0 0 512 512" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
        <rect width="512" height="512" rx="110" fill="#154374"/>
        <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
        <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
        <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
        <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
      </svg>
      <span class="logo-text">Bajet</span>
    </button>
    {#if view !== 'list'}
      <button class="header-home" onclick={showList} title="Dashboard">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 9.5L12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V9.5z"/>
          <path d="M9 21V12h6v9"/>
        </svg>
      </button>
    {/if}
  </header>

  <main>
    {#if view === 'list'}
      {#if loadError}
        <div class="api-error">
          <p>Tidak dapat terhubung ke server.</p>
          <small>{loadError}</small>
          <button onclick={loadCycles}>Coba lagi</button>
        </div>
      {:else}
        <CycleList {cycles} onSelect={showDetail} onNew={showCreate} />
      {/if}
    {:else if view === 'create'}
      <CreateCycle onCreated={onCreated} onCancel={showList} />
    {:else if view === 'detail'}
      <CycleDetail cycleId={selectedId} onBack={showList} />
    {/if}
  </main>
</div>

<style>
  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  header {
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--sapphire-dark);
    padding: 0 16px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    box-shadow: 0 2px 8px rgba(21,67,116,0.25);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
    background: none;
    padding: 0;
  }

  .logo-text {
    font-family: var(--font-heading);
    font-size: 20px;
    font-weight: 800;
    color: var(--banana);
    letter-spacing: -0.5px;
  }

  .header-home {
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255,255,255,0.1);
    color: rgba(255,255,255,0.8);
    padding: 6px;
    border-radius: var(--radius-xs);
    width: 36px;
    height: 36px;
    transition: background 0.15s;
  }
  .header-home:hover { background: rgba(255,255,255,0.2); color: white; }

  main { flex: 1; }

  .api-error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-muted);
    max-width: 400px;
    margin: 0 auto;
  }
  .api-error p { font-family: var(--font-heading); font-size: 16px; font-weight: 700; color: var(--text); margin-bottom: 8px; }
  .api-error small { display: block; font-size: 12px; color: var(--danger); margin-bottom: 20px; }
  .api-error button {
    background: var(--primary);
    color: white;
    font-size: 14px;
    font-weight: 700;
    padding: 10px 20px;
    border-radius: var(--radius-sm);
  }
</style>
