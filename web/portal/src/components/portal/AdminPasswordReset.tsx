"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { clearAccessToken, resetUserPassword } from "@/services/auth";
import type { PortalCopy } from "./types";
import AdminPINReset from "./AdminPINReset";

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export default function AdminPasswordReset({ copy: t, currentUserId }: { copy: PortalCopy; currentUserId: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [messageKind, setMessageKind] = useState<"error" | "success">("error");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    const userId = String(form.get("user_id") ?? "").trim();
    const newPassword = String(form.get("new_password") ?? "");
    const confirmation = String(form.get("confirm_password") ?? "");
    const byteLength = new TextEncoder().encode(newPassword).length;
    if (!uuidPattern.test(userId)) { setMessage(t.invalidUserId); setMessageKind("error"); return; }
    if (byteLength < 10 || byteLength > 72) { setMessage(t.passwordInvalid); setMessageKind("error"); return; }
    if (newPassword !== confirmation) { setMessage(t.passwordMismatch); setMessageKind("error"); return; }

    setBusy(true); setMessage("");
    try {
      const { response, error } = await resetUserPassword(userId, newPassword);
      if (response.status === 204) {
        setMessage(t.passwordResetSuccess); setMessageKind("success");
        formElement.reset();
        if (userId.toLowerCase() === currentUserId.toLowerCase()) {
          clearAccessToken();
          router.replace("/auth?passwordChanged=1");
          router.refresh();
        }
        return;
      }
      const code = error?.error?.code;
      const copy = response.status === 401 ? t.passwordSessionExpired
        : response.status === 403 ? t.adminForbidden
        : code === "INVALID_USER_ID" ? t.invalidUserId
        : code === "USER_NOT_FOUND" ? t.userNotFound
        : response.status === 400 ? t.passwordInvalid
        : response.status === 429 ? t.passwordRateLimit
        : t.passwordGenericError;
      setMessage(copy); setMessageKind("error");
    } catch {
      setMessage(t.passwordNetworkError); setMessageKind("error");
    } finally { setBusy(false); }
  }

  return <><section className="password-panel admin-password-panel">
    <div className="password-panel-heading"><h2>{t.adminResetTitle}</h2><p>{t.adminResetDesc}</p></div>
    <form className="password-form" onSubmit={submit}>
      <label>{t.targetUserId}<input name="user_id" inputMode="text" autoComplete="off" required /></label>
      <label>{t.newPassword}<input name="new_password" type="password" autoComplete="new-password" required minLength={10} maxLength={72} /></label>
      <label>{t.confirmPassword}<input name="confirm_password" type="password" autoComplete="new-password" required minLength={10} maxLength={72} /></label>
      {message && <p className={`password-message ${messageKind}`} role="status">{message}</p>}
      <button type="submit" disabled={busy}>{busy ? t.passwordWorking : t.resetPassword}</button>
    </form>
  </section><AdminPINReset currentUserId={currentUserId} /></>;
}
