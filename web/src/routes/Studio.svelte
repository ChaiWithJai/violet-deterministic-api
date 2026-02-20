<script lang="ts">
  import { onMount } from 'svelte';
  import FileExplorer from '../components/studio/FileExplorer.svelte';
  import PreviewPanes from '../components/studio/PreviewPanes.svelte';
  import TerminalPanel from '../components/studio/TerminalPanel.svelte';
  import JobTimeline from '../components/studio/JobTimeline.svelte';
  import SSEIndicator from '../components/studio/SSEIndicator.svelte';
  import ConsoleLog from '../components/studio/ConsoleLog.svelte';
  import QualityGate from '../components/verification/QualityGate.svelte';
  import VerificationReport from '../components/verification/VerificationReport.svelte';
  import JTBDCoverage from '../components/verification/JTBDCoverage.svelte';
  import Badge from '../components/shared/Badge.svelte';
  import CodeBlock from '../components/shared/CodeBlock.svelte';
  import { jobStore } from '../lib/stores/job.svelte';
  import { authStore } from '../lib/stores/auth.svelte';
  import {
    getStudioArtifacts,
    getStudioVerification,
    getStudioJTBD,
    getStudioPreviewUrl,
    getStudioBundleUrl,
    runStudioTarget,
  } from '../lib/api/endpoints';
  import type { VerificationResult, JTBDCoverage as JTBDType } from '../lib/api/types';

  interface Props {
    jobId: string;
  }

  let { jobId }: Props = $props();

  let activeTab = $state<'preview' | 'code' | 'workload' | 'verification'>('preview');
  let artifacts = $state<Record<string, string>>({});
  let selectedFile = $state('');
  let verification = $state<VerificationResult[]>([]);
  let jtbd = $state<JTBDType[]>([]);
  let sidebarCollapsed = $state(false);

  let previewUrl = $derived(getStudioPreviewUrl(jobId, authStore.baseUrl));
  let bundleUrl = $derived(getStudioBundleUrl(jobId, authStore.baseUrl));

  onMount(() => {
    jobStore.connect(jobId);
    loadArtifacts();
    return () => jobStore.disconnect();
  });

  async function loadArtifacts() {
    const res = await getStudioArtifacts(jobId);
    if (res.ok) {
      artifacts = res.data;
      const keys = Object.keys(artifacts);
      if (keys.length > 0 && !selectedFile) selectedFile = keys[0];
    }
  }

  async function loadVerification() {
    const [vRes, jRes] = await Promise.all([
      getStudioVerification(jobId),
      getStudioJTBD(jobId),
    ]);
    if (vRes.ok) verification = vRes.data;
    if (jRes.ok) jtbd = jRes.data;
  }

  async function handleRunTarget(target: string) {
    await runStudioTarget(jobId, target);
    loadVerification();
  }
</script>

<div class="studio-page">
  <!-- Header -->
  <div class="studio-header">
    <div class="header-left">
      <h2 class="studio-title">
        {jobStore.job?.prompt?.slice(0, 60) || 'Studio'}
        {#if jobStore.job?.prompt && jobStore.job.prompt.length > 60}...{/if}
      </h2>
      <SSEIndicator connected={jobStore.connected} status={jobStore.job?.phase} />
    </div>
    <div class="header-right">
      {#if jobStore.job}
        <Badge variant="accent">{jobStore.job.status}</Badge>
        <Badge variant="default">{jobStore.job.phase}</Badge>
        <span class="progress-text">{Math.round(jobStore.job.progress || 0)}%</span>
      {/if}
      <a href={bundleUrl} class="bundle-link" download>Download Bundle</a>
    </div>
  </div>

  <!-- Three-column layout -->
  <div class="studio-grid" class:sidebar-collapsed={sidebarCollapsed}>
    <!-- Left: File Explorer -->
    <aside class="studio-sidebar">
      <FileExplorer {artifacts} {selectedFile} onselect={(f) => selectedFile = f} />
    </aside>

    <!-- Center: Main content -->
    <section class="studio-main">
      <div class="tab-bar">
        {#each ['preview', 'code', 'workload', 'verification'] as tab}
          <button
            class="tab-btn"
            class:active={activeTab === tab}
            onclick={() => { activeTab = tab as typeof activeTab; if (tab === 'verification') loadVerification(); }}
          >
            {tab}
          </button>
        {/each}
      </div>

      <div class="tab-content">
        {#if activeTab === 'preview'}
          <PreviewPanes {previewUrl} />
        {:else if activeTab === 'code'}
          {#if selectedFile && artifacts[selectedFile]}
            <CodeBlock code={artifacts[selectedFile]} lang={selectedFile.split('.').pop() || ''} maxHeight="600px" />
          {:else}
            <div class="empty-state">Select a file to view</div>
          {/if}
        {:else if activeTab === 'workload'}
          <JobTimeline events={jobStore.events} currentPhase={jobStore.job?.phase || ''} />
        {:else if activeTab === 'verification'}
          <div class="verification-tab">
            <QualityGate onrun={handleRunTarget} results={verification} />
            {#if verification.length > 0}
              <VerificationReport results={verification} />
            {/if}
            {#if jtbd.length > 0}
              <JTBDCoverage items={jtbd} />
            {/if}
          </div>
        {/if}
      </div>
    </section>

    <!-- Right: Terminal -->
    <aside class="studio-terminal" class:collapsed={sidebarCollapsed}>
      <button class="collapse-btn" onclick={() => sidebarCollapsed = !sidebarCollapsed}>
        {sidebarCollapsed ? '◀' : '▶'}
      </button>
      {#if !sidebarCollapsed}
        <TerminalPanel events={jobStore.events} {jobId} />
      {/if}
    </aside>
  </div>
</div>

<style>
  .studio-page {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    height: calc(100dvh - 88px);
  }

  .studio-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-md);
    flex-shrink: 0;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: var(--space-md);
  }

  .studio-title {
    font-family: var(--font-display);
    font-size: 1.125rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--text-primary);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }

  .progress-text {
    font-family: var(--font-code);
    font-size: 0.75rem;
    color: var(--text-secondary);
  }

  .bundle-link {
    padding: 6px 14px;
    border-radius: var(--radius-pill);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-secondary);
    transition: all var(--duration-fast) var(--ease-out);
  }
  .bundle-link:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  .studio-grid {
    display: grid;
    grid-template-columns: 240px 1fr 280px;
    gap: var(--space-md);
    flex: 1;
    min-height: 0;
  }

  .studio-grid.sidebar-collapsed {
    grid-template-columns: 240px 1fr 40px;
  }

  .studio-sidebar {
    min-height: 0;
    overflow: hidden;
  }

  .studio-main {
    display: flex;
    flex-direction: column;
    min-height: 0;
    gap: var(--space-sm);
  }

  .tab-bar {
    display: flex;
    gap: 2px;
    padding: 2px;
    background: var(--bg-surface);
    border-radius: var(--radius-sm);
    width: fit-content;
    flex-shrink: 0;
  }

  .tab-btn {
    padding: 6px 16px;
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    text-transform: capitalize;
    transition: all var(--duration-fast) var(--ease-out);
  }
  .tab-btn:hover { color: var(--text-primary); }
  .tab-btn.active {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .tab-content {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .verification-tab {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: var(--text-tertiary);
    font-style: italic;
  }

  .studio-terminal {
    position: relative;
    min-height: 0;
  }

  .collapse-btn {
    position: absolute;
    top: 0;
    left: -12px;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    font-size: 0.625rem;
    color: var(--text-tertiary);
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  @media (max-width: 1024px) {
    .studio-grid {
      grid-template-columns: 1fr;
    }
    .studio-sidebar, .studio-terminal {
      display: none;
    }
  }
</style>
