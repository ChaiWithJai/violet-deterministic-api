<script lang="ts">
  import { authStore } from '../../lib/stores/auth.svelte';
  import PillButton from '../shared/PillButton.svelte';

  interface Props {
    children: import('svelte').Snippet;
  }

  let { children }: Props = $props();

  let token = $state(authStore.token);
  let baseUrl = $state(authStore.baseUrl);

  function connect() {
    authStore.token = token;
    authStore.baseUrl = baseUrl;
  }
</script>

{#if authStore.isAuthenticated}
  {@render children()}
{:else}
  <div class="auth-gate">
    <div class="auth-card">
      <div class="auth-logo">⬡</div>
      <h1 class="auth-title">Violet Deterministic API</h1>
      <p class="auth-subtitle">Enter your API credentials to connect</p>

      <div class="auth-form">
        <label class="field">
          <span class="field-label">API Base URL</span>
          <input
            type="url"
            class="field-input"
            placeholder="http://localhost:4020"
            bind:value={baseUrl}
          />
        </label>

        <label class="field">
          <span class="field-label">Bearer Token</span>
          <input
            type="password"
            class="field-input"
            placeholder="dev-token"
            bind:value={token}
            onkeydown={(e) => e.key === 'Enter' && connect()}
          />
        </label>

        <PillButton size="lg" onclick={connect} disabled={!token}>
          Connect
        </PillButton>
      </div>

      <p class="auth-hint">
        Dev tokens: <code>dev-token</code> or <code>ops-token</code>
      </p>
    </div>
  </div>
{/if}

<style>
  .auth-gate {
    min-height: 100dvh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-lg);
  }

  .auth-card {
    width: 100%;
    max-width: 420px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-lg);
  }

  .auth-logo {
    font-size: 3rem;
    line-height: 1;
    color: var(--accent);
  }

  .auth-title {
    font-family: var(--font-display);
    font-size: 1.5rem;
    font-weight: 600;
    letter-spacing: -0.03em;
    color: var(--text-primary);
  }

  .auth-subtitle {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin-top: -8px;
  }

  .auth-form {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    text-align: left;
  }

  .field-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .field-input {
    width: 100%;
    padding: 10px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    font-family: var(--font-code);
    font-size: 0.8125rem;
    color: var(--text-primary);
    transition: border-color var(--duration-fast) var(--ease-out);
  }

  .field-input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 2px var(--accent-glow);
  }

  .field-input::placeholder {
    color: var(--text-tertiary);
  }

  .auth-hint {
    font-size: 0.75rem;
    color: var(--text-tertiary);
  }

  .auth-hint code {
    font-family: var(--font-code);
    color: var(--text-secondary);
    background: var(--bg-surface);
    padding: 1px 6px;
    border-radius: 4px;
  }
</style>
