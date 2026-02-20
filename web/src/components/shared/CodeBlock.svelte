<script lang="ts">
  interface Props {
    code: string;
    lang?: string;
    maxHeight?: string;
  }

  let { code, lang = '', maxHeight = '400px' }: Props = $props();

  let copied = $state(false);

  function copy() {
    navigator.clipboard.writeText(code);
    copied = true;
    setTimeout(() => { copied = false; }, 1500);
  }
</script>

<div class="code-block">
  <div class="code-header">
    {#if lang}<span class="lang">{lang}</span>{/if}
    <button class="copy-btn" onclick={copy}>
      {copied ? 'Copied' : 'Copy'}
    </button>
  </div>
  <pre style="max-height: {maxHeight}"><code>{code}</code></pre>
</div>

<style>
  .code-block {
    border-radius: var(--radius-md);
    background: rgba(0, 0, 0, 0.4);
    border: 1px solid var(--border-subtle);
    overflow: hidden;
  }

  .code-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-sm) var(--space-md);
    border-bottom: 1px solid var(--border-subtle);
    background: rgba(255, 255, 255, 0.02);
  }

  .lang {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .copy-btn {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    transition: all var(--duration-fast) var(--ease-out);
  }
  .copy-btn:hover {
    background: var(--bg-surface);
    color: var(--text-secondary);
  }

  pre {
    padding: var(--space-md);
    overflow: auto;
    margin: 0;
  }

  code {
    font-family: var(--font-code);
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--text-primary);
    white-space: pre;
  }
</style>
