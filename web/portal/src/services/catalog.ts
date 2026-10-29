import { apiFetch } from "./auth";

export type Category = { id: string; name: string; slug: string; description: string };
export type Variant = { id: string; sku: string; name: string; attributes: Record<string, string>; price_minor: number; currency: string; inventory: { quantity: number; reserved: number; available: number; version: number } };
export type Product = { id: string; merchant_id: string; category_id: string; category_name: string; name: string; description: string; status: "ACTIVE" | "DRAFT"; variants: Variant[] };
export type Page<T extends string, U> = Record<T, U[]> & { pagination: { limit: number; offset: number; has_more: boolean } };

export async function catalogRequest<T>(path: string, method = "GET", body?: object, pin?: string): Promise<T> {
  const headers: Record<string,string> = { ...(body ? { "Content-Type": "application/json" } : {}), ...(pin ? { "X-PIN": pin } : {}) };
  const response = await apiFetch(`/api/v1/${path}`, { method, headers, ...(body ? { body: JSON.stringify(body) } : {}) });
  if (!response.ok) { const result = await response.json().catch(() => ({})); throw new Error(result.error?.message ?? `Request failed (${response.status})`); }
  return response.status === 204 ? undefined as T : response.json() as Promise<T>;
}
export async function allCategories() {
  const result: Category[] = [];
  for (let offset = 0; ; offset += 100) { const page = await catalogRequest<Page<"categories", Category>>(`categories?limit=100&offset=${offset}`); result.push(...page.categories); if (!page.pagination.has_more) return result; }
}
export function parseAttributes(raw: string) {
  const result: Record<string, string> = {};
  for (const line of raw.split("\n").map(line => line.trim()).filter(Boolean)) {
    const separator = line.indexOf("=");
    if (separator < 1 || !line.slice(separator + 1).trim()) throw new Error("Attributes: use one name=value per line.");
    const key = line.slice(0, separator).trim();
    if (Object.prototype.hasOwnProperty.call(result, key)) throw new Error("Duplicate attribute name.");
    Object.defineProperty(result, key, { value: line.slice(separator + 1).trim(), enumerable: true, configurable: true });
  }
  return result;
}
