import Link from "next/link";
import type { Language } from "@/i18n/provider";
import { featuredItems } from "./data";
import type { LandingCopy } from "./types";

export default function FeaturedProducts({ copy: t, language }: { copy: LandingCopy; language: Language }) {
  const money = (value: number) => new Intl.NumberFormat(language === "vi" ? "vi-VN" : "en-US", { style: "currency", currency: "VND", maximumFractionDigits: 0 }).format(value);
  return <section className="landing-featured" id="featured"><div className="landing-section-heading"><div><span className="landing-kicker">{t.products}</span><h2>{t.productTitle}</h2><p>{t.productDesc}</p></div><Link className="landing-text-link" href="/customer">{t.viewStore} <span>→</span></Link></div><div className="landing-featured-grid">{featuredItems.map((item, index) => <article className="landing-product" key={item.icon}><div className={`landing-product-art ${item.tone}`}><span className="landing-product-label">{t.demo}</span><div>{item.icon}</div><span className="landing-product-index">0{index + 1}</span></div><span className="landing-product-category">{t.categories[index]}</span><h3>{t.productsList[index]}</h3><strong>{money(item.price)}</strong></article>)}</div><p className="demo-note">{t.sampleNote}</p></section>;
}
