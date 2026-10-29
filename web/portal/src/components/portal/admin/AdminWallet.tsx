import type { AdminCopy } from "@/i18n/admin";
import type { PortalCopy } from "../types";
import { AdminDataTable, AdminDirectory, AdminMetrics } from "./AdminPrimitives";
import { adminSections, type AdminSection } from "./navigation";

const columns = {
  wallets: ["wallet", "owner", "type", "currency", "balance", "status"],
  transactions: ["reference", "type", "source", "destination", "amount", "status", "time"],
  reconciliation: ["period", "ledger", "actual", "difference", "status"],
  fees: ["transactionType", "calculation", "fee", "effective", "status"],
} as const;

export default function AdminWallet({ copy: t, portalCopy, section }: { copy: AdminCopy; portalCopy: PortalCopy; section: AdminSection }) {
  if (section === "overview") return <><AdminMetrics copy={t} labels={[t.metrics.wallets, t.metrics.balance, t.metrics.transactions, t.metrics.reconciliation]} /><AdminDirectory copy={t} tab={2} sections={adminSections[2]} /></>;
  const selected = section as keyof typeof columns;
  return <AdminDataTable key={section} copy={t} portalCopy={portalCopy} title={t.sections[selected]} columns={columns[selected]} statuses={selected === "wallets" ? ["ACTIVE", "SUSPENDED", "CLOSED"] : selected === "transactions" ? ["PENDING", "COMPLETED", "FAILED", "REFUNDED"] : selected === "reconciliation" ? ["PENDING", "COMPLETED", "FAILED"] : undefined} />;
}
