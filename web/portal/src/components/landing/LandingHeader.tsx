import Link from "next/link";
import type { Language } from "@/i18n/provider";
import type { LandingCopy } from "./types";

export default function LandingHeader({ copy: t, language, setLanguage }: { copy: LandingCopy; language: Language; setLanguage: (language: Language) => void }) {
  return <header className="landing-header"><Link className="brand" href="/"><span className="brand-mark">s</span><span>supay<span className="brand-dot">.</span></span></Link><nav><a href="#platform">{t.learn}</a><a href="#featured">{t.products}</a><a href="#wallet">{t.wallet}</a></nav><div className="landing-actions"><div className="language-switch" aria-label={language === "vi" ? "Chọn ngôn ngữ" : "Choose language"}><button type="button" aria-pressed={language === "vi"} className={language === "vi" ? "language-option active" : "language-option"} onClick={() => setLanguage("vi")}>VI</button><span>/</span><button type="button" aria-pressed={language === "en"} className={language === "en" ? "language-option active" : "language-option"} onClick={() => setLanguage("en")}>EN</button></div><Link href="/auth" className="landing-login">{t.signIn}</Link></div></header>;
}
