"use client";

import { useMemo, type Dispatch, type SetStateAction } from "react";
import type { Language } from "@/i18n/provider";
import { categories, productNames, products } from "./data";
import type { PortalCopy, ProductId } from "./types";

type Props = { copy: PortalCopy; language: Language; category: string; setCategory: Dispatch<SetStateAction<string>>; saved: ProductId[]; setSaved: Dispatch<SetStateAction<ProductId[]>>; onAddToCart: () => void; onNavigateToShop: () => void; showHero?: boolean };

export default function CustomerStorefront({ copy: t, language, category, setCategory, saved, setSaved, onAddToCart, onNavigateToShop, showHero = true }: Props) {
  const visibleProducts = useMemo(() => category === "all" ? products : products.filter((product) => product.category === category), [category]);
  const money = (amount: number) => new Intl.NumberFormat(language === "vi" ? "vi-VN" : "en-US", { style: "currency", currency: "VND", maximumFractionDigits: 0 }).format(amount);
  return <>
    {showHero && <section className="hero-banner"><div className="hero-copy"><div className="hero-pill"><span className="sparkle">✳</span> {t.collection}</div><h1>{t.hero}<br /><em>{t.heroAccent}</em><br />{t.heroEnd}</h1><p>{t.descriptions.user}</p><button className="hero-cta" onClick={() => { onNavigateToShop(); document.getElementById("featured")?.scrollIntoView({ behavior: "smooth" }); }}>{t.exploreCollection} <span>→</span></button><div className="hero-pagination"><span className="page-dot active-dot"/><span className="page-dot"/><span className="page-dot"/><span className="hero-count">01 <i>/</i> 03</span></div></div>
      <div className="hero-art"><div className="sun-disc"/><div className="art-caption"><span>The everyday edit</span><strong>{t.lifestyle}</strong></div><div className="art-bag">✿</div><div className="art-flower flower-one">✿</div><div className="art-flower flower-two">✳</div><div className="art-orbit orbit-one"/><div className="art-orbit orbit-two"/><div className="art-label">SS / 26</div></div><div className="hero-edge">Curated for you · Supay Studio</div>
    </section>}
    {showHero && <div className="trust-strip"><div><span className="trust-icon">◇</span><span><strong>{t.selectedCarefully}</strong><small>{t.story}</small></span></div><div><span className="trust-icon">↗</span><span><strong>{t.nationwide}</strong><small>{t.gentle}</small></span></div><div><span className="trust-icon">♡</span><span><strong>{t.safeShopping}</strong><small>{t.easyReturns}</small></span></div></div>}
    <section className="product-section" id="featured"><div className="section-top"><div><span className="section-kicker">{t.selected}</span><h2>{t.picks}<span>.</span></h2><p>{t.picksDesc}</p></div><button className="view-all" onClick={() => { setCategory("all"); onNavigateToShop(); }}>{t.seeAll} <span>→</span></button></div>
      <div className="filter-row">{["all", ...categories].map((item) => <button key={item} onClick={() => setCategory(item)} className={category === item ? "filter-chip chosen" : "filter-chip"}>{item === "all" ? t.all : t.categories[item as keyof typeof t.categories]}</button>)}</div>
      <div className="product-grid">{visibleProducts.map((product, index) => <article className="product-card" key={product.id}><div className={`product-art ${product.tone}`}><span className="product-tag">{index === 0 ? t.best : index === 2 ? t.loved : t.new}</span><button className={saved.includes(product.id) ? "save-button saved" : "save-button"} aria-label="Save product" onClick={() => setSaved((current) => current.includes(product.id) ? current.filter((item) => item !== product.id) : [...current, product.id])}>{saved.includes(product.id) ? "♥" : "♡"}</button><div className="product-shape"><span>{product.icon}</span></div><span className="art-index">0{index + 1}</span></div><div className="product-info"><span className="product-category">{t.categories[product.category]}</span><h3>{productNames[language][product.id as ProductId]}</h3><div className="product-pricing"><strong>{money(product.price)}</strong><del>{money(product.old)}</del></div><div className="product-meta"><span>★ 4.9 <i>·</i> {t.sold} {product.sold}</span><button aria-label="Add to bag" onClick={onAddToCart}>＋</button></div></div></article>)}</div>
    </section>
  </>;
}
