export type DeterministicTool = {
  name: string;
  method: "POST" | "GET" | "PATCH";
  path: string;
  tenantScoped: true;
  idempotent: boolean;
};

export const TOOL_CONTRACTS: DeterministicTool[] = [
  { name: "plan", method: "POST", path: "/v1/agents/plan", tenantScoped: true, idempotent: true },
  { name: "act", method: "POST", path: "/v1/agents/act", tenantScoped: true, idempotent: true },
  { name: "verify", method: "POST", path: "/v1/agents/verify", tenantScoped: true, idempotent: true },
  { name: "deploy", method: "POST", path: "/v1/agents/deploy", tenantScoped: true, idempotent: true },
];

export const APP_DOMAIN = "saas";
