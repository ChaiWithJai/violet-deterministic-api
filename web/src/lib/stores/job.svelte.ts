import type { StudioJob, StudioEvent } from '../api/types';
import { createJobStream, type JobStream } from '../api/sse';

function createJobStore() {
  let job = $state<StudioJob | null>(null);
  let events = $state<StudioEvent[]>([]);
  let connected = $state(false);
  let error = $state<string | null>(null);
  let stream: JobStream | null = null;

  return {
    get job() { return job; },
    get events() { return events; },
    get connected() { return connected; },
    get error() { return error; },

    connect(jobId: string) {
      this.disconnect();
      connected = true;
      error = null;
      events = [];

      stream = createJobStream(
        jobId,
        (j) => { job = j; },
        (e) => { events = [...events, e]; },
        (err) => { error = String(err); connected = false; },
      );
    },

    disconnect() {
      stream?.stop();
      stream = null;
      connected = false;
    },

    clear() {
      this.disconnect();
      job = null;
      events = [];
      error = null;
    },
  };
}

export const jobStore = createJobStore();
