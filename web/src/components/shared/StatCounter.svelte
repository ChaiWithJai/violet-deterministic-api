<script lang="ts">
  interface Props {
    label: string;
    value: string | number;
    sublabel?: string;
  }

  let { label, value, sublabel = '' }: Props = $props();
</script>

<div class="stat">
  <span class="stat-value">{value}</span>
  <span class="stat-label">{label}</span>
  {#if sublabel}
    <span class="stat-sublabel">{sublabel}</span>
  {/if}
</div>

<style>
  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    padding: var(--space-md);
  }

  .stat-value {
    font-family: var(--font-display);
    font-size: 1.75rem;
    font-weight: 600;
    letter-spacing: -0.03em;
    color: var(--text-primary);
  }

  .stat-label {
    font-family: var(--font-body);
    font-size: 0.75rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .stat-sublabel {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
  }
</style>
