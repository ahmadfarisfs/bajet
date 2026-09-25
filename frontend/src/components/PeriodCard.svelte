<script>
  import { fmtShort, fmtIDR, isActive, daysLeft, daysUntil, isOverdue } from '../lib/utils.js'
  import { api } from '../lib/api.js'
  import { i18n } from '../lib/i18n.js'

  // `adjusted` is the budget after spreading earlier surpluses/deficits (see adjustedBudgets).
  let { period, adjusted = null, onUpdate } = $props()

  let local    = $state({ ...period })
  let inflight = $state(false)

  $effect(() => { if (!inflight) local = { ...period } })

  let showForm   = $state(false)
  let resultType = $state('sisa')
  let amount     = $state('')
  let error      = $state('')

  let isCurrent = $derived(local.status === 'open' && isActive(local.start_date, local.end_date))
  let isFuture  = $derived(local.status === 'open' && daysUntil(local.start_date) > 0)
  let overdue   = $derived(isOverdue(local))
  let remaining = $derived(daysLeft(local.end_date))
  let startsIn  = $derived(daysUntil(local.start_date))
  let showAdjusted = $derived(local.status === 'open' && adjusted != null && adjusted !== local.budget)

  async function submit() {
    const val = parseFloat(amount)
    if (isNaN(val) || val < 0) { error = $i18n.invalidAmt; return }

    const prev = { ...local }
    inflight = true
    local = { ...local, status: 'completed', result_type: resultType, result_amount: val }
    showForm = false; amount = ''; error = ''

    try {
      await api.checkIn(period.id, { result_type: resultType, result_amount: val })
      onUpdate()
    } catch (e) {
      local = prev; showForm = true
      if (e.message !== 'session_expired') error = e.message
    } finally {
      inflight = false
    }
  }

  async function undo() {
    if (!confirm($i18n.undoConfirm)) return
    const prev = { ...local }
    inflight = true
    local = { ...local, status: 'open', result_type: '', result_amount: 0 }
    try {
      await api.undoCheckIn(period.id)
      onUpdate()
    } catch (e) {
      local = prev; alert(e.message)
    } finally {
      inflight = false
    }
  }
</script>

<div
  class="card"
  class:completed={local.status === 'completed'}
  class:current={isCurrent}
  class:overdue
  class:future={isFuture}
  class:saving={inflight}
>
  <div class="header">
    <span class="badge">P{local.period_number}</span>
    <div class="meta">
      <span class="dates">{fmtShort(local.start_date)} – {fmtShort(local.end_date)}</span>
      {#if isCurrent}
        <span class="tag" class:urgent={remaining <= 1}>{remaining <= 0 ? $i18n.lastDay : $i18n.daysLeftN(remaining)}</span>
      {:else if overdue}
        <span class="tag urgent">{$i18n.overdueBadge}</span>
      {:else if isFuture}
        <span class="tag muted">{$i18n.startsIn(startsIn)}</span>
      {/if}
    </div>
    <div class="budget">
      <span class="num">Rp {fmtIDR(showAdjusted ? adjusted : local.budget)}</span>
      {#if showAdjusted}
        <small class="planned num">Rp {fmtIDR(local.budget)}</small>
      {/if}
    </div>
  </div>

  {#if local.status === 'completed'}
    <div class="result" class:sisa={local.result_type === 'sisa'} class:defisit={local.result_type === 'defisit'}>
      <span class="result-type">{local.result_type === 'sisa' ? $i18n.surplusLabel : $i18n.deficitLabel}</span>
      <span class="result-amount num">Rp {fmtIDR(local.result_amount)}</span>
      <button class="undo-btn" onclick={undo} disabled={inflight}>{$i18n.undoBtn}</button>
    </div>
  {:else if isFuture}
    <!-- Nothing to do until it starts -->
  {:else if showForm}
    <div class="form">
      <div class="toggle" role="radiogroup">
        <button class="toggle-btn" role="radio" aria-checked={resultType === 'sisa'}
          class:active={resultType === 'sisa'} onclick={() => resultType = 'sisa'}>
          {$i18n.surplus}
        </button>
        <button class="toggle-btn defisit" role="radio" aria-checked={resultType === 'defisit'}
          class:active={resultType === 'defisit'} onclick={() => resultType = 'defisit'}>
          {$i18n.deficit}
        </button>
      </div>
      <div class="input-row">
        <span class="prefix">Rp</span>
        <input type="number" inputmode="numeric" placeholder="0" min="0" bind:value={amount}
          onkeydown={(e) => e.key === 'Enter' && submit()} />
      </div>
      {#if error}<p class="err">{error}</p>{/if}
      <div class="actions">
        <button class="btn-cancel" onclick={() => { showForm = false; error = '' }}>{$i18n.cancel}</button>
        <button class="btn-submit" onclick={submit} disabled={inflight}>{$i18n.save}</button>
      </div>
    </div>
  {:else}
    <button class="checkin-btn" class:strong={isCurrent || overdue} onclick={() => showForm = true}>
      {$i18n.checkIn}
    </button>
  {/if}
</div>

<style>
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 14px 16px;
    transition: border-color 0.2s, opacity 0.2s;
  }
  .card.current { border-color: var(--pumpkin); box-shadow: 0 0 0 3px var(--pumpkin-light); }
  .card.overdue { border-color: #f2b27a; background: #fffaf4; }
  .card.future  { background: var(--surface-2); }
  .card.future .dates, .card.future .budget { color: var(--text-light); }
  .card.saving  { opacity: 0.8; }

  .header { display: flex; align-items: center; gap: 10px; }
  .badge {
    flex-shrink: 0;
    width: 34px; height: 34px;
    display: grid; place-items: center;
    border-radius: 10px;
    font-family: var(--font-heading); font-size: 13px; font-weight: 800;
    background: var(--sapphire-light); color: var(--sapphire-dark);
  }
  .current .badge { background: var(--pumpkin); color: #fff; }
  .overdue .badge { background: var(--pumpkin-light); color: var(--pumpkin); }
  .completed .badge { background: var(--surface-2); color: var(--text-muted); }
  .future .badge { background: var(--border); color: var(--text-light); }

  .meta { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
  .dates { font-size: 14px; font-weight: 600; }
  .tag { font-size: 12px; font-weight: 600; color: var(--pumpkin); }
  .tag.urgent { color: var(--danger); }
  .tag.muted { color: var(--text-light); font-weight: 500; }

  .budget { display: flex; flex-direction: column; align-items: flex-end; font-family: var(--font-heading); font-size: 15px; font-weight: 700; }
  .planned { font-family: var(--font-body); font-size: 11px; font-weight: 500; color: var(--text-light); text-decoration: line-through; }

  .result {
    display: flex; align-items: center; gap: 8px;
    margin-top: 12px;
    padding: 9px 12px;
    border-radius: var(--radius-xs);
    font-size: 14px;
  }
  .result.sisa    { background: var(--success-light); color: var(--success); }
  .result.defisit { background: var(--danger-light);  color: var(--danger); }
  .result-type   { font-weight: 600; }
  .result-amount { font-family: var(--font-heading); font-weight: 700; flex: 1; }
  .undo-btn {
    font-size: 12px; font-weight: 600;
    padding: 4px 10px; border-radius: 999px;
    background: rgba(255,255,255,0.7);
  }

  .form { margin-top: 12px; display: flex; flex-direction: column; gap: 10px; }
  .toggle {
    display: grid; grid-template-columns: 1fr 1fr;
    padding: 3px; gap: 3px;
    background: var(--surface-2);
    border-radius: var(--radius-xs);
  }
  .toggle-btn {
    padding: 9px;
    border-radius: 8px;
    font-family: var(--font-heading); font-size: 14px; font-weight: 700;
    background: transparent; color: var(--text-muted);
    transition: background 0.15s, color 0.15s;
  }
  .toggle-btn.active { background: var(--surface); color: var(--success); box-shadow: var(--shadow-sm); }
  .toggle-btn.defisit.active { color: var(--danger); }

  .input-row {
    display: flex; align-items: center;
    border: 1.5px solid var(--border-strong);
    border-radius: var(--radius-xs);
    background: var(--surface);
    transition: border-color 0.15s;
  }
  .input-row:focus-within { border-color: var(--sapphire); }
  .prefix { padding: 0 4px 0 14px; color: var(--text-muted); font-weight: 600; }
  input {
    flex: 1; min-width: 0;
    border: none; outline: none; background: transparent;
    height: 48px; padding: 0 12px 0 4px;
    font-family: var(--font-heading); font-size: 20px; font-weight: 700;
    color: var(--text);
  }
  .err { font-size: 12px; color: var(--danger); }

  .actions { display: grid; grid-template-columns: 1fr 2fr; gap: 8px; }
  .btn-cancel {
    padding: 12px; border-radius: var(--radius-xs);
    font-size: 14px; font-weight: 600;
    background: var(--surface-2); color: var(--text-muted);
  }
  .btn-submit {
    padding: 12px; border-radius: var(--radius-xs);
    font-family: var(--font-heading); font-size: 15px; font-weight: 800;
    background: var(--sapphire-dark); color: var(--banana);
  }
  .btn-submit:disabled { opacity: 0.6; }

  .checkin-btn {
    width: 100%;
    margin-top: 12px;
    padding: 11px;
    border-radius: var(--radius-xs);
    font-family: var(--font-heading); font-size: 14px; font-weight: 700;
    background: var(--sapphire-light); color: var(--sapphire-dark);
    transition: filter 0.15s;
  }
  .checkin-btn.strong { background: var(--pumpkin); color: #fff; }
  .checkin-btn:hover { filter: brightness(0.97); }
</style>
