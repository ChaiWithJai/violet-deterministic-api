import type { App } from '../api/types';

function createAppStore() {
  let apps = $state<App[]>([]);
  let active = $state<App | null>(null);
  let loading = $state(false);

  return {
    get apps() { return apps; },
    get active() { return active; },
    get loading() { return loading; },

    setApps(list: App[]) {
      apps = list;
    },

    setActive(a: App | null) {
      active = a;
    },

    setLoading(v: boolean) {
      loading = v;
    },

    updateActive(partial: Partial<App>) {
      if (active) {
        active = { ...active, ...partial };
      }
    },
  };
}

export const appStore = createAppStore();
