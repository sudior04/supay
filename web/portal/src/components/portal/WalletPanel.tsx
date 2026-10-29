import type { PortalCopy } from "./types";

type Props = { copy: PortalCopy; money: (amount: number) => string; merchant: boolean };

export default function WalletPanel({ copy: t, money, merchant }: Props) {
  const transactions = [
    ["＋", t.deposit, merchant ? "+ ₫ 1.250.000" : "+ ₫ 850.000", "29/10/2026 · 10:42"],
    ["↗", t.storeTransaction, "− ₫ 329.000", "28/10/2026 · 16:18"],
    ["⇧", t.transfer, "− ₫ 500.000", "27/10/2026 · 09:30"],
  ];
  return <section className="wallet-page"><div className="wallet-intro"><span className="section-kicker">{merchant ? t.finance : t.wallet}</span><h1>{merchant ? t.finance : t.wallet}</h1><p>{merchant ? t.merchantWalletDesc : t.walletDesc}</p></div>
    <article className="wallet-balance"><div className="wallet-balance-top"><span>{t.available}</span><span className="wallet-brand-mark">s</span></div><strong>{money(merchant ? 48620000 : 5280000)}</strong><div className="wallet-balance-foot"><span>{merchant ? t.income : t.balance}</span><span>•••• &nbsp; 2841</span></div><div className="wallet-actions"><button>{t.topup} <span>＋</span></button><button>{t.withdraw} <span>↗</span></button></div></article>
    <div className="wallet-metrics"><article><span>{t.income}</span><strong>{money(merchant ? 12840000 : 1250000)}</strong><small>↗ 8.2% {t.vsPrevious}</small></article><article><span>{t.transactions}</span><strong>{merchant ? "84" : "26"}</strong><small>{t.status}</small></article></div>
    <article className="wallet-transactions"><div className="panel-heading"><div><span className="section-kicker">{t.recent}</span><h2>{t.transactions}</h2></div><button>{t.seeAll} →</button></div>{transactions.map(([icon, title, amount, date]) => <div className="wallet-transaction" key={title}><span className="transaction-icon">{icon}</span><div><strong>{title}</strong><small>{date}</small></div><b className={amount.startsWith("+") ? "positive" : ""}>{amount}</b></div>)}</article>
  </section>;
}
