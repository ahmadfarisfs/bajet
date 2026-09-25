<script>
  import { untrack } from 'svelte'
  import { api } from '../lib/api.js'
  import { todayStr, addDays, daysBetween, buildPeriods, fmtIDR } from '../lib/utils.js'
  import { i18n } from '../lib/i18n.js'

  // `initial` prefills the form (e.g. "start next cycle"); `editing` switches it to edit mode.
  let { onCreated, onSaved, onCancel, initial = null, editing = null } = $props()

  // The form is seeded once when it opens; later prop changes shouldn't overwrite typing.
  const seed = untrack(() => editing ?? initial ?? {})
  const day = (iso) => iso?.substring(0, 10)

  let startDate    = $state(day(seed.start_date) ?? todayStr())
  let endDate      = $state(day(seed.end_date) ?? addDays(todayStr(), 29))
  let totalBudget  = $state(seed.total_budget ?? 2000000)
  let divisionMode = $state(seed.division_mode ?? 'equal')
  let numPeriods   = $state(seed.num_periods || (seed.periods ?? []).length || 4)
  let loading = $state(false)
  let error = $state('')

  // After a check-in, only budget and mode may change (the backend enforces the same).
  const locked = untrack(() => !!editing && (editing.periods ?? []).some(p => p.status === 'completed'))

  const MODES = [
    { key: 'equal',      title: 'modeEqual',     sub: 'modeEqualSub' },
    { key: 'behavioral', title: 'modeBehav',     sub: 'modeBehavSub' },
    { key: 'menurun',    title: 'modeMenurun',   sub: 'modeMenurunSub' },
    { key: 'progresif',  title: 'modeProgresif', sub: 'modeProgrSub' },
  ]

  let n = $derived(Math.min(Math.max(Math.round(Number(numPeriods)) || 1, 1), 12))
  let previewPeriods = $derived.by(() => {
    if (!startDate || !endDate || daysBetween(startDate, endDate) < n) return []
    return buildPeriods(startDate, endDate, Number(totalBudget) || 0, divisionMode, n)
  })

  function fmtDM(d) {
    const [, m, dd] = d.split('-')
    return `${dd}/${m}`
  }

  async function submit() {
    if (!startDate || !endDate) { error = $i18n.errDates; return }
    if (!(Number(totalBudget) > 0)) { error = $i18n.errBudget; return }
    loading = true; error = ''
    const body = {
      start_date: startDate,
      end_date: endDate,
      total_budget: Number(totalBudget),
      division_mode: divisionMode,
      num_periods: n,
    }
    try {
      if (editing) onSaved(await api.updateCycle(editing.id, body))
      else         onCreated(await api.createCycle(body))
    } catch (e) {
      if (e.message !== 'session_expired') error = e.message
    } finally {
      loading = false
    }
  }
</script>

<div class="page">
  <div class="topbar">
    <button class="back" onclick={onCancel}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
      {$i18n.backCancel}
    </button>
  </div>

  <h1>{editing ? $i18n.editCycle : $i18n.newCycle}</h1>

  <section class="group">
    <div class="row2">
      <label>
        <span>{$i18n.startDate}</span>
        <input type="date" bind:value={startDate} disabled={locked} />
      </label>
      <label>
        <span>{$i18n.endDate}</span>
        <input type="date" bind:value={endDate} disabled={locked} />
      </label>
    </div>
    <label>
      <span>{$i18n.numPeriods}</span>
      <div class="stepper">
        <button type="button" onclick={() => numPeriods = Math.max(1, n - 1)} disabled={locked || n <= 1} aria-label="−">−</button>
        <span class="num">{n}</span>
        <button type="button" onclick={() => numPeriods = Math.min(12, n + 1)} disabled={locked || n >= 12} aria-label="+">+</button>
      </div>
    </label>
    {#if locked}<p class="note">{$i18n.lockedHint}</p>{/if}
  </section>

  <section class="group">
    <label>
      <span>{$i18n.totalBudget}</span>
      <div class="money">
        <span class="prefix">Rp</span>
        <input type="number" inputmode="numeric" min="1" bind:value={totalBudget} />
      </div>
    </label>

    <div class="field">
      <span class="field-label">{$i18n.divMode}</span>
      <div class="modes">
        {#each MODES as m}
          <button type="button" class="mode" class:active={divisionMode === m.key} onclick={() => divisionMode = m.key}>
            <strong>{$i18n[m.title]}</strong>
            <small>{$i18n[m.sub]}</small>
          </button>
        {/each}
      </div>
    </div>
  </section>

  {#if previewPeriods.length > 0}
    <section class="group preview">
      <span class="field-label">{$i18n.previewLabel}</span>
      {#each previewPeriods as p}
        <div class="prow">
          <span class="p-num">P{p.period_number}</span>
          <span class="p-date">{fmtDM(p.start_date)} – {fmtDM(p.end_date)}</span>
          <span class="p-days num">{$i18n.daysShort(p.days)}</span>
          <span class="p-budget num">Rp {fmtIDR(p.budget)}</span>
        </div>
      {/each}
    </section>
  {/if}

  {#if error}<p class="err">{error}</p>{/if}

  <button class="btn-primary" onclick={submit} disabled={loading}>
    {#if editing}
      {loading ? $i18n.saving : $i18n.saveChanges}
    {:else}
      {loading ? $i18n.creating : $i18n.createCycle}
    {/if}
  </button>
</div>

<style>
  .page {
    max-width: var(--page-max);
    margin: 0 auto;
    padding: 12px 16px calc(48px + var(--safe-bottom));
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .back {
    display: inline-flex; align-items: center; gap: 2px;
    background: none;
    font-family: var(--font-heading); font-size: 15px; font-weight: 700;
    color: var(--sapphire-dark);
    padding: 8px 8px 8px 0;
  }
  h1 { font-family: var(--font-heading); font-size: 26px; font-weight: 800; letter-spacing: -0.5px; }

  .group {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 16px;
    display: flex; flex-direction: column; gap: 14px;
  }
  .row2 { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }

  label, .field { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
  label > span, .field-label { font-size: 12px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.4px; }

  input[type="date"] {
    width: 100%; min-width: 0;
    height: 46px; padding: 0 10px;
    border: 1.5px solid var(--border-strong); border-radius: var(--radius-xs);
    font-size: 15px; color: var(--text); background: var(--surface);
    outline: none;
  }
  input:focus { border-color: var(--sapphire); }
  input:disabled { background: var(--surface-2); color: var(--text-light); }

  .stepper {
    display: grid; grid-template-columns: 46px 1fr 46px;
    align-items: center;
    height: 46px;
    border: 1.5px solid var(--border-strong); border-radius: var(--radius-xs);
    overflow: hidden;
  }
  .stepper span { text-align: center; font-family: var(--font-heading); font-size: 18px; font-weight: 800; }
  .stepper button { height: 100%; font-size: 20px; font-weight: 600; background: var(--surface-2); color: var(--sapphire-dark); }
  .stepper button:disabled { color: var(--text-light); }

  .note { font-size: 12px; color: var(--pumpkin); background: var(--pumpkin-light); padding: 8px 10px; border-radius: var(--radius-xs); }

  .money {
    display: flex; align-items: center;
    height: 56px;
    border: 1.5px solid var(--border-strong); border-radius: var(--radius-xs);
  }
  .money:focus-within { border-color: var(--sapphire); }
  .prefix { padding: 0 4px 0 14px; font-weight: 600; color: var(--text-muted); }
  .money input {
    flex: 1; min-width: 0;
    border: none; outline: none; background: transparent;
    padding: 0 12px 0 4px;
    font-family: var(--font-heading); font-size: 24px; font-weight: 800; color: var(--text);
  }

  .modes { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
  .mode {
    display: flex; flex-direction: column; align-items: flex-start; gap: 2px;
    text-align: left;
    padding: 11px 12px;
    border-radius: var(--radius-xs);
    background: var(--surface-2);
    border: 1.5px solid transparent;
    transition: border-color 0.15s, background 0.15s;
  }
  .mode strong { font-family: var(--font-heading); font-size: 14px; }
  .mode small { font-size: 12px; color: var(--text-muted); }
  .mode.active { background: var(--banana-light); border-color: var(--banana-dark); }
  .mode.active strong { color: var(--sapphire-dark); }

  .preview { gap: 0; }
  .preview .field-label { margin-bottom: 8px; }
  .prow {
    display: grid; grid-template-columns: 36px 1fr 36px auto;
    align-items: center; gap: 8px;
    padding: 8px 0;
    font-size: 13px;
  }
  .prow + .prow { border-top: 1px solid var(--border); }
  .p-num { font-family: var(--font-heading); font-weight: 800; color: var(--sapphire); }
  .p-days { color: var(--text-light); }
  .p-budget { font-weight: 700; text-align: right; }

  .err { font-size: 13px; color: var(--danger); }

  .btn-primary {
    width: 100%;
    padding: 16px;
    border-radius: var(--radius-sm);
    font-family: var(--font-heading); font-size: 16px; font-weight: 800;
    background: var(--sapphire-dark); color: var(--banana);
    box-shadow: var(--shadow);
  }
  .btn-primary:disabled { opacity: 0.6; }
</style>
