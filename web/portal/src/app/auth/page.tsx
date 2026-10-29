"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { authMessages } from "@/i18n/auth";
import { useI18n } from "@/i18n/provider";
import AuthBrandPanel from "@/components/auth/AuthBrandPanel";
import AuthFormPanel from "@/components/auth/AuthFormPanel";
import type { AuthMode } from "@/components/auth/types";
import { authenticate, type UserRole } from "@/services/auth";

function destinationFor(role: UserRole) {
  switch (role.toUpperCase()) {
    case "MERCHANT": return "/merchant";
    case "ADMIN": return "/admin";
    default: return "/customer";
  }
}

export default function AuthPage() {
  const router = useRouter();
  const { language, setLanguage } = useI18n();
  const copy = authMessages[language];
  const [mode, setMode] = useState<AuthMode>("login");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("mode") === "register") setMode("register");
    if (params.get("passwordChanged") === "1") setMessage(copy.passwordChanged);
    if (params.get("pinChanged") === "1") setMessage(language === "vi" ? "Đã đổi PIN. Vui lòng đăng nhập lại." : "PIN changed. Please sign in again.");
  }, [copy.passwordChanged, language]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setMessage("");
    const form = new FormData(event.currentTarget);
    const body: { email: string; password: string; name?: string } = { email: String(form.get("email") ?? "").trim(), password: String(form.get("password") ?? "") };
    if (mode === "register") body.name = String(form.get("name") ?? "").trim();
    try {
      const { response, result } = await authenticate(mode, body);
      if (!response.ok) { setMessage(result.error?.message ?? (mode === "login" ? copy.invalidLogin : copy.invalidRegister)); return; }
      if (!result.user?.role) { setMessage(copy.missingRole); return; }
      router.replace(destinationFor(result.user.role));
    } catch {
      setMessage(copy.network);
    } finally {
      setBusy(false);
    }
  }

  function toggleMode() {
    setMode((current) => current === "login" ? "register" : "login");
    setMessage("");
  }

  return <main className="auth-page"><AuthBrandPanel language={language} /><AuthFormPanel copy={copy} language={language} setLanguage={setLanguage} mode={mode} busy={busy} message={message} onSubmit={submit} onToggleMode={toggleMode} /></main>;
}
