<script lang="ts">
  import PromptInput from '../components/prompt/PromptInput.svelte';
  import ClarifyChat from '../components/prompt/ClarifyChat.svelte';
  import ConfirmationForm from '../components/prompt/ConfirmationForm.svelte';
  import { agentPlan, agentClarify, createStudioJob } from '../lib/api/endpoints';
  import type { AgentPlanResponse, AgentClarifyResponse, ClarifyQuestion, Confirmation, StudioJobRequest } from '../lib/api/types';

  interface Props {
    onnavigate: (route: string) => void;
  }

  let { onnavigate }: Props = $props();

  type Phase = 'prompt' | 'clarify' | 'confirm';

  let phase = $state<Phase>('prompt');
  let loading = $state(false);
  let planResult = $state<AgentPlanResponse | null>(null);
  let clarifyResult = $state<AgentClarifyResponse | null>(null);
  let questions = $state<ClarifyQuestion[]>([]);
  let confirmation = $state<Partial<Confirmation>>({});
  let prompt = $state('');
  let error = $state<string | null>(null);

  async function handlePrompt(text: string) {
    prompt = text;
    loading = true;
    error = null;

    const res = await agentPlan({ prompt: text });
    loading = false;

    if (!res.ok) {
      error = (res.data as any).error || 'Failed to create plan';
      return;
    }

    planResult = res.data;

    // Plan returns a suggested_blueprint — seed the confirmation and go to clarify
    confirmation = {
      prompt: text,
      app_name: res.data.name,
      plan: res.data.suggested_blueprint.plan,
      region: res.data.suggested_blueprint.region,
    };

    // Start clarify to get questions
    const clarifyRes = await agentClarify({
      prompt: text,
      confirmation,
    });

    if (!clarifyRes.ok) {
      error = (clarifyRes.data as any).error || 'Clarify failed';
      return;
    }

    clarifyResult = clarifyRes.data;
    confirmation = clarifyRes.data.updated_confirmation;

    if (clarifyRes.data.ready_to_generate) {
      phase = 'confirm';
    } else {
      questions = clarifyRes.data.questions;
      phase = 'clarify';
    }
  }

  async function handleClarify(answers: Record<string, string>) {
    loading = true;
    error = null;

    const res = await agentClarify({
      prompt,
      confirmation,
      answers,
    });
    loading = false;

    if (!res.ok) {
      error = (res.data as any).error || 'Clarify failed';
      return;
    }

    clarifyResult = res.data;
    confirmation = res.data.updated_confirmation;

    if (res.data.ready_to_generate) {
      phase = 'confirm';
    } else {
      questions = res.data.questions;
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

    onnavigate(`/studio/${res.data.job_id}`);
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
    <ClarifyChat {questions} onanswer={handleClarify} {loading} summary={clarifyResult?.summary} />
  {:else if phase === 'confirm'}
    <ConfirmationForm {confirmation} checks={planResult?.checks ?? []} ongenerate={handleGenerate} {loading} />
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
