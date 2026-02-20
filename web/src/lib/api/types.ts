// --- Auth ---
export interface HealthResponse {
  status: string;
  service: string;
  policy_version: string;
  data_version: string;
  idempotency_cleanup_deleted_total: number;
}

// --- Decisions ---
export interface Candidate {
  item_id: string;
  score?: number;
  tags?: string[];
}

export interface DecisionRequest {
  tenant_id?: string;
  user_id: string;
  context_keys: Record<string, string>;
  candidates: Candidate[];
  limit?: number;
}

export interface RankedItem {
  item_id: string;
  score: number;
  rank: number;
  source: string;
  tags?: string[];
}

export interface StageResult {
  stage: string;
  items: RankedItem[];
  latency_ms: number;
  error?: string;
}

export interface DecisionResponse {
  decision_id: string;
  decision_hash: string;
  policy_version: string;
  data_version: string;
  ranked_items: RankedItem[];
  stages: StageResult[];
  dependency_status: 'healthy' | 'degraded';
  latency_ms: number;
  timestamp: string;
}

export interface ReplayRequest {
  decision_id: string;
}

export interface ReplayResponse {
  original: DecisionResponse;
  replayed: DecisionResponse;
  hashes_match: boolean;
}

export interface FeedbackRequest {
  tenant_id?: string;
  user_id: string;
  item_id: string;
  feedback_type: string;
}

// --- Apps ---
export interface App {
  id: string;
  tenant_id: string;
  name: string;
  plan: string;
  region: string;
  feature_flags: Record<string, boolean>;
  blueprint: Record<string, unknown>;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface CreateAppRequest {
  name: string;
  plan?: string;
  region?: string;
}

export interface MutationRequest {
  mutation_class: string;
  path: string;
  value: unknown;
}

export interface MutationResponse {
  app_id: string;
  mutation_class: string;
  path: string;
  value: unknown;
  version: number;
  policy_check: 'pass' | 'fail';
  policy_reason?: string;
}

export interface VerifyResponse {
  app_id: string;
  status: 'pass' | 'fail';
  checks: { name: string; status: string; message?: string }[];
}

export interface DeployIntentResponse {
  app_id: string;
  deploy_type: string;
  status: string;
  approval_required: boolean;
}

// --- Agents ---
export interface AgentPlanRequest {
  prompt: string;
  context?: Record<string, unknown>;
}

export interface AgentPlanResponse {
  plan_id: string;
  steps: { step: number; action: string; description: string }[];
  ready_to_generate: boolean;
  clarify_questions?: string[];
}

export interface AgentClarifyRequest {
  plan_id: string;
  answers: Record<string, string>;
}

export interface AgentClarifyResponse {
  plan_id: string;
  ready_to_generate: boolean;
  clarify_questions?: string[];
  updated_steps?: { step: number; action: string; description: string }[];
}

// --- Studio ---
export interface StudioJobRequest {
  prompt: string;
  app_name?: string;
  domain?: string;
  template?: string;
  source_system?: string;
  primary_users?: string[];
  core_workflows?: string[];
  data_entities?: string[];
  deployment_target?: string;
  region?: string;
}

export interface StudioJob {
  id: string;
  tenant_id: string;
  prompt: string;
  status: string;
  phase: string;
  progress: number;
  artifacts?: Record<string, string>;
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface StudioEvent {
  type: string;
  phase?: string;
  progress?: number;
  message?: string;
  data?: unknown;
}

export interface VerificationResult {
  target: string;
  status: 'pass' | 'fail' | 'pending';
  checks: { name: string; status: string; message?: string }[];
}

export interface JTBDCoverage {
  jtbd: string;
  covered: boolean;
  evidence?: string;
}

// --- LLM ---
export interface LLMProvider {
  name: string;
  type: string;
  models: string[];
  default_model: string;
  status: string;
}

export interface LLMInferRequest {
  provider?: string;
  model?: string;
  prompt: string;
  system?: string;
  max_tokens?: number;
  temperature?: number;
}

export interface LLMInferResponse {
  provider: string;
  model: string;
  content: string;
  usage: { prompt_tokens: number; completion_tokens: number };
  latency_ms: number;
}

// --- Tools ---
export interface Tool {
  name: string;
  description: string;
  category: string;
  parameters: Record<string, unknown>;
}

// --- API Response wrapper ---
export interface ApiResult<T> {
  ok: boolean;
  status: number;
  data: T;
}
