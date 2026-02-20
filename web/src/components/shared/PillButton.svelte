<script lang="ts">
  interface Props {
    variant?: 'primary' | 'secondary' | 'ghost';
    size?: 'sm' | 'md' | 'lg';
    disabled?: boolean;
    onclick?: () => void;
    children: import('svelte').Snippet;
  }

  let { variant = 'primary', size = 'md', disabled = false, onclick, children }: Props = $props();
</script>

<button
  class="pill-button {variant} {size}"
  {disabled}
  {onclick}
>
  {@render children()}
</button>

<style>
  .pill-button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-sm);
    border-radius: var(--radius-pill);
    font-family: var(--font-body);
    font-weight: 500;
    transition: all var(--duration-fast) var(--ease-out);
    white-space: nowrap;
    cursor: pointer;
  }

  .pill-button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  /* Sizes */
  .sm { padding: 6px 14px; font-size: 0.75rem; }
  .md { padding: 8px 20px; font-size: 0.8125rem; }
  .lg { padding: 12px 28px; font-size: 0.875rem; }

  /* Variants */
  .primary {
    background: var(--accent);
    color: var(--text-on-accent);
  }
  .primary:hover:not(:disabled) {
    background: var(--accent-hover);
    box-shadow: var(--shadow-glow);
  }

  .secondary {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    color: var(--text-primary);
  }
  .secondary:hover:not(:disabled) {
    background: var(--bg-surface-hover);
    border-color: var(--border-medium);
  }

  .ghost {
    background: transparent;
    color: var(--text-secondary);
  }
  .ghost:hover:not(:disabled) {
    background: var(--bg-surface);
    color: var(--text-primary);
  }
</style>
