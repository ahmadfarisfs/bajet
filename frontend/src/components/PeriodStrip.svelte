<script>
  import { isActive, isOverdue, daysBetween } from '../lib/utils.js'

  // One segment per period, sized by its length in days and colored by outcome.
  let { periods = [], tone = 'light' } = $props()

  function state(p) {
    if (p.status === 'completed') return p.result_amount === 0 ? 'even' : p.result_type === 'sisa' ? 'sisa' : 'defisit'
    if (isOverdue(p)) return 'overdue'
    if (isActive(p.start_date, p.end_date)) return 'current'
    return 'future'
  }
</script>

<div class="strip" class:dark={tone === 'dark'} aria-hidden="true">
  {#each periods as p (p.id ?? p.period_number)}
    <span class="seg {state(p)}" style="flex-grow: {daysBetween(p.start_date, p.end_date)}"></span>
  {/each}
</div>

<style>
  .strip {
    display: flex;
    gap: 3px;
    height: 6px;
  }
  .seg {
    flex-basis: 0;
    border-radius: 3px;
    background: var(--border);
    transition: background 0.3s;
  }
  .seg.sisa    { background: var(--success); }
  .seg.defisit { background: var(--danger); }
  .seg.even    { background: var(--sapphire); }
  .seg.current { background: var(--pumpkin); }
  .seg.overdue {
    background: repeating-linear-gradient(135deg, var(--pumpkin) 0 4px, var(--pumpkin-light) 4px 7px);
  }

  .dark .seg        { background: rgba(255,255,255,0.18); }
  .dark .seg.sisa    { background: #6ee7a0; }
  .dark .seg.defisit { background: #ff8a80; }
  .dark .seg.even    { background: #9cc5f4; }
  .dark .seg.current { background: var(--banana); }
  .dark .seg.overdue {
    background: repeating-linear-gradient(135deg, var(--banana) 0 4px, rgba(255,255,255,0.18) 4px 7px);
  }
</style>
