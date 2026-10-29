import Link from "next/link";
import type { LandingCopy } from "./types";

export default function LandingHero({ copy: t }: { copy: LandingCopy }) {
  return <section className="landing-hero"><div className="landing-hero-copy"><div className="landing-eyebrow"><span>✳</span> {t.top}</div><h1>{t.title.split("\n").map((line, index) => <span key={line}>{index === 1 ? <em>{line}</em> : line}<br /></span>)}</h1><p>{t.intro}</p><div className="landing-cta-row"><Link className="landing-primary-cta" href="/customer">{t.shop}<span>→</span></Link><Link className="landing-secondary-cta" href="/auth?mode=register">{t.create}</Link></div><div className="landing-proof"><span>✦</span> {t.benefits}</div></div><div className="landing-hero-visual"><div className="landing-orb"/><div className="landing-visual-label">{t.visualEyebrow}</div><div className="landing-visual-card card-back"><span>✿</span></div><div className="landing-visual-card card-front"><span>Supay</span><strong>{t.visualMessage}</strong><i>✳</i></div><div className="landing-visual-caption">{t.visualCaption}</div></div><div className="landing-hero-index">01 <span>/ 03</span></div></section>;
}
