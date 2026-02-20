<script lang="ts">
  import Nav from './components/shell/Nav.svelte';
  import Layout from './components/shell/Layout.svelte';
  import AuthGate from './components/shell/AuthGate.svelte';
  import Home from './routes/Home.svelte';
  import Studio from './routes/Studio.svelte';
  import Decisions from './routes/Decisions.svelte';
  import Blueprint from './routes/Blueprint.svelte';
  import Governance from './routes/Governance.svelte';
  import Settings from './routes/Settings.svelte';

  // Hash-based routing
  let hash = $state(window.location.hash.slice(1) || '/');

  function navigate(path: string) {
    window.location.hash = '#' + path;
  }

  // Listen for hash changes
  $effect(() => {
    function onHashChange() {
      hash = window.location.hash.slice(1) || '/';
    }
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  });

  // Parse route params
  interface RouteMatch {
    route: string;
    params: Record<string, string>;
  }

  function matchRoute(h: string): RouteMatch {
    if (h === '/' || h === '') return { route: 'home', params: {} };
    if (h === '/decisions') return { route: 'decisions', params: {} };
    if (h === '/governance') return { route: 'governance', params: {} };
    if (h === '/settings') return { route: 'settings', params: {} };

    const studioMatch = h.match(/^\/studio\/(.+)$/);
    if (studioMatch) return { route: 'studio', params: { jobId: studioMatch[1] } };

    const appMatch = h.match(/^\/apps\/(.+)$/);
    if (appMatch) return { route: 'blueprint', params: { appId: appMatch[1] } };

    return { route: 'home', params: {} };
  }

  let currentRoute = $derived(matchRoute(hash));
</script>

<AuthGate>
  <Nav currentRoute={hash} onnavigate={navigate} />
  <Layout>
    {#if currentRoute.route === 'home'}
      <Home onnavigate={navigate} />
    {:else if currentRoute.route === 'studio'}
      <Studio jobId={currentRoute.params.jobId} />
    {:else if currentRoute.route === 'decisions'}
      <Decisions />
    {:else if currentRoute.route === 'blueprint'}
      <Blueprint appId={currentRoute.params.appId} />
    {:else if currentRoute.route === 'governance'}
      <Governance />
    {:else if currentRoute.route === 'settings'}
      <Settings />
    {/if}
  </Layout>
</AuthGate>
