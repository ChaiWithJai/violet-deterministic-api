export interface AgentPlanRequest {
  prompt: string;
  target: "managed";
}

export interface AgentActRequest {
  mutationClass: string;
  payload: Record<string, unknown>;
}

export interface AgentVerifyResponse {
  verdict: "pass" | "fail";
  checks: Array<{ id: string; status: "pass" | "fail"; evidence: string }>;
}
