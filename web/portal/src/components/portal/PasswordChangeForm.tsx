"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { clearAccessToken, changeOwnPassword } from "@/services/auth";
import type { PortalCopy } from "./types";
import PINSettings from "./PINSettings";

export default function PasswordChangeForm({ copy: t }: { copy: PortalCopy }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [messageKind, setMessageKind] = useState<"error" | "success">("error");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const currentPassword = String(form.get("current_password") ?? "");
    const newPassword = String(form.get("new_password") ?? "");
    const confirmation = String(form.get("confirm_password") ?? "");
    const byteLength = new TextEncoder().encode(newPassword).length;
    if (byteLength < 10 || byteLength > 72) { setMessage(t.passwordInvalid); setMessageKind("error"); return; }
    if (newPassword !== confirmation) { setMessage(t.passwordMismatch); setMessageKind("error"); return; }

    setBusy(true); setMessage("");
    try {
      const { response, error } = await changeOwnPassword(currentPassword, newPassword);
      if (response.status === 204) {
        clearAccessToken();
        router.replace("/auth?passwordChanged=1");
        router.refresh();
        return;
      }
      const code = error?.error?.code;
      const copy = code === "INVALID_CREDENTIALS" ? t.passwordCurrentInvalid
        : code === "PASSWORD_UNCHANGED" ? t.passwordUnchanged
        : response.status === 400 ? t.passwordInvalid
        : response.status === 429 ? t.passwordRateLimit
        : response.status === 401 ? t.passwordSessionExpired
        : t.passwordGenericError;
      setMessage(copy); setMessageKind("error");
    } catch {
      setMessage(t.passwordNetworkError); setMessageKind("error");
    } finally { setBusy(false); }
  }

  return <><section className="password-panel">
    <div className="password-panel-heading"><h2>{t.passwordSettingsTitle}</h2><p>{t.passwordSettingsDesc}</p></div>
    <form className="password-form" onSubmit={submit}>
      <label>{t.currentPassword}<input name="current_password" type="password" autoComplete="current-password" required maxLength={72} /></label>
      <label>{t.newPassword}<input name="new_password" type="password" autoComplete="new-password" required minLength={10} maxLength={72} /></label>
      <label>{t.confirmPassword}<input name="confirm_password" type="password" autoComplete="new-password" required minLength={10} maxLength={72} /></label>
      {message && <p className={`password-message ${messageKind}`} role="status">{message}</p>}
      <button type="submit" disabled={busy}>{busy ? t.passwordWorking : t.changePassword}</button>
    </form>
  </section><PINSettings /></>;
}
