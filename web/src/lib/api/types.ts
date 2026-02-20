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

export interface VerifyCheck {
  id: string;
  status: 'pass' | 'fail';
  evidence: string;
}

export interface VerifyResponse {
  report_id: string;
  app_id: string;
  tenant_id: string;
  verdict: string;
  checks: VerifyCheck[];
  policy_version: string;
  data_version: string;
  generated_at: string;
}

export interface DeployIntentResponse {
  intent_id: string;
  app_id: string;
  tenant_id: string;
  target: string;
  approval_required: boolean;
  status: string;
  profile: Record<string, unknown>;
  policy_version: string;
  data_version: string;
  requested_at: string;
  orchestration_hints: { next: string[] };
}

// --- Agents ---
export interface AgentPlanRequest {
  prompt: string;
  name?: string;
}

export interface AgentPlanResponse {
  plan_id: string;
  tenant_id: string;
  name: string;
  suggested_blueprint: {
    plan: string;
    region: string;
  };
  checks: string[];
  policy_version: string;
  data_version: string;
}

export interface ClarifyQuestion {
  id: string;
  field: string;
  prompt: string;
  why: string;
  options?: string[];
}

export interface Confirmation {
  prompt: string;
  app_name: string;
  domain: string;
  template: string;
  source_system: string;
  primary_users: string[];
  core_workflows: string[];
  data_entities: string[];
  deployment_target: string;
  region: string;
  plan: string;
  generation_depth: 'prototype' | 'pilot' | 'production-candidate';
  integrations: string[];
  constraints: string[];
}

export interface AgentClarifyRequest {
  prompt: string;
  confirmation: Partial<Confirmation>;
  answers?: Record<string, string>;
}

export interface AgentClarifyResponse {
  clarification_id: string;
  tenant_id: string;
  answer_count: number;
  ready_to_generate: boolean;
  remaining_questions: number;
  missing_fields: string[];
  summary: string;
  updated_confirmation: Confirmation;
  questions: ClarifyQuestion[];
  policy_version: string;
  data_version: string;
}

export interface AgentActResponse {
  mutation_id: string;
  policy_version: string;
  app: App;
  actor: string;
  subject: string;
}

export interface AgentVerifyResponse {
  report_id: string;
  app_id: string;
  tenant_id: string;
  verdict: string;
  checks: VerifyCheck[];
  policy_version: string;
  data_version: string;
  generated_at: string;
  actor: string;
  subject: string;
}

export interface AgentDeployResponse {
  intent_id: string;
  app_id: string;
  tenant_id: string;
  target: string;
  approval_required: boolean;
  status: string;
  profile: Record<string, unknown>;
  policy_version: string;
  data_version: string;
  requested_at: string;
  orchestration_hints: { next: string[] };
  actor: string;
  subject: string;
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
  generation_depth?: 'prototype' | 'pilot' | 'production-candidate';
}

export interface WorkloadItem {
  phase: string;
  task: string;
  owner: string;
  estimate_hours: number;
  status: string;
}

export interface FileArtifact {
  path: string;
  language: string;
  content: string;
}

export interface ManifestFile {
  path: string;
  language: string;
  category: string;
  size_bytes: number;
}

export interface RunTarget {
  name: string;
  description: string;
  command: string;
}

export interface ArtifactManifest {
  generated_at: string;
  workspace_path: string;
  files: ManifestFile[];
  run_targets: RunTarget[];
}

export interface VerificationReport {
  report_id: string;
  verdict: string;
  checks: VerifyCheck[];
  depth_label: 'prototype' | 'pilot' | 'production-candidate';
  behavioral_pass_rate: number;
  generated_at: string;
}

export interface JTBDCoverageItem {
  id: string;
  task: string;
  status: 'pass' | 'fail';
  evidence: string;
}

export interface StudioJob {
  job_id: string;
  tenant_id: string;
  status: string;
  depth_label: 'prototype' | 'pilot' | 'production-candidate';
  created_at: string;
  updated_at: string;
  workspace_path: string;
  confirmation: Confirmation;
  workload: WorkloadItem[];
  files: FileArtifact[];
  artifact_manifest: ArtifactManifest;
  verification_report: VerificationReport;
  jtbd_coverage: JTBDCoverageItem[];
  terminal_logs: string[];
  console_logs: string[];
  preview_workload: string;
  preview_code_path: string;
  preview_terminal: string;
  preview_console: string;
}

export interface StudioRunResponse {
  target: string;
  status: 'pass' | 'fail';
  checks: VerifyCheck[];
  generated_at: string;
}

export interface TerminalResponse {
  command: string;
  output: string[];
  cwd: string;
}

export interface ConsoleResponse {
  logs: string[];
}

// --- LLM ---
export interface LLMModel {
  name: string;
  context_window: number;
}

export interface LLMProvider {
  name: string;
  status: string;
  models: LLMModel[];
}

export interface LLMProvidersResponse {
  default_provider: string;
  default_model: string;
  providers: LLMProvider[];
}

export interface LLMInferRequest {
  provider?: string;
  model?: string;
  prompt: string;
  system?: string;
  max_tokens?: number;
  temperature?: number;
}

export interface LLMInferResult {
  text: string;
  provider: string;
  model: string;
  tokens_input: number;
  tokens_output: number;
  inference_ms: number;
}

export interface LLMInferResponse {
  tenant_id: string;
  result: LLMInferResult;
  hooks: unknown[];
}

// --- Tools ---
export interface Tool {
  name: string;
  description: string;
  method: string;
  path: string;
  cli: string;
}

// --- SSE Events ---
export interface StudioEvent {
  type: string;
  phase?: string;
  message?: string;
  data?: unknown;
}

// --- API Response wrapper ---
export interface ApiResult<T> {
  ok: boolean;
  status: number;
  data: T;
}
