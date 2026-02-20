<script lang="ts">
  import type { VerificationReport as VerificationReportType } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    report: VerificationReportType;
  }

  let { report }: Props = $props();
</script>

<div class="verification-report">
  <div class="report-header">
    <span class="report-id">Report {report.report_id}</span>
    <Badge variant={report.verdict === 'pass' ? 'pass' : 'fail'}>
      {report.verdict}
    </Badge>
  </div>
  <div class="report-meta">
    <span class="meta-time">Generated: {report.generated_at}</span>
  </div>
  <div class="report-checks">
    {#each report.checks as check}
      <div class="check-row">
        <span class="check-status" class:pass={check.status === 'pass'} class:fail={check.status !== 'pass'}>
          {check.status === 'pass' ? '\u2713' : '\u2717'}
        </span>
        <span class="check-id">{check.id}</span>
        <span class="check-evidence">{check.evidence}</span>
      </div>
    {/each}
  </div>
</div>

<style>
  .verification-report {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .report-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .report-id {
    font-family: var(--font-code);
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .report-meta {
    font-size: 0.75rem;
    color: var(--text-tertiary);
  }

  .report-checks {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .check-row {
    display: flex;
    align-items: baseline;
    gap: var(--space-sm);
    font-size: 0.8125rem;
  }

  .check-status {
    font-size: 0.75rem;
    font-weight: 700;
    flex-shrink: 0;
  }
  .check-status.pass { color: var(--pass); }
  .check-status.fail { color: var(--fail); }

  .check-id {
    color: var(--text-primary);
    font-weight: 500;
  }

  .check-evidence {
    color: var(--text-tertiary);
    font-size: 0.75rem;
  }
</style>
