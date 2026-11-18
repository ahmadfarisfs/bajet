<script>
  import { registerSW } from 'virtual:pwa-register'
  import { i18n } from '../lib/i18n.js'

  // `raised` lifts the banner above the bottom tab bar when it is showing.
  let { raised = false } = $props()

  let needRefresh = $state(false)
  let dismissed = $state(false)
  let updating = $state(false)

  // A new version waits (registerType: 'prompt') until the user taps Update.
  const updateSW = registerSW({
    onNeedRefresh() { needRefresh = true; dismissed = false },
    onRegisteredSW(_url, reg) {
      if (!reg) return
      // Browsers only look for a new version when the app is opened. An installed
      // app can stay open in the background for days, so also check when it comes
      // back to the foreground and every hour while it is open.
      const check = () => {
        if (!navigator.onLine || reg.installing) return
        reg.update().catch(() => {})
      }
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') check()
      })
      setInterval(check, 60 * 60 * 1000)
    },
  })

  function update() {
    updating = true
    updateSW(true) // activates the waiting version, then reloads the page
    // Fallback in case the page wasn't controlled by the old version and no reload comes.
    setTimeout(() => location.reload(), 3000)
  }
</script>

{#if needRefresh && !dismissed}
  <div class="update" class:raised role="status">
    <span class="msg">{$i18n.updateAvailable}</span>
    <button class="go" onclick={update} disabled={updating}>
      {updating ? $i18n.updating : $i18n.updateBtn}
    </button>
    <button class="close" onclick={() => dismissed = true} aria-label={$i18n.updateLater}>
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
    </button>
  </div>
{/if}

<style>
  .update {
    position: fixed;
    left: 50%;
    transform: translateX(-50%);
    bottom: calc(16px + var(--safe-bottom));
    z-index: 30;
    width: calc(100% - 32px);
    max-width: 420px;
    display: flex; align-items: center; gap: 10px;
    padding: 10px 10px 10px 16px;
    border-radius: 16px;
    background: var(--sapphire-dark);
    color: #fff;
    box-shadow: var(--shadow-lg);
    animation: up 0.25s ease-out;
  }
  .update.raised { bottom: calc(var(--tabbar-h) + 22px + var(--safe-bottom)); }
  .msg { flex: 1; font-size: 14px; font-weight: 600; }
  .go {
    padding: 8px 14px;
    border-radius: 999px;
    background: var(--banana, #F2E942);
    color: var(--sapphire-dark);
    font-family: var(--font-heading); font-size: 14px; font-weight: 800;
  }
  .go:disabled { opacity: 0.7; }
  .close {
    width: 32px; height: 32px;
    display: grid; place-items: center;
    border-radius: 999px;
    background: none; color: rgba(255,255,255,0.75);
  }
  @keyframes up { from { opacity: 0; transform: translate(-50%, 12px); } }
</style>
