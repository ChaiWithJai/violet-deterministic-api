<script lang="ts">
  interface Props {
    currentRoute: string;
    onnavigate: (route: string) => void;
  }

  let { currentRoute, onnavigate }: Props = $props();

  const routes = [
    { path: '/', label: 'Home', icon: '⬡' },
    { path: '/studio', label: 'Studio', icon: '◈' },
    { path: '/decisions', label: 'Decisions', icon: '◇' },
    { path: '/apps', label: 'Blueprint', icon: '△' },
    { path: '/governance', label: 'Governance', icon: '◎' },
    { path: '/settings', label: 'Settings', icon: '⚙' },
  ];

  function isActive(routePath: string): boolean {
    if (routePath === '/') return currentRoute === '/';
    return currentRoute.startsWith(routePath);
  }
</script>

<nav class="nav-pill">
  {#each routes as route}
    <button
      class="nav-item"
      class:active={isActive(route.path)}
      onclick={() => onnavigate(route.path)}
    >
      <span class="nav-icon">{route.icon}</span>
      <span class="nav-label">{route.label}</span>
    </button>
  {/each}
</nav>

<style>
  .nav-pill {
    position: fixed;
    top: var(--space-md);
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 4px;
    background: var(--bg-glass);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: var(--glass-border);
    border-radius: var(--radius-pill);
    z-index: 100;
    box-shadow: var(--shadow-lg);
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    border-radius: var(--radius-pill);
    font-family: var(--font-body);
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--text-secondary);
    transition: all var(--duration-fast) var(--ease-out);
    white-space: nowrap;
  }

  .nav-item:hover {
    color: var(--text-primary);
    background: var(--bg-surface);
  }

  .nav-item.active {
    color: var(--text-on-accent);
    background: var(--accent);
    box-shadow: var(--shadow-glow);
  }

  .nav-icon {
    font-size: 0.875rem;
    line-height: 1;
  }

  .nav-label {
    line-height: 1;
  }

  @media (max-width: 768px) {
    .nav-label {
      display: none;
    }
    .nav-item {
      padding: 8px 12px;
    }
  }
</style>
