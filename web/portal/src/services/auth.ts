export type UserRole = "CUSTOMER" | "MERCHANT" | "ADMIN" | string;
export type AuthUser = { id: string; name: string; email: string; role: UserRole };
export type AuthSession = { access_token: string; user: AuthUser };
export type AuthApiError = { error?: { message?: string; code?: string } };
export type AuthAction = "login" | "register";
export type AuthResult = Partial<AuthSession> & AuthApiError;

let accessToken: string | null = null;
let refreshInFlight: Promise<{ response: Response; session: AuthSession | null }> | null = null;

export function clearAccessToken() {
  accessToken = null;
}

function setAccessToken(token: string | null | undefined) {
  accessToken = token || null;
}

async function readJson<T>(response: Response): Promise<T> {
  return response.json() as Promise<T>;
}

export async function authenticate(mode: AuthAction, payload: { email: string; password: string; name?: string; pin?: string }) {
  const response = await fetch(`/api/v1/auth/${mode}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const result = await readJson<AuthResult>(response);
  if (response.ok) setAccessToken(result.access_token);
  return { response, result };
}

export async function apiFetch(path: string, init: RequestInit = {}) {
  if (!accessToken) { const refreshed = await refreshSession(); if (!refreshed.session) throw new Error("Please sign in again."); }
  const request = () => fetch(path, { ...init, headers: { ...Object.fromEntries(new Headers(init.headers)), Authorization: `Bearer ${accessToken}` } });
  let response = await request();
  if (response.status === 401) {
    const result = await response.clone().json().catch(() => ({}));
    if (["UNAUTHORIZED", "INVALID_TOKEN"].includes(result.error?.code)) {
      const refreshed = await refreshSession(); if (refreshed.session) response = await request();
    }
  }
  return response;
}

export async function refreshSession() {
  if (refreshInFlight) return refreshInFlight;
  const pending = (async () => {
    const response = await fetch("/api/v1/auth/refresh", { method: "POST" });
    if (!response.ok) {
      clearAccessToken();
      return { response, session: null };
    }
    const session = await readJson<AuthSession>(response);
    setAccessToken(session.access_token);
    return { response, session };
  })();
  refreshInFlight = pending;
  try {
    return await pending;
  } finally {
    if (refreshInFlight === pending) refreshInFlight = null;
  }
}

export async function signOut() {
  try {
    return await fetch("/api/v1/auth/logout", { method: "POST" });
  } finally {
    clearAccessToken();
  }
}

async function authenticatedRequest(path: string, method: "PUT" | "POST", body: object) {
  if (!accessToken) throw new Error("Your session has expired. Please sign in again.");
  return fetch(path, {
    method,
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(body),
  });
}

export async function changeOwnPassword(currentPassword: string, newPassword: string) {
  const response = await authenticatedRequest("/api/v1/users/me/password", "PUT", { current_password: currentPassword, new_password: newPassword });
  const error = response.ok ? null : await readJson<AuthApiError>(response).catch((): AuthApiError => ({}));
  return { response, error };
}

export async function resetUserPassword(userId: string, newPassword: string) {
  const response = await authenticatedRequest(`/api/v1/users/${encodeURIComponent(userId)}/password/reset`, "POST", { new_password: newPassword });
  const error = response.ok ? null : await readJson<AuthApiError>(response).catch((): AuthApiError => ({}));
  return { response, error };
}
