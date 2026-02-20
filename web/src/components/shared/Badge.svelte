<script lang="ts">
  interface Props {
    variant?: 'default' | 'pass' | 'fail' | 'degraded' | 'info' | 'accent';
    children: import('svelte').Snippet;
  }

  let { variant = 'default', children }: Props = $props();
</script>

<span class="badge {variant}">
  {@render children()}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 10px;
    border-radius: var(--radius-pill);
    font-family: var(--font-code);
    font-size: 0.6875rem;
    font-weight: 500;
    letter-spacing: 0.02em;
    white-space: nowrap;
  }

  .default {
    background: var(--bg-surface);
    color: var(--text-secondary);
  }
  .pass {
    background: var(--pass-subtle);
    color: var(--pass);
  }
  .fail {
    background: var(--fail-subtle);
    color: var(--fail);
  }
  .degraded {
    background: var(--degraded-subtle);
    color: var(--degraded);
  }
  .info {
    background: var(--info-subtle);
    color: var(--info);
  }
  .accent {
    background: var(--accent-subtle);
    color: var(--accent);
  }
</style>
