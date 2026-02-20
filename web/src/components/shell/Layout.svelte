<script lang="ts">
  interface Props {
    children: import('svelte').Snippet;
  }

  let { children }: Props = $props();
</script>

<main class="layout">
  {@render children()}
</main>

<style>
  .layout {
    min-height: 100dvh;
    padding-top: 72px;
    padding-left: var(--space-lg);
    padding-right: var(--space-lg);
    padding-bottom: var(--space-2xl);
    max-width: 1400px;
    margin: 0 auto;
  }

  @media (max-width: 768px) {
    .layout {
      padding-left: var(--space-md);
      padding-right: var(--space-md);
    }
  }
</style>
