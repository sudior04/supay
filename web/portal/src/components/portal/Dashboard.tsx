import type { Language } from "@/i18n/provider";
import type { PortalCopy, PortalRole } from "./types";
import PasswordChangeForm from "./PasswordChangeForm";

type Props = { role: PortalRole; copy: PortalCopy; language: Language; money: (amount: number) => string; activeNavIndex?: number; currentUserId?: string };

export default function Dashboard({ role, copy: t, language, money, activeNavIndex = 0, currentUserId = "" }: Props) {
  const merchant = role === "merchant";
  const adminHeading = activeNavIndex === 1 ? t.shopSystem : activeNavIndex === 2 ? t.walletSystem : t.headings.admin;
  const stats = merchant
    ? [[t.today, "128", "+12.8%", "◷"], [t.monthlyRevenue, money(48600000), "+8.2%", "↗"], [t.activeProducts, "246", "+18", "◫"], [t.rating, "4.9 / 5", "+0.2", "☆"]]
    : [[t.users, "24,580", "+8.4%", "♙"], [t.today, "1,284", "+12.8%", "◷"], [t.walletVolume, money(2480000000), "+6.2%", "◈"], [t.uptime, "99.98%", t.status, "✳"]];
  return <>
    <section className="dashboard-welcome"><div><span className="section-kicker">{merchant ? t.manage : "Supay control center"}</span><h1>{merchant ? t.headings.merchant : adminHeading}</h1><p>{merchant ? t.descriptions.merchant : activeNavIndex === 1 ? t.commerceAdminDesc : activeNavIndex === 2 ? t.walletAdminDesc : t.descriptions.admin}</p></div><div className="welcome-date">{language === "vi" ? "Thứ tư, 29 tháng 10" : "Wednesday, October 29"} <span>✳</span></div></section>
    <div className="dashboard-stats">{stats.map(([label, value, change, icon]) => <article className="stat-card" key={label}><div className="stat-top"><span>{label}</span><i>{icon}</i></div><strong>{value}</strong><div className="stat-change"><span>↗ {change}</span> {t.vsPrevious}</div><div className="stat-sparkline"><span/><span/><span/><span/><span/><span/><span/></div></article>)}</div>
    <section className="dashboard-panels"><article className="panel-card revenue-panel"><div className="panel-heading"><div><span className="section-kicker">{t.overview}</span><h2>{merchant ? t.monthlyRevenue : adminHeading}</h2></div><button>{t.period}⌄</button></div><div className="chart-summary"><strong>{money(48620000)}</strong><span>↗ 8.2% {t.vsPrevious}</span></div><div className="chart"><div className="chart-lines"><i/><i/><i/><i/></div><div className="chart-bars">{[35, 52, 43, 68, 54, 77, 61, 86, 72, 94, 68, 82].map((height, index) => <span key={index} style={{ height: `${height}%` }} className={index === 9 ? "highlight-bar" : ""}/>)}</div></div><div className="chart-labels"><span>01 OCT</span><span>07 OCT</span><span>14 OCT</span><span>21 OCT</span><span>29 OCT</span></div></article>
      <article className="panel-card activity-panel"><div className="panel-heading"><div><span className="section-kicker">{t.recent}</span><h2>{t.activities}</h2></div><button>•••</button></div><div className="activity-list">{[["₫", t.newOrder, t.customerName, money(1250000)], ["◈", t.payment, t.walletLine, `+ ${money(850000)}`], ["＋", t.newItems, t.toteLine, t.status], ["✓", t.completedOrder, t.orderLine, money(490000)]].map(([icon, title, detail, value]) => <div className="activity-row" key={title}><span className="activity-icon">{icon}</span><div><strong>{title}</strong><small>{detail}</small></div><b>{value}</b></div>)}</div><button className="activity-more">{t.seeAll} <span>→</span></button></article></section>
    <div className="dashboard-bottom"><div><span className="bottom-icon">✦</span><span><strong>{merchant ? t.suggestion : t.system}</strong><small>{merchant ? t.suggestionDesc : t.normal}</small></span></div><button>{merchant ? t.addNew : t.report}</button></div>
    {role !== "user" && <PasswordChangeForm copy={t} />}
  </>;
}
