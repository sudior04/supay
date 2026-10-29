import type { PortalCopy } from "./types";
import ProductManager from "./catalog/ProductManager";

type Props = { tab: 1 | 2 | 3; copy: PortalCopy; money: (amount: number) => string };
const orders = [
  { id: "SP-24081", buyer: "Nguyễn Minh Anh", date: "29/10/2026", amount: 1250000, statusKey: "orderPending" },
  { id: "SP-24080", buyer: "Trần Gia Hân", date: "29/10/2026", amount: 329000, statusKey: "orderShipping" },
  { id: "SP-24079", buyer: "Lê Hoàng Nam", date: "28/10/2026", amount: 790000, statusKey: "orderCompleted" },
] as const;

export default function MerchantWorkspace({ tab, copy: t, money }: Props) {
  if (tab === 1) return <ProductManager />;
  const title = tab === 2 ? t.ordersWorkspace : t.storeWorkspace;
  const description = tab === 2 ? t.ordersWorkspaceDesc : t.storeWorkspaceDesc;
  return <section className="workspace-page"><div className="workspace-heading"><div><span className="section-kicker">{t.manage}</span><h1>{title}</h1><p>{description}</p></div></div>
    {tab === 2 && <div className="workspace-table-wrap"><table className="workspace-table"><thead><tr><th>{t.orderNumber}</th><th>{t.buyer}</th><th>{t.orderDate}</th><th>{t.amount}</th><th>{t.productStatus}</th><th/></tr></thead><tbody>{orders.map((order) => <tr key={order.id}><td>#{order.id}</td><td>{order.buyer}</td><td>{order.date}</td><td>{money(order.amount)}</td><td><span className="table-status">{t[order.statusKey]}</span></td><td><button className="table-more" aria-label={t.viewDetails}>•••</button></td></tr>)}</tbody></table></div>}
    {tab === 3 && <div className="store-summary-grid"><article><span>{t.monthlyRevenue}</span><strong>{money(48600000)}</strong><small>↗ +8.2% {t.vsPrevious}</small></article><article><span>{t.activeProducts}</span><strong>246</strong><small>{t.status}</small></article><article><span>{t.transactions}</span><strong>84</strong><small>{t.cashflow}</small></article></div>}
  </section>;
}
