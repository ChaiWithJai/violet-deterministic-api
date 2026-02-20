<script lang="ts">
  import type { StudioEvent } from '../../lib/api/types';

  interface Props {
    events: StudioEvent[];
    currentPhase: string;
  }

  let { events, currentPhase }: Props = $props();

  const phases = ['plan', 'scaffold', 'generate', 'verify', 'bundle'];

  function phaseStatus(phase: string): 'done' | 'active' | 'pending' {
    const idx = phases.indexOf(phase);
    const currentIdx = phases.indexOf(currentPhase);
    if (currentIdx < 0) return 'pending';
    if (idx < currentIdx) return 'done';
    if (idx === currentIdx) return 'active';
    return 'pending';
  }

  function phaseEvents(phase: string): StudioEvent[] {
    return events.filter((e) => e.phase === phase);
  }
</script>

<div class="timeline">
  {#each phases as phase}
    {@const status = phaseStatus(phase)}
    <div class="timeline-phase {status}">
      <div class="phase-dot">
        {#if status === 'done'}✓{:else if status === 'active'}●{:else}○{/if}
      </div>
      <div class="phase-content">
        <span class="phase-name">{phase}</span>
        {#each phaseEvents(phase) as event}
          <p class="phase-event">{event.message || event.type}</p>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .timeline {
    display: flex;
    flex-direction: column;
    gap: 0;
    position: relative;
    padding-left: var(--space-md);
  }

  .timeline::before {
    content: '';
    position: absolute;
    left: 11px;
    top: 12px;
    bottom: 12px;
    width: 2px;
    background: var(--border-subtle);
  }

  .timeline-phase {
    display: flex;
    align-items: flex-start;
    gap: var(--space-md);
    padding: var(--space-sm) 0;
    position: relative;
  }

  .phase-dot {
    flex-shrink: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    font-size: 0.6875rem;
    font-weight: 700;
    background: var(--bg-elevated);
    border: 2px solid var(--border-subtle);
    color: var(--text-tertiary);
    z-index: 1;
  }

  .done .phase-dot {
    background: var(--pass);
    border-color: var(--pass);
    color: #fff;
  }

  .active .phase-dot {
    background: var(--accent);
    border-color: var(--accent);
    color: #fff;
    box-shadow: 0 0 8px var(--accent-glow);
  }

  .phase-content {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding-top: 2px;
  }

  .phase-name {
    font-size: 0.8125rem;
    font-weight: 600;
    text-transform: capitalize;
    color: var(--text-primary);
  }

  .pending .phase-name {
    color: var(--text-tertiary);
  }

  .phase-event {
    font-size: 0.75rem;
    color: var(--text-secondary);
    font-family: var(--font-code);
  }
</style>
