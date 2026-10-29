"use client";

import Link from "next/link";
import type { FormEvent } from "react";
import type { Language } from "@/i18n/provider";
import type { AuthCopy, AuthMode } from "./types";

type Props = { copy: AuthCopy; language: Language; setLanguage: (language: Language) => void; mode: AuthMode; busy: boolean; message: string; onSubmit: (event: FormEvent<HTMLFormElement>) => void; onToggleMode: () => void };

export default function AuthFormPanel({ copy: t, language, setLanguage, mode, busy, message, onSubmit, onToggleMode }: Props) {
  return <section className="auth-card"><div className="auth-card-top"><Link href="/" className="auth-back">{t.back}</Link><div className="language-switch" role="group" aria-label={t.language}><button type="button" aria-pressed={language === "vi"} className={language === "vi" ? "language-option active" : "language-option"} onClick={() => setLanguage("vi")}>VI</button><span>/</span><button type="button" aria-pressed={language === "en"} className={language === "en" ? "language-option active" : "language-option"} onClick={() => setLanguage("en")}>EN</button></div></div><p className="eyebrow">{t.eyebrow}</p><h1>{mode === "login" ? t.loginTitle : t.registerTitle}</h1><p className="auth-description">{mode === "login" ? t.loginDescription : t.registerDescription}</p>
    <form onSubmit={onSubmit} className="auth-form">
      {mode === "register" && <label>{t.name}<input name="name" autoComplete="name" placeholder={t.namePlaceholder} required maxLength={120} /></label>}
      <label>{t.email}<input name="email" type="email" autoComplete="email" placeholder={t.emailPlaceholder} required /></label>
      <label>{t.password}<input name="password" type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} placeholder={mode === "register" ? t.passwordHint : t.passwordPlaceholder} required minLength={mode === "register" ? 10 : undefined} maxLength={72} /></label>
      <button disabled={busy}>{busy ? t.working : mode === "login" ? t.login : t.register}<span>→</span></button>
    </form>
    {message && <p className="auth-message" role="status">{message}</p>}<p className="auth-switch">{mode === "login" ? t.noAccount : t.hasAccount} <button className="text-button" onClick={onToggleMode}>{mode === "login" ? t.registerLink : t.loginLink}</button></p><div className="auth-wallet-note"><span>◈</span><p>{t.walletNote}</p></div>
  </section>;
}
