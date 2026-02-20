import type { DecisionResponse, ReplayResponse } from '../api/types';

function createDecisionStore() {
  let decisions = $state<DecisionResponse[]>([]);
  let activeReplay = $state<ReplayResponse | null>(null);
  let loading = $state(false);

  return {
    get decisions() { return decisions; },
    get activeReplay() { return activeReplay; },
    get loading() { return loading; },

    addDecision(d: DecisionResponse) {
      decisions = [d, ...decisions];
    },

    setReplay(r: ReplayResponse | null) {
      activeReplay = r;
    },

    setLoading(v: boolean) {
      loading = v;
    },

    clear() {
      decisions = [];
      activeReplay = null;
      loading = false;
    },
  };
}

export const decisionStore = createDecisionStore();
