import Link from "next/link";
import type { LandingCopy } from "./types";

export default function PlatformOverview({ copy: t }: { copy: LandingCopy }) {
  return <section className="landing-platform" id="platform"><div className="landing-section-heading"><div><span className="landing-kicker">{t.learn}</span><h2>{t.services}</h2></div><span className="landing-section-number">{t.sectionCount}</span></div><div className="platform-grid"><article className="platform-card"><span className="platform-icon">⌕</span><span className="platform-number">01</span><h3>{t.commerce}</h3><p>{t.commerceDesc}</p><Link href="/customer">{t.quickShop} <span>→</span></Link></article><article className="platform-card"><span className="platform-icon">⌂</span><span className="platform-number">02</span><h3>{t.merchants}</h3><p>{t.merchantsDesc}</p><Link href="/merchant">{t.merchantSpace} <span>→</span></Link></article><article className="platform-card wallet-platform-card" id="wallet"><span className="platform-icon">◈</span><span className="platform-number">03</span><h3>{t.wallet}</h3><p>{t.walletDesc}</p><Link href="/auth?mode=register">{t.walletCta} <span>→</span></Link></article></div></section>;
}
