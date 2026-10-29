"use client";

import Link from "next/link";
import "./portal-header.css";
import type { Language } from "@/i18n/provider";
import type { PortalCopy, PortalRole } from "./types";

type Props = {
  role: PortalRole; copy: PortalCopy; language: Language; setLanguage: (language: Language) => void;
  activeNavIndex: number; navHrefs: readonly string[]; cart: number; accountName: string; onSignOut: () => void;
};

export default function PortalHeader({ role, copy: t, language, setLanguage, activeNavIndex, navHrefs, cart, accountName, onSignOut }: Props) {
  const navigation = t.nav[role];
  return <header className="site-header portal-header">
    <div className="portal-header-row">
      <Link className="brand" href="/" aria-label="Supay"><span className="brand-mark">s</span><span className="portal-brand-name">supay<span className="brand-dot">.</span></span></Link>
      <nav className="portal-header-nav" aria-label={language === "vi" ? "Điều hướng chính" : "Primary navigation"}>{navigation.map((item, index) => <Link key={navHrefs[index]} href={navHrefs[index]} className={activeNavIndex === index ? "portal-header-tab active" : "portal-header-tab"} aria-current={activeNavIndex === index ? "page" : undefined}>{item}</Link>)}</nav>
      {role !== "admin" && <label className="portal-header-search"><span aria-hidden="true">⌕</span><input aria-label={t.search} placeholder={t.search} /></label>}
      <div className="portal-header-tools">
        <span className="role-badge">{t.roles[role]}</span>
        <div className="language-switch" role="group" aria-label={t.selectLanguage}><button type="button" aria-pressed={language === "vi"} className={language === "vi" ? "language-option active" : "language-option"} onClick={() => setLanguage("vi")}>VI</button><span>/</span><button type="button" aria-pressed={language === "en"} className={language === "en" ? "language-option active" : "language-option"} onClick={() => setLanguage("en")}>EN</button></div>
        {role === "user" && <div className="portal-shopping-actions"><button type="button" className="icon-button portal-favorites" aria-label={t.savedItems}>♡</button><button type="button" className="icon-button bag-button" aria-label={language === "vi" ? "Giỏ hàng" : "Shopping bag"}>♧{cart > 0 && <i>{cart}</i>}</button></div>}
        <button type="button" className="account-link account-button" onClick={onSignOut} title={t.logout} aria-label={t.logout}><span className="avatar">{accountName.slice(0, 1).toUpperCase() || "A"}</span><span className="portal-account-label">{t.logout}</span></button>
      </div>
    </div>
  </header>;
}
