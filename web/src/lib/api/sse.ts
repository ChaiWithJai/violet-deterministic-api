import type { StudioEvent, StudioJob } from './types';
import { authStore } from '../stores/auth.svelte';

export interface JobStream {
  stop(): void;
}

export function createJobStream(
  jobId: string,
  onJob: (job: StudioJob) => void,
  onEvent: (event: StudioEvent) => void,
  onError?: (err: unknown) => void,
): JobStream {
  let stopped = false;
  let es: EventSource | null = null;
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  const eventsUrl = `${authStore.baseUrl}/v1/studio/jobs/${jobId}/events?token=${encodeURIComponent(authStore.token)}`;
  const jobUrl = `${authStore.baseUrl}/v1/studio/jobs/${jobId}`;

  function fetchJob() {
    fetch(jobUrl, {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
      .then((r) => r.json())
      .then((data) => {
        if (!stopped) onJob(data);
      })
      .catch((err) => {
        if (!stopped && onError) onError(err);
      });
  }

  function startPolling() {
    if (pollTimer) return;
    pollTimer = setInterval(() => {
      if (stopped) {
        if (pollTimer) clearInterval(pollTimer);
        return;
      }
      fetchJob();
    }, 3000);
  }

  function connect() {
    if (stopped) return;

    try {
      es = new EventSource(eventsUrl);

      es.onmessage = (e) => {
        try {
          const event: StudioEvent = JSON.parse(e.data);
          onEvent(event);
        } catch {
          // ignore parse errors
        }
      };

      es.addEventListener('job', (e) => {
        try {
          const job: StudioJob = JSON.parse((e as MessageEvent).data);
          onJob(job);
        } catch {
          // ignore
        }
      });

      es.onerror = () => {
        if (stopped) return;
        es?.close();
        es = null;
        // Fallback to polling
        startPolling();
        // Retry SSE after 5s
        setTimeout(connect, 5000);
      };
    } catch {
      // EventSource not available, use polling
      startPolling();
    }
  }

  // Initial fetch, then start SSE
  fetchJob();
  connect();

  return {
    stop() {
      stopped = true;
      es?.close();
      if (pollTimer) clearInterval(pollTimer);
    },
  };
}
