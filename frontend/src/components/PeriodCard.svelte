<script>
  import { fmtShort, fmtIDR } from '../lib/utils.js'
  import { api } from '../lib/api.js'

  let { period, onUpdate } = $props()

  let showForm = $state(false)
  let resultType = $state('sisa')
  let amount = $state('')
  let loading = $state(false)
  let error = $state('')

  async function submit() {
    const val = parseFloat(amount)
    if (isNaN(val) || val < 0) { error = 'Masukkan jumlah yang valid'; return }
    loading = true; error = ''
    try {
      await api.checkIn(period.id, { result_type: resultType, result_amount: val })
      showForm = false; amount = ''
      onUpdate()
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }

  async function undo() {
    if (!confirm('Batalkan check-in ini?')) return
    loading = true
    try {
      await api.undoCheckIn(period.id)
      onUpdate()
    } catch (e) {
      alert(e.message)
    } finally {
      loading = false
    }
  }
</script>

<div class="card" class:completed={period.status === 'completed'} class:open={period.status === 'open'}>
  <div class="header">
    <div class="label">
      <span class="badge">P{period.period_number}</span>
      <span class="dates">{fmtShort(period.start_date)} – {fmtShort(period.end_date)}</span>
    </div>
    <div class="budget">Rp {fmtIDR(period.budget)}</div>
  </div>

  {#if period.status === 'completed'}
    <div class="result" class:sisa={period.result_type === 'sisa'} class:defisit={period.result_type === 'defisit'}>
      <span class="result-type">{period.result_type === 'sisa' ? '✓ Sisa' : '✗ Defisit'}</span>
      <span class="result-amount">Rp {fmtIDR(period.result_amount)}</span>
      <button class="undo-btn" onclick={undo} disabled={loading}>Batal</button>
    </div>
  {:else}
    {#if showForm}
      <div class="form">
        <div class="toggle">
          <button
            class="toggle-btn"
            class:active={resultType === 'sisa'}
            onclick={() => resultType = 'sisa'}
          >Sisa</button>
          <button
            class="toggle-btn defisit"
            class:active={resultType === 'defisit'}
            onclick={() => resultType = 'defisit'}
          >Defisit</button>
        </div>
        <div class="input-row">
          <span class="prefix">Rp</span>
          <input
            type="number"
            placeholder="0"
            min="0"
            bind:value={amount}
            onkeydown={(e) => e.key === 'Enter' && submit()}
          />
        </div>
        {#if error}<p class="err">{error}</p>{/if}
        <div class="actions">
          <button class="btn-cancel" onclick={() => { showForm = false; error = '' }}>Batal</button>
          <button class="btn-submit" onclick={submit} disabled={loading}>
            {loading ? '...' : 'Simpan'}
          </button>
        </div>
      </div>
    {:else}
      <button class="checkin-btn" onclick={() => showForm = true}>
        + Check-in
      </button>
    {/if}
  {/if}
</div>

<style>
  .card {
    background: var(--surface);
    border-radius: var(--radius-sm);
    padding: 14px 16px;
    border-left: 4px solid var(--border);
    transition: border-color 0.2s;
  }
  .card.open { border-left-color: var(--primary); }
  .card.completed { border-left-color: var(--border); }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .label {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .badge {
    background: var(--primary-light);
    color: var(--primary);
    font-size: 12px;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 20px;
  }
  .dates {
    font-size: 13px;
    color: var(--text-muted);
  }
  .budget {
    font-size: 13px;
    font-weight: 600;
    color: var(--text);
  }

  .result {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
    padding: 8px 12px;
    border-radius: var(--radius-sm);
    font-size: 14px;
  }
  .result.sisa { background: var(--success-light); color: var(--success); }
  .result.defisit { background: var(--danger-light); color: var(--danger); }

  .result-type { font-weight: 600; }
  .result-amount { font-weight: 700; flex: 1; }
  .undo-btn {
    font-size: 11px;
    padding: 3px 8px;
    border-radius: 6px;
    background: rgba(0,0,0,0.08);
    color: inherit;
    opacity: 0.7;
  }
  .undo-btn:hover { opacity: 1; }

  .form { margin-top: 12px; }
  .toggle {
    display: flex;
    gap: 4px;
    margin-bottom: 10px;
  }
  .toggle-btn {
    flex: 1;
    padding: 8px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    font-weight: 600;
    background: var(--surface-2);
    color: var(--text-muted);
    border: 2px solid transparent;
    transition: all 0.15s;
  }
  .toggle-btn.active {
    background: var(--success-light);
    color: var(--success);
    border-color: var(--success);
  }
  .toggle-btn.defisit.active {
    background: var(--danger-light);
    color: var(--danger);
    border-color: var(--danger);
  }

  .input-row {
    display: flex;
    align-items: center;
    border: 1.5px solid var(--border);
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: var(--surface);
  }
  .prefix {
    padding: 0 12px;
    color: var(--text-muted);
    font-size: 13px;
    font-weight: 600;
    border-right: 1.5px solid var(--border);
    background: var(--surface-2);
    height: 40px;
    display: flex;
    align-items: center;
  }
  input {
    flex: 1;
    border: none;
    padding: 0 12px;
    height: 40px;
    font-size: 15px;
    font-weight: 600;
    color: var(--text);
    background: transparent;
    outline: none;
  }
  .err {
    font-size: 12px;
    color: var(--danger);
    margin-top: 6px;
  }
  .actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
  }
  .btn-cancel {
    flex: 1;
    padding: 10px;
    border-radius: var(--radius-sm);
    font-size: 14px;
    font-weight: 600;
    background: var(--surface-2);
    color: var(--text-muted);
  }
  .btn-submit {
    flex: 2;
    padding: 10px;
    border-radius: var(--radius-sm);
    font-size: 14px;
    font-weight: 600;
    background: var(--primary);
    color: white;
  }
  .btn-submit:disabled { opacity: 0.6; }

  .checkin-btn {
    width: 100%;
    margin-top: 10px;
    padding: 9px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    font-weight: 600;
    background: var(--primary-light);
    color: var(--primary);
    transition: background 0.15s;
  }
  .checkin-btn:hover { background: #dde4fd; }
</style>
