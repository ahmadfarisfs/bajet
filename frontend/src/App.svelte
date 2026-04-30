<script>
  import { onMount } from 'svelte'
  import { api } from './lib/api.js'
  import { isSignedIn, signIn, signOut, getUser } from './lib/auth.js'
  import CycleList from './components/CycleList.svelte'
  import CycleDetail from './components/CycleDetail.svelte'
  import CreateCycle from './components/CreateCycle.svelte'
  import Overview from './components/Overview.svelte'
  import Landing from './components/Landing.svelte'

  const USES_BACKEND = !!import.meta.env.VITE_API_URL

  let view = $state('list')
  let selectedId = $state(null)
  let cycles = $state([])
  let loadError = $state('')
  let signedIn = $state(!USES_BACKEND || isSignedIn())
  let user = $state(getUser())
  let gSigninEl = $state(null)
  let loading = $state(false)
  // Show landing to first-time visitors (not yet signed in and haven't dismissed it)
  let showLanding = $state(!signedIn && !sessionStorage.getItem('bajet_seen_landing'))

  function onLandingGetStarted() {
    sessionStorage.setItem('bajet_seen_landing', '1')
    showLanding = false
  }

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
    loading = true
    try {
      cycles = await api.getCycles()
      loadError = ''
    } catch (e) {
      loadError = e.message
    } finally {
      loading = false
    }
  }

  $effect(() => { if (signedIn) loadCycles() })

  function showCreate() { view = 'create' }
  function showDetail(id) { selectedId = id; view = 'detail' }
  function showList() { view = 'list'; loadCycles() }
  function showOverview() { view = 'overview' }

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
    showLanding = true
    sessionStorage.removeItem('bajet_seen_landing')
  }

  let showTabBar = $derived(signedIn && (view === 'list' || view === 'overview'))
</script>

{#if showLanding}
  <Landing onGetStarted={onLandingGetStarted} />
{:else}
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
    {#if view === 'detail' || view === 'create'}
      <button class="header-home" onclick={showList} title="Dashboard">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 9.5L12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V9.5z"/>
          <path d="M9 21V12h6v9"/>
        </svg>
      </button>
    {/if}
    {#if USES_BACKEND && signedIn}
      <button class="header-signout" onclick={handleSignOut} title="Sign out">
        {#if user?.picture}
          <img src={user.picture} alt={user.name} class="avatar" referrerpolicy="no-referrer" />
        {/if}
        <span>Sign out</span>
      </button>
    {/if}
  </header>

  <main>
    {#if !signedIn}
      <div class="signin-screen">
        <svg width="48" height="48" viewBox="0 0 512 512" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
          <rect width="512" height="512" rx="100" fill="#4f46e5"/>
          <rect x="88"  y="210" width="96" height="215" rx="16" fill="white" opacity="0.65"/>
          <rect x="208" y="118" width="96" height="307" rx="16" fill="white"/>
          <rect x="328" y="158" width="96" height="267" rx="16" fill="white" opacity="0.85"/>
          <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="white" opacity="0.45"/>
        </svg>
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
      <CycleDetail cycleId={selectedId} onBack={showList} />
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
{/if}

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

  .header-signout {
    display: flex;
    align-items: center;
    gap: 6px;
    background: none;
    color: var(--text-muted);
    font-size: 12px;
    padding: 6px 8px;
    border-radius: 8px;
  }
  .header-signout:hover { background: var(--surface-2); color: var(--text); }
  .avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    object-fit: cover;
  }

  .signin-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    min-height: calc(100vh - 54px);
    padding: 40px 20px;
    text-align: center;
  }
  .signin-screen h1 { font-size: 28px; font-weight: 800; color: var(--primary); margin: 0; }
  .signin-screen p  { color: var(--text-muted); margin: 0 0 8px; }

  main { flex: 1; }

  /* ── Bottom tab bar ── */
  .tab-bar {
    position: fixed;
    bottom: 0;
    left: 0; right: 0;
    z-index: 20;
    background: var(--surface);
    border-top: 1px solid var(--border);
    display: flex;
    height: 60px;
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
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 600;
    padding: 8px;
    transition: color 0.15s;
  }
  .tab.active { color: var(--primary); }
  .tab.active svg { stroke: var(--primary); }
  .tab:not(.active):hover { color: var(--text); }

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
