import Link from "next/link";
import type { PortalCopy, PortalRole } from "./types";

type Props = { role: PortalRole; title: string; items: string[]; hrefs?: string[]; selectedIndex: number; copy: PortalCopy; onSelect: (index: number) => void };

export default function PortalSidebar({ role, title, items, hrefs, selectedIndex, copy: t, onSelect }: Props) {
  const customer = role === "user";
  return <aside className="sidebar"><div className="sidebar-heading"><span className="sidebar-kicker">{customer ? t.explore : role === "merchant" ? t.manage : t.control}</span><h2>{customer ? t.categoryTitle : title}</h2></div>
    <nav className="sidebar-links" aria-label={title}>{items.map((item, index) => {
      const content = <><span className="side-icon" aria-hidden="true">{["✳", "◈", "⌂", "▧", "✿", "↗"][index % 6]}</span>{item}<span className="side-arrow" aria-hidden="true">›</span></>;
      const className = index === selectedIndex ? "sidebar-link selected" : "sidebar-link";
      return hrefs ? <Link key={hrefs[index]} href={hrefs[index]} className={className} aria-current={index === selectedIndex ? "page" : undefined}>{content}</Link> : <button key={item} className={className} onClick={() => onSelect(index)}>{content}</button>;
    })}</nav>
    {role !== "admin" && <div className="sidebar-note"><span className="note-icon">✦</span><strong>{customer ? t.chooseHelp : t.supportSupay}</strong><p>{customer ? t.discoverDesc : t.supportDesc}</p><button>{customer ? t.browse : t.guide} <span>→</span></button></div>}
    <div className="sidebar-footer">{role === "admin" ? <>Supay <span>·</span> {t.control}</> : <>{t.safeShopping} <span>·</span> {t.nationwide}</>}</div>
  </aside>;
}
