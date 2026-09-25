<script>
  import { onMount, untrack } from 'svelte'
  import { api } from './lib/api.js'
  import { isSignedIn, signIn, signOut, getUser } from './lib/auth.js'
  import { i18n, lang } from './lib/i18n.js'
  import { syncing, sessionExpired } from './lib/sync.js'
  import CycleList from './components/CycleList.svelte'
  import CycleDetail from './components/CycleDetail.svelte'
  import CreateCycle from './components/CreateCycle.svelte'
  import Overview from './components/Overview.svelte'
  import { nextCycleDraft, latestCycle } from './lib/utils.js'

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
        callback: async ({ credential }) => {
          try {
            const res = await fetch(`${import.meta.env.VITE_API_URL}/api/auth/google`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ credential }),
            })
            const data = await res.json()
            if (!res.ok) throw new Error(data.error || 'Sign-in failed')
            signIn(data.token, data)
            signedIn = true
            user = getUser()
          } catch (e) {
            console.error('Sign-in failed:', e)
          }
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

  // untrack: loadCycles reads `cycles`, which would otherwise make every
  // successful load re-trigger this effect in an endless fetch loop.
  $effect(() => { if (signedIn) untrack(loadCycles) })

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

  function handleSignOut() {
    signOut()
    signedIn = false
    user = null
    cycles = []
    view = 'list'
  }

  let showTabBar = $derived(signedIn && (view === 'list' || view === 'overview'))

  // ── Re-auth overlay (session expired) ─────────────────────────────────────
  let reAuthEl = $state(null)

  $effect(() => {
    if (!USES_BACKEND || !$sessionExpired || !reAuthEl) return
    function renderBtn() {
      window.google.accounts.id.initialize({
        client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
        callback: async ({ credential }) => {
          try {
            const res = await fetch(`${import.meta.env.VITE_API_URL}/api/auth/google`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ credential }),
            })
            const data = await res.json()
            if (!res.ok) throw new Error(data.error || 'Sign-in failed')
            signIn(data.token, data)
            sessionExpired.set(false)
            signedIn = true
            user = getUser()
            loadCycles()
          } catch (e) {
            console.error('Re-auth failed:', e)
          }
        },
      })
      window.google.accounts.id.renderButton(reAuthEl, {
        theme: 'outline', size: 'large', shape: 'pill',
      })
    }
    if (window.google) renderBtn()
    else window.onGoogleLibraryLoad = renderBtn
  })
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
      {#if USES_BACKEND && signedIn}
        <button class="chip user" onclick={handleSignOut} title={$i18n.signOut}>
          {#if user?.picture}
            <img src={user.picture} alt="" class="avatar" referrerpolicy="no-referrer" />
          {:else}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/>
            </svg>
          {/if}
          <span>{$i18n.signOut}</span>
        </button>
      {/if}
    </div>
  </header>

  <main>
    {#if !signedIn}
      <div class="signin-screen">
        <svg width="72" height="72" viewBox="0 0 512 512" fill="none" aria-hidden="true">
        <rect width="512" height="512" rx="110" fill="#154374"/>
        <rect x="88"  y="210" width="96" height="215" rx="16" fill="#F2E942" opacity="0.55"/>
        <rect x="208" y="118" width="96" height="307" rx="16" fill="#F2E942"/>
        <rect x="328" y="158" width="96" height="267" rx="16" fill="#F2E942" opacity="0.80"/>
        <rect x="68"  y="430" width="376" height="10"  rx="5"  fill="#F2E942" opacity="0.35"/>
      </svg>
        <h1>Bajet</h1>
        <p>Period budgeting, simplified.</p>
        <div bind:this={gSigninEl}></div>
      </div>
    {:else if view === 'list'}
      {#if loadError}
        <div class="api-error">
          <p>{$i18n.cannotConnect}</p>
          <small>{loadError}</small>
          <button onclick={loadCycles}>{$i18n.tryAgain}</button>
        </div>
      {:else}
        <CycleList {cycles} {loading} onSelect={showDetail} onNew={showCreate} onStartNext={startNext} />
      {/if}
    {:else if view === 'overview'}
      <Overview {cycles} />
    {:else if view === 'create'}
      <CreateCycle initial={formInitial} editing={formEditing}
        onCreated={openCycle} onSaved={openCycle} onCancel={cancelForm} />
    {:else if view === 'detail'}
      <CycleDetail cycleId={selectedId} initialCycle={selectedCycle} isLatest={selectedId === latestId}
        onBack={showList} onEdit={editCycle} onStartNext={startNext} />
    {/if}
  </main>

  <!-- Sync indicator -->
  {#if $syncing}
    <div class="sync-bar"></div>
  {/if}

  <!-- Session-expired re-auth overlay -->
  {#if USES_BACKEND && $sessionExpired}
    <div class="reauth-overlay">
      <div class="reauth-card">
        <div class="reauth-icon">
          <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <h2 class="reauth-title">{$i18n.sessionExpiredTitle}</h2>
        <p class="reauth-sub">{$i18n.sessionExpiredSub}</p>
        <div bind:this={reAuthEl} class="reauth-btn-wrap"></div>
      </div>
    </div>
  {/if}

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
  .chip.user { padding-left: 4px; }
  .avatar { width: 26px; height: 26px; border-radius: 50%; object-fit: cover; }

  main { flex: 1; }

  /* ── Sign-in screen ── */
  .signin-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    min-height: calc(100vh - 60px);
    padding: 40px 20px;
    text-align: center;
  }
  .signin-screen h1 {
    font-family: var(--font-heading);
    font-size: 36px;
    font-weight: 800;
    color: var(--sapphire-dark);
    letter-spacing: -1.2px;
    margin-top: 8px;
  }
  .signin-screen p { color: var(--text-muted); font-size: 16px; margin-bottom: 12px; }

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

  /* ── Sync bar ── */
  .sync-bar {
    position: fixed;
    left: 0; right: 0;
    bottom: 0;
    height: 3px;
    z-index: 30;
    overflow: hidden;
  }
  .sync-bar::after {
    content: '';
    position: absolute;
    top: 0; bottom: 0;
    left: -40%;
    width: 40%;
    border-radius: 2px;
    background: var(--sapphire);
    animation: sync-sweep 1.2s ease-in-out infinite;
  }
  @keyframes sync-sweep {
    0%   { left: -40%; }
    100% { left: 110%; }
  }

  /* ── Re-auth overlay ── */
  .reauth-overlay {
    position: fixed;
    inset: 0;
    z-index: 200;
    background: rgba(14,47,83,0.55);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }
  .reauth-card {
    background: var(--surface);
    border-radius: var(--radius);
    padding: 28px 24px;
    max-width: 340px;
    width: 100%;
    text-align: center;
    box-shadow: var(--shadow-lg);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }
  .reauth-icon {
    width: 60px; height: 60px;
    border-radius: 50%;
    background: var(--pumpkin-light);
    color: var(--pumpkin);
    display: grid;
    place-items: center;
  }
  .reauth-title { font-family: var(--font-heading); font-size: 20px; font-weight: 800; }
  .reauth-sub { font-size: 14px; color: var(--text-muted); line-height: 1.5; }
  .reauth-btn-wrap { margin-top: 8px; }

  /* ── Error ── */
  .api-error {
    text-align: center;
    padding: 64px 24px;
    color: var(--text-muted);
    max-width: 400px;
    margin: 0 auto;
  }
  .api-error p { font-family: var(--font-heading); font-size: 17px; font-weight: 700; color: var(--text); margin-bottom: 6px; }
  .api-error small { display: block; font-size: 12px; color: var(--danger); margin-bottom: 20px; }
  .api-error button {
    background: var(--sapphire-dark);
    color: var(--banana);
    font-family: var(--font-heading);
    font-size: 14px;
    font-weight: 700;
    padding: 12px 22px;
    border-radius: var(--radius-sm);
  }
</style>
