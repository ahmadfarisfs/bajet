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
    <button class="logo" onclick={showList}>💰 Bajet</button>
    {#if view !== 'list'}
      <button class="header-home" onclick={showList} title="Dashboard">⌂</button>
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
    background: var(--surface);
    border-bottom: 1px solid var(--border);
    padding: 0 16px;
    height: 54px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    box-shadow: 0 1px 4px rgba(0,0,0,0.05);
  }

  .logo {
    background: none;
    font-size: 17px;
    font-weight: 800;
    color: var(--primary);
    letter-spacing: -0.3px;
    padding: 0;
  }

  .header-home {
    background: none;
    font-size: 18px;
    color: var(--text-muted);
    padding: 4px 8px;
    border-radius: 6px;
  }
  .header-home:hover { background: var(--surface-2); color: var(--text); }

  main { flex: 1; }

  .api-error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-muted);
    max-width: 400px;
    margin: 0 auto;
  }
  .api-error p { font-size: 16px; font-weight: 600; color: var(--text); margin-bottom: 8px; }
  .api-error small { display: block; font-size: 12px; color: var(--danger); margin-bottom: 20px; }
  .api-error button {
    background: var(--primary);
    color: white;
    font-size: 14px;
    font-weight: 600;
    padding: 10px 20px;
    border-radius: var(--radius-sm);
  }
</style>
