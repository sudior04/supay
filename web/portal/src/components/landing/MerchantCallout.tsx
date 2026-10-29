import Link from "next/link";
import type { LandingCopy } from "./types";

export default function MerchantCallout({ copy: t }: { copy: LandingCopy }) {
  return <section className="merchant-callout"><div><span className="landing-kicker">{t.merchantLabel}</span><h2>{t.join}</h2><p>{t.joinDesc}</p><Link href="/auth">{t.joinCta} <span>→</span></Link></div><div className="merchant-callout-art"><span>✳</span><strong>your<br />little<br />corner.</strong><i>Supay stores</i></div></section>;
}
