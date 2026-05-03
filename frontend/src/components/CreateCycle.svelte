<script>
  import { api } from '../lib/api.js'
  import { todayStr, addDays } from '../lib/utils.js'

  let { onCreated, onCancel } = $props()

  let startDate = $state(todayStr())
  let endDate = $state(addDays(todayStr(), 29))
  let totalBudget = $state(2000000)
  let divisionMode = $state('equal')
  let loading = $state(false)
  let error = $state('')

  async function submit() {
    if (!startDate || !endDate) { error = 'Tanggal harus diisi'; return }
    if (totalBudget <= 0) { error = 'Budget harus lebih dari 0'; return }
    loading = true; error = ''
    try {
      const cycle = await api.createCycle({
        start_date: startDate,
        end_date: endDate,
        total_budget: Number(totalBudget),
        division_mode: divisionMode,
      })
      onCreated(cycle)
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }
</script>

<div class="view">
  <div class="topbar">
    <button class="back" onclick={onCancel}>← Batal</button>
  </div>

  <div class="card">
    <h2>Buat Cycle Baru</h2>

    <label>
      Tanggal Mulai
      <input type="date" bind:value={startDate} />
    </label>

    <label>
      Tanggal Selesai
      <input type="date" bind:value={endDate} />
    </label>

    <label>
      Total Budget (IDR)
      <div class="input-row">
        <span class="prefix">Rp</span>
        <input type="number" min="1" bind:value={totalBudget} />
      </div>
    </label>

    <div class="field">
      <span class="field-label">Mode Pembagian</span>
      <div class="toggle">
        <button
          class="toggle-btn"
          class:active={divisionMode === 'equal'}
          onclick={() => divisionMode = 'equal'}
        >
          <strong>Equal</strong>
          <small>Merata (8/8/7/7)</small>
        </button>
        <button
          class="toggle-btn"
          class:active={divisionMode === 'behavioral'}
          onclick={() => divisionMode = 'behavioral'}
        >
          <strong>Behavioral</strong>
          <small>Awal lebih panjang (9/8/7/7)</small>
        </button>
      </div>
    </div>

    {#if error}<p class="err">{error}</p>{/if}

    <button class="btn-primary" onclick={submit} disabled={loading}>
      {loading ? 'Membuat...' : 'Buat Cycle'}
    </button>
  </div>

  <p class="hint">
    Cycle akan dibagi menjadi <strong>4 periode</strong> otomatis berdasarkan rentang tanggal.
  </p>
</div>

<style>
  .view {
    max-width: 480px;
    margin: 0 auto;
    padding: 16px 16px 48px;
  }
  .topbar { margin-bottom: 20px; }
  .back {
    background: none;
    font-family: var(--font-heading);
    font-size: 14px; font-weight: 700;
    color: var(--primary);
    padding: 6px 0;
  }

  .card {
    background: var(--surface);
    border-radius: var(--radius);
    padding: 24px 20px;
    box-shadow: var(--shadow-sm);
    display: flex;
    flex-direction: column;
    gap: 18px;
    border-top: 4px solid var(--sapphire-dark);
  }
  h2 {
    font-family: var(--font-heading);
    font-size: 20px; font-weight: 800;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 13px; font-weight: 600;
    color: var(--text-muted);
  }
  label input[type="date"],
  label input[type="number"] {
    border: 1.5px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0 12px;
    height: 44px;
    font-size: 15px;
    color: var(--text);
    background: var(--surface);
    outline: none;
    width: 100%;
    transition: border-color 0.15s;
  }
  label input:focus { border-color: var(--primary); }

  .input-row {
    display: flex;
    align-items: center;
    border: 1.5px solid var(--border);
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: var(--surface);
    transition: border-color 0.15s;
  }
  .input-row:focus-within { border-color: var(--primary); }
  .prefix {
    padding: 0 12px;
    color: var(--text-muted);
    font-size: 13px; font-weight: 600;
    border-right: 1.5px solid var(--border);
    background: var(--surface-2);
    height: 44px;
    display: flex;
    align-items: center;
  }
  .input-row input {
    flex: 1;
    border: none !important;
    padding: 0 12px;
    height: 44px;
    font-family: var(--font-heading);
    font-size: 15px; font-weight: 700;
    color: var(--text);
    background: transparent;
    outline: none;
  }

  .field { display: flex; flex-direction: column; gap: 6px; }
  .field-label { font-size: 13px; font-weight: 600; color: var(--text-muted); }

  .toggle { display: flex; gap: 8px; }
  .toggle-btn {
    flex: 1;
    padding: 10px 8px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    background: var(--surface-2);
    color: var(--text-muted);
    border: 2px solid transparent;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    transition: all 0.15s;
  }
  .toggle-btn strong {
    font-family: var(--font-heading);
    font-size: 13px;
    color: var(--text);
  }
  .toggle-btn small { font-size: 11px; }
  .toggle-btn.active {
    background: var(--banana-light);
    border-color: var(--banana-dark);
  }
  .toggle-btn.active strong { color: var(--sapphire-dark); }
  .toggle-btn.active small  { color: var(--sapphire-dark); }

  .err { font-size: 13px; color: var(--danger); margin-top: -8px; }

  .btn-primary {
    width: 100%;
    padding: 14px;
    border-radius: var(--radius-sm);
    font-family: var(--font-heading);
    font-size: 15px; font-weight: 800;
    background: var(--sapphire-dark);
    color: var(--banana);
    letter-spacing: 0.2px;
    transition: background 0.15s;
  }
  .btn-primary:hover:not(:disabled) { background: var(--primary); color: white; }
  .btn-primary:disabled { opacity: 0.6; }

  .hint {
    text-align: center;
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 16px;
    line-height: 1.5;
  }
</style>
