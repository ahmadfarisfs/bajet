<script>
  import { onMount } from 'svelte'
  import { api } from './lib/api.js'
  import { isSignedIn, signIn, signOut, getUser } from './lib/auth.js'
  import CycleList from './components/CycleList.svelte'
  import CycleDetail from './components/CycleDetail.svelte'
  import CreateCycle from './components/CreateCycle.svelte'
  import Overview from './components/Overview.svelte'

  const USES_BACKEND = !!import.meta.env.VITE_API_URL

  const CACHE_KEY = 'bajet_cycles_cache'
  function readCache() {
    try { return JSON.parse(localStorage.getItem(CACHE_KEY) || 'null') } catch { return null }
  }
  function writeCache(data) {
    localStorage.setItem(CACHE_KEY, JSON.stringify(data))
  }

  const cached = readCache()

  let view = $state('list')
  let selectedId = $state(null)
  let selectedCycle = $state(null)
  let cycles = $state(cached || [])
  let loadError = $state('')
  let signedIn = $state(!USES_BACKEND || isSignedIn())
  let user = $state(getUser())
  let gSigninEl = $state(null)
  let loading = $state(!cached)

  $effect(() => {
    if (!USES_BACKEND || signedIn || !gSigninEl) return
    function renderBtn() {
      window.google.accounts.id.initialize({
        client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
        callback: ({ credential }) => {
          signIn(credential)
          signedIn = true
          user = getUser()
        },
      })
      window.google.accounts.id.renderButton(gSigninEl, {
        theme: 'outline', size: 'large', shape: 'pill',
      })
    }
    if (window.google) renderBtn()
    else window.onGoogleLibraryLoad = renderBtn
  })

  async function loadCycles() {
    if (!cycles.length) loading = true
    try {
      const fresh = await api.getCycles()
      cycles = fresh
      writeCache(fresh)
      loadError = ''
    } catch (e) {
      if (!cycles.length) loadError = e.message
    } finally {
      loading = false
    }
  }

  $effect(() => { if (signedIn) loadCycles() })

  function showCreate()   { view = 'create' }
  function showDetail(id) {
    selectedId = id
    selectedCycle = cycles.find(c => c.id === id) ?? null
    view = 'detail'
  }
  function showList()     { view = 'list'; loadCycles() }
  function showOverview()     { view = 'overview' }

  function onCreated(cycle) {
    view = 'detail'
    selectedId = cycle.id
    loadCycles()
  }

  function handleSignOut() {
    signOut()
    signedIn = false
    user = null
    cycles = []
    view = 'list'
  }

  let showTabBar = $derived(signedIn && (view === 'list' || view === 'overview'))
</script>

<div class="app">
  <header>
    <button class="logo" onclick={showList}>
      <svg width="28" height="28" viewBox="0 0 512 512" fill="none" aria-hidden="true">
        <rect width="512" height="512" rx="110" fill="#154374"/>
        <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
        <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
        <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
        <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
      </svg>
      <span class="logo-text">Bajet</span>
    </button>

    <div class="header-right">
      {#if view === 'detail' || view === 'create'}
        <button class="header-icon-btn" onclick={showList} title="Dashboard">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 9.5L12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V9.5z"/>
            <path d="M9 21V12h6v9"/>
          </svg>
        </button>
      {/if}
      {#if USES_BACKEND && signedIn}
        <button class="header-user" onclick={handleSignOut} title="Sign out">
          {#if user?.picture}
            <img src={user.picture} alt={user.name} class="avatar" referrerpolicy="no-referrer" />
          {:else}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/>
            </svg>
          {/if}
          <span>Keluar</span>
        </button>
      {/if}
    </div>
  </header>

  <main>
    {#if !signedIn}
      <div class="signin-screen">
        <div class="signin-logo">
          <svg width="64" height="64" viewBox="0 0 512 512" fill="none" aria-hidden="true">
            <rect width="512" height="512" rx="110" fill="#154374"/>
            <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
            <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
            <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
            <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
          </svg>
        </div>
        <h1>Bajet</h1>
        <p>Period budgeting, simplified.</p>
        <div bind:this={gSigninEl}></div>
      </div>
    {:else if view === 'list'}
      {#if loadError}
        <div class="api-error">
          <p>Tidak dapat terhubung ke server.</p>
          <small>{loadError}</small>
          <button onclick={loadCycles}>Coba lagi</button>
        </div>
      {:else}
        <CycleList {cycles} {loading} onSelect={showDetail} onNew={showCreate} />
      {/if}
    {:else if view === 'overview'}
      <Overview {cycles} />
    {:else if view === 'create'}
      <CreateCycle onCreated={onCreated} onCancel={showList} />
    {:else if view === 'detail'}
      <CycleDetail cycleId={selectedId} initialCycle={selectedCycle} onBack={showList} />
    {/if}
  </main>

  {#if showTabBar}
    <nav class="tab-bar">
      <button class="tab" class:active={view === 'list'} onclick={showList}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="3" width="7" height="7" rx="1"/>
          <rect x="14" y="3" width="7" height="7" rx="1"/>
          <rect x="3" y="14" width="7" height="7" rx="1"/>
          <rect x="14" y="14" width="7" height="7" rx="1"/>
        </svg>
        <span>Cycle</span>
      </button>
      <button class="tab" class:active={view === 'overview'} onclick={showOverview}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="20" x2="18" y2="10"/>
          <line x1="12" y1="20" x2="12" y2="4"/>
          <line x1="6"  y1="20" x2="6"  y2="14"/>
        </svg>
        <span>Overview</span>
      </button>
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
    background: var(--sapphire-dark);
    padding: 0 16px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    box-shadow: 0 2px 8px rgba(21,67,116,0.3);
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

  .header-right {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .header-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255,255,255,0.12);
    color: rgba(255,255,255,0.85);
    border-radius: var(--radius-xs);
    width: 36px; height: 36px;
    transition: background 0.15s;
  }
  .header-icon-btn:hover { background: rgba(255,255,255,0.22); }

  .header-user {
    display: flex;
    align-items: center;
    gap: 6px;
    background: rgba(255,255,255,0.12);
    color: rgba(255,255,255,0.85);
    font-size: 12px;
    font-weight: 600;
    padding: 5px 10px 5px 6px;
    border-radius: 20px;
    transition: background 0.15s;
  }
  .header-user:hover { background: rgba(255,255,255,0.22); }
  .avatar {
    width: 22px; height: 22px;
    border-radius: 50%;
    object-fit: cover;
  }

  main { flex: 1; }

  /* ── Sign-in screen ── */
  .signin-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 14px;
    min-height: calc(100vh - 56px);
    padding: 40px 20px;
    text-align: center;
  }
  .signin-logo {
    background: none;
    display: flex;
  }
  .signin-screen h1 {
    font-family: var(--font-heading);
    font-size: 32px;
    font-weight: 800;
    color: var(--sapphire-dark);
    letter-spacing: -1px;
    margin: 0;
  }
  .signin-screen p {
    color: var(--text-muted);
    font-size: 15px;
    margin: 0 0 8px;
  }

  /* ── Tab bar ── */
  .tab-bar {
    position: fixed;
    bottom: 0; left: 0; right: 0;
    z-index: 20;
    background: var(--surface);
    border-top: 1px solid var(--border);
    display: flex;
    height: 62px;
    box-shadow: 0 -2px 12px rgba(0,0,0,0.06);
  }
  .tab {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 3px;
    background: none;
    color: var(--text-light);
    font-size: 11px;
    font-weight: 600;
    padding: 8px;
    transition: color 0.15s;
    border-top: 2px solid transparent;
  }
  .tab.active {
    color: var(--sapphire-dark);
    border-top-color: var(--sapphire-dark);
  }
  .tab.active svg { stroke: var(--sapphire-dark); }
  .tab:not(.active):hover { color: var(--text-muted); }

  /* ── Error ── */
  .api-error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-muted);
    max-width: 400px;
    margin: 0 auto;
  }
  .api-error p {
    font-family: var(--font-heading);
    font-size: 16px;
    font-weight: 700;
    color: var(--text);
    margin-bottom: 8px;
  }
  .api-error small { display: block; font-size: 12px; color: var(--danger); margin-bottom: 20px; }
  .api-error button {
    background: var(--sapphire-dark);
    color: var(--banana);
    font-family: var(--font-heading);
    font-size: 14px;
    font-weight: 700;
    padding: 10px 20px;
    border-radius: var(--radius-sm);
  }
</style>
