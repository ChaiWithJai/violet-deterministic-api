<script lang="ts">
  interface Props {
    questions: string[];
    onanswer: (answers: Record<string, string>) => void;
    loading?: boolean;
  }

  let { questions, onanswer, loading = false }: Props = $props();

  let answers = $state<Record<string, string>>({});

  function submit() {
    if (!loading) {
      onanswer(answers);
    }
  }

  const quickAnswers: Record<string, string[]> = {
    default: ['Yes', 'No', 'Maybe later'],
  };

  function setQuickAnswer(qIdx: number, answer: string) {
    answers[String(qIdx)] = answer;
  }
</script>

<div class="clarify-chat">
  {#each questions as question, i}
    <div class="qa-card">
      <div class="qa-question">
        <span class="qa-icon">?</span>
        <p>{question}</p>
      </div>
      <div class="qa-answer">
        <div class="quick-pills">
          {#each quickAnswers.default as pill}
            <button
              class="quick-pill"
              class:active={answers[String(i)] === pill}
              onclick={() => setQuickAnswer(i, pill)}
            >
              {pill}
            </button>
          {/each}
        </div>
        <input
          type="text"
          class="qa-input"
          placeholder="Or type a custom answer..."
          bind:value={answers[String(i)]}
        />
      </div>
    </div>
  {/each}

  <button
    class="continue-btn"
    onclick={submit}
    disabled={loading || Object.keys(answers).length < questions.length}
  >
    {loading ? 'Processing...' : 'Continue'}
  </button>
</div>

<style>
  .clarify-chat {
    max-width: 720px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .qa-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .qa-question {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    padding: var(--space-md);
    background: rgba(124, 58, 237, 0.04);
    border-bottom: 1px solid var(--border-subtle);
  }

  .qa-icon {
    flex-shrink: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent-subtle);
    color: var(--accent);
    border-radius: 50%;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .qa-question p {
    font-size: 0.875rem;
    color: var(--text-primary);
    line-height: 1.5;
  }

  .qa-answer {
    padding: var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .quick-pills {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-xs);
  }

  .quick-pill {
    padding: 4px 14px;
    border-radius: var(--radius-pill);
    font-size: 0.75rem;
    font-weight: 500;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-secondary);
    transition: all var(--duration-fast) var(--ease-out);
  }

  .quick-pill:hover {
    border-color: var(--border-medium);
    color: var(--text-primary);
  }

  .quick-pill.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--text-on-accent);
  }

  .qa-input {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-size: 0.8125rem;
    color: var(--text-primary);
  }
  .qa-input:focus {
    border-color: var(--accent);
  }
  .qa-input::placeholder {
    color: var(--text-tertiary);
  }

  .continue-btn {
    align-self: flex-end;
    padding: 10px 28px;
    border-radius: var(--radius-pill);
    background: var(--accent);
    color: var(--text-on-accent);
    font-weight: 600;
    font-size: 0.875rem;
    transition: all var(--duration-fast) var(--ease-out);
  }
  .continue-btn:hover:not(:disabled) {
    background: var(--accent-hover);
    box-shadow: var(--shadow-glow);
  }
  .continue-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
