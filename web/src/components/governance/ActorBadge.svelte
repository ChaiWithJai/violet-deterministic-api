<script lang="ts">
  interface Props {
    actor: 'human' | 'agent';
  }

  let { actor }: Props = $props();
</script>

<span class="actor-badge {actor}">
  <span class="actor-icon">{actor === 'human' ? '👤' : '🤖'}</span>
  <span class="actor-label">{actor}</span>
</span>

<style>
  .actor-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 10px;
    border-radius: var(--radius-pill);
    font-size: 0.6875rem;
    font-weight: 500;
  }

  .human {
    background: var(--info-subtle);
    color: var(--info);
  }

  .agent {
    background: var(--accent-subtle);
    color: var(--accent);
  }

  .actor-icon {
    font-size: 0.625rem;
  }

  .actor-label {
    text-transform: capitalize;
  }
</style>
