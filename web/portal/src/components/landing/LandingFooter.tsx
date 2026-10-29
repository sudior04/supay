import Link from "next/link";
import type { LandingCopy } from "./types";

export default function LandingFooter({ copy: t }: { copy: LandingCopy }) {
  return <footer className="landing-footer"><Link className="brand" href="/"><span className="brand-mark">s</span><span>supay<span className="brand-dot">.</span></span></Link><span>{t.footer}</span><div><Link href="/auth">{t.account}</Link><span>VI / EN</span></div></footer>;
}
