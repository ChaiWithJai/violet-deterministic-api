<script lang="ts">
  import type { ClarifyQuestion } from '../../lib/api/types';

  interface Props {
    questions: ClarifyQuestion[];
    onanswer: (answers: Record<string, string>) => void;
    loading?: boolean;
    summary?: string;
  }

  let { questions, onanswer, loading = false, summary = '' }: Props = $props();

  let answers = $state<Record<string, string>>({});

  function submit() {
    if (!loading) {
      onanswer(answers);
    }
  }

  function setAnswer(questionId: string, value: string) {
    answers[questionId] = value;
  }

  let answeredCount = $derived(Object.keys(answers).length);
</script>

<div class="clarify-chat">
  {#if summary}
    <div class="summary-bar">{summary}</div>
  {/if}

  {#each questions as question}
    <div class="qa-card">
      <div class="qa-question">
        <span class="qa-icon">?</span>
        <div class="qa-text">
          <p class="qa-prompt">{question.prompt}</p>
          <p class="qa-why">{question.why}</p>
        </div>
      </div>
      <div class="qa-answer">
        {#if question.options && question.options.length > 0}
          <div class="quick-pills">
            {#each question.options as option}
              <button
                class="quick-pill"
                class:active={answers[question.id] === option}
                onclick={() => setAnswer(question.id, option)}
              >
                {option}
              </button>
            {/each}
          </div>
        {/if}
        <input
          type="text"
          class="qa-input"
          placeholder="Or type a custom answer..."
          value={answers[question.id] ?? ''}
          oninput={(e) => setAnswer(question.id, (e.target as HTMLInputElement).value)}
        />
      </div>
    </div>
  {/each}

  <button
    class="continue-btn"
    onclick={submit}
    disabled={loading || answeredCount < questions.length}
  >
    {loading ? 'Processing...' : `Continue (${answeredCount}/${questions.length})`}
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

  .summary-bar {
    padding: var(--space-sm) var(--space-md);
    background: var(--accent-subtle);
    border: 1px solid var(--border-accent);
    border-radius: var(--radius-md);
    font-size: 0.8125rem;
    color: var(--accent);
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

  .qa-text {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .qa-prompt {
    font-size: 0.875rem;
    color: var(--text-primary);
    line-height: 1.5;
  }

  .qa-why {
    font-size: 0.75rem;
    color: var(--text-tertiary);
    font-style: italic;
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
