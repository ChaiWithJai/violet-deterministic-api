<script lang="ts">
  interface Props {
    previewUrl: string;
  }

  let { previewUrl }: Props = $props();

  let activeView = $state<'web' | 'mobile'>('web');
</script>

<div class="preview-panes">
  <div class="preview-toolbar">
    <button class="view-btn" class:active={activeView === 'web'} onclick={() => activeView = 'web'}>
      Desktop
    </button>
    <button class="view-btn" class:active={activeView === 'mobile'} onclick={() => activeView = 'mobile'}>
      Mobile
    </button>
  </div>

  <div class="preview-container" class:mobile-view={activeView === 'mobile'}>
    {#if activeView === 'mobile'}
      <div class="phone-chrome">
        <div class="phone-notch"></div>
        <iframe src={previewUrl} title="Mobile preview" class="preview-frame mobile"></iframe>
      </div>
    {:else}
      <iframe src={previewUrl} title="Desktop preview" class="preview-frame desktop"></iframe>
    {/if}
  </div>
</div>

<style>
  .preview-panes {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    height: 100%;
  }

  .preview-toolbar {
    display: flex;
    gap: 2px;
    padding: 2px;
    background: var(--bg-surface);
    border-radius: var(--radius-sm);
    width: fit-content;
  }

  .view-btn {
    padding: 6px 16px;
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    transition: all var(--duration-fast) var(--ease-out);
  }
  .view-btn:hover { color: var(--text-primary); }
  .view-btn.active {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .preview-container {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .preview-frame {
    border: none;
    background: white;
  }

  .preview-frame.desktop {
    width: 100%;
    height: 100%;
  }

  .phone-chrome {
    width: 320px;
    height: 560px;
    border: 2px solid var(--border-medium);
    border-radius: 32px;
    padding: 8px;
    background: #000;
    display: flex;
    flex-direction: column;
    align-items: center;
    overflow: hidden;
  }

  .phone-notch {
    width: 100px;
    height: 24px;
    background: #000;
    border-radius: 0 0 16px 16px;
    margin-bottom: 4px;
    flex-shrink: 0;
  }

  .preview-frame.mobile {
    width: 100%;
    flex: 1;
    border-radius: 0 0 24px 24px;
  }
</style>
