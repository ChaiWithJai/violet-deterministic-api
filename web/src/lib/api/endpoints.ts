import { api, idemKey, joinApiUrl } from './client';
import type {
  HealthResponse,
  DecisionRequest,
  DecisionResponse,
  ReplayRequest,
  ReplayResponse,
  FeedbackRequest,
  App,
  CreateAppRequest,
  MutationRequest,
  MutationResponse,
  VerifyResponse,
  DeployIntentResponse,
  AgentPlanRequest,
  AgentPlanResponse,
  AgentClarifyRequest,
  AgentClarifyResponse,
  StudioJobRequest,
  StudioJob,
  StudioRunResponse,
  ArtifactManifest,
  VerificationReport,
  TerminalResponse,
  ConsoleResponse,
  LLMProvidersResponse,
  LLMInferRequest,
  LLMInferResponse,
  Tool,
} from './types';

// --- Health ---
export const getHealth = () => api<HealthResponse>('/v1/health');

// --- Decisions ---
export const createDecision = (req: DecisionRequest) =>
  api<DecisionResponse>('/v1/decisions', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

export const replayDecision = (req: ReplayRequest) =>
  api<ReplayResponse>('/v1/replay', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

export const sendFeedback = (req: FeedbackRequest) =>
  api<{ status: string }>('/v1/feedback', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

// --- Apps ---
export const createApp = (req: CreateAppRequest) =>
  api<App>('/v1/apps', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

export const getApp = (id: string) => api<App>(`/v1/apps/${id}`);

export const patchApp = (id: string, body: Partial<App>) =>
  api<App>(`/v1/apps/${id}`, { method: 'PATCH', body });

export const createMutation = (appId: string, req: MutationRequest) =>
  api<MutationResponse>(`/v1/apps/${appId}/mutations`, {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

export const verifyApp = (appId: string) =>
  api<VerifyResponse>(`/v1/apps/${appId}/verify`, {
    method: 'POST',
    body: {},
    idempotencyKey: idemKey(),
  });

export const deploySelfHost = (appId: string) =>
  api<DeployIntentResponse>(`/v1/apps/${appId}/deploy-intents/self-host`, {
    method: 'POST',
    body: {},
    idempotencyKey: idemKey(),
  });

export const deployManaged = (appId: string) =>
  api<DeployIntentResponse>(`/v1/apps/${appId}/deploy-intents/managed`, {
    method: 'POST',
    body: {},
    idempotencyKey: idemKey(),
  });

// --- Agents ---
export const agentPlan = (req: AgentPlanRequest) =>
  api<AgentPlanResponse>('/v1/agents/plan', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

export const agentClarify = (req: AgentClarifyRequest) =>
  api<AgentClarifyResponse>('/v1/agents/clarify', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

// --- Studio ---
export const createStudioJob = (req: StudioJobRequest) =>
  api<StudioJob>('/v1/studio/jobs', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

export const getStudioJob = (id: string) =>
  api<StudioJob>(`/v1/studio/jobs/${id}`);

export const getStudioArtifacts = (id: string) =>
  api<ArtifactManifest>(`/v1/studio/jobs/${id}/artifacts`);

export const runStudioTarget = (id: string, target: string) =>
  api<StudioRunResponse>(`/v1/studio/jobs/${id}/run`, {
    method: 'POST',
    body: { target },
    idempotencyKey: idemKey(),
  });

export const getStudioVerification = (id: string) =>
  api<VerificationReport>(`/v1/studio/jobs/${id}/verification`);

export const getStudioJTBD = (id: string) =>
  api<{ jtbd_coverage: import('./types').JTBDCoverageItem[] }>(`/v1/studio/jobs/${id}/jtbd`);

export const sendTerminalCommand = (id: string, command: string) =>
  api<TerminalResponse>(`/v1/studio/jobs/${id}/terminal`, {
    method: 'POST',
    body: { command },
    idempotencyKey: idemKey(),
  });

export const getConsoleLogs = (id: string) =>
  api<ConsoleResponse>(`/v1/studio/jobs/${id}/console`);

export const getStudioPreviewUrl = (jobId: string, baseUrl: string) =>
  joinApiUrl(baseUrl, `/v1/studio/jobs/${jobId}/preview`);

export const getStudioBundleUrl = (jobId: string, baseUrl: string) =>
  joinApiUrl(baseUrl, `/v1/studio/jobs/${jobId}/bundle`);

// --- LLM ---
export const getLLMProviders = () =>
  api<LLMProvidersResponse>('/v1/llm/providers');

export const llmInfer = (req: LLMInferRequest) =>
  api<LLMInferResponse>('/v1/llm/infer', {
    method: 'POST',
    body: req,
    idempotencyKey: idemKey(),
  });

// --- Tools ---
export const getTools = () => api<{ tools: Tool[] }>('/v1/tools');
