<script lang="ts">
  import PromptInput from '../components/prompt/PromptInput.svelte';
  import ClarifyChat from '../components/prompt/ClarifyChat.svelte';
  import ConfirmationForm from '../components/prompt/ConfirmationForm.svelte';
  import { agentPlan, agentClarify, createStudioJob } from '../lib/api/endpoints';
  import type { AgentPlanResponse, StudioJobRequest } from '../lib/api/types';

  interface Props {
    onnavigate: (route: string) => void;
  }

  let { onnavigate }: Props = $props();

  type Phase = 'prompt' | 'clarify' | 'confirm';

  let phase = $state<Phase>('prompt');
  let loading = $state(false);
  let plan = $state<AgentPlanResponse | null>(null);
  let clarifyQuestions = $state<string[]>([]);
  let error = $state<string | null>(null);

  async function handlePrompt(prompt: string) {
    loading = true;
    error = null;

    const res = await agentPlan({ prompt });
    loading = false;

    if (!res.ok) {
      error = (res.data as any).error || 'Failed to create plan';
      return;
    }

    plan = res.data;

    if (res.data.clarify_questions && res.data.clarify_questions.length > 0) {
      clarifyQuestions = res.data.clarify_questions;
      phase = 'clarify';
    } else {
      phase = 'confirm';
    }
  }

  async function handleClarify(answers: Record<string, string>) {
    if (!plan) return;
    loading = true;
    error = null;

    const res = await agentClarify({ plan_id: plan.plan_id, answers });
    loading = false;

    if (!res.ok) {
      error = (res.data as any).error || 'Clarify failed';
      return;
    }

    if (res.data.ready_to_generate || !res.data.clarify_questions?.length) {
      plan = { ...plan!, steps: res.data.updated_steps || plan!.steps, ready_to_generate: true };
      phase = 'confirm';
    } else {
      clarifyQuestions = res.data.clarify_questions!;
    }
  }

  async function handleGenerate(req: StudioJobRequest) {
    loading = true;
    error = null;

    const res = await createStudioJob(req);
    loading = false;

    if (!res.ok) {
      error = (res.data as any).error || 'Failed to create job';
      return;
    }

    onnavigate(`/studio/${res.data.id}`);
  }
</script>

<div class="home-page">
  <div class="hero">
    <h1 class="hero-title">Build your app with<br /><span class="hero-accent">deterministic AI</span></h1>
    <p class="hero-subtitle">Describe what you need. VDA plans, generates, verifies, and deploys — with every decision auditable and replayable.</p>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if phase === 'prompt'}
    <PromptInput onsubmit={handlePrompt} {loading} />
  {:else if phase === 'clarify'}
    <ClarifyChat questions={clarifyQuestions} onanswer={handleClarify} {loading} />
  {:else if phase === 'confirm' && plan}
    <ConfirmationForm {plan} ongenerate={handleGenerate} {loading} />
  {/if}
</div>

<style>
  .home-page {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
  }

  .hero {
    text-align: center;
    padding-top: var(--space-2xl);
  }

  .hero-title {
    font-family: var(--font-display);
    font-size: clamp(2rem, 5vw, 3rem);
    font-weight: 700;
    letter-spacing: -0.04em;
    line-height: 1.1;
    color: var(--text-primary);
    margin-bottom: var(--space-md);
  }

  .hero-accent {
    background: linear-gradient(135deg, var(--accent), #a78bfa);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .hero-subtitle {
    font-size: 1rem;
    color: var(--text-secondary);
    max-width: 560px;
    margin: 0 auto;
    line-height: 1.6;
  }

  .error-banner {
    max-width: 720px;
    margin: 0 auto;
    padding: var(--space-md);
    background: var(--fail-subtle);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-md);
    color: var(--fail);
    font-size: 0.875rem;
    text-align: center;
  }
</style>
