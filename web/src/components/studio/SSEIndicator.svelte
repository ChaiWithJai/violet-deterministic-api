<script lang="ts">
  interface Props {
    connected: boolean;
    status?: string;
  }

  let { connected, status = '' }: Props = $props();
</script>

<div class="sse-indicator" class:connected>
  <span class="sse-dot"></span>
  <span class="sse-label">{connected ? (status || 'Connected') : 'Disconnected'}</span>
</div>

<style>
  .sse-indicator {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: var(--radius-pill);
    background: var(--fail-subtle);
    font-size: 0.6875rem;
    font-family: var(--font-code);
    color: var(--fail);
  }

  .sse-indicator.connected {
    background: var(--pass-subtle);
    color: var(--pass);
  }

  .sse-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }

  .connected .sse-dot {
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }

  @media (prefers-reduced-motion: reduce) {
    .connected .sse-dot {
      animation: none;
    }
  }
</style>
