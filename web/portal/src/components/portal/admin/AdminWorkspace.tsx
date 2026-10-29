import "./admin.css";
import { adminMessages } from "@/i18n/admin";
import type { Language } from "@/i18n/provider";
import type { AuthUser } from "@/services/auth";
import type { PortalCopy } from "../types";
import PasswordChangeForm from "../PasswordChangeForm";
import AdminOverview from "./AdminOverview";
import AdminUsers from "./AdminUsers";
import AdminCommerce from "./AdminCommerce";
import AdminWallet from "./AdminWallet";
import { AdminDataTable, AdminHeading } from "./AdminPrimitives";
import type { AdminSection } from "./navigation";

export default function AdminWorkspace({ tab, section, language, portalCopy, account }: { tab: number; section: AdminSection; language: Language; portalCopy: PortalCopy; account: AuthUser }) {
  const t = adminMessages[language];
  const group = tab === 1 ? t.groups.commerce : tab === 2 ? t.groups.wallet : t.groups.system;
  return <section className="admin-workspace"><AdminHeading title={section === "overview" ? group.title : t.sections[section]} description={section === "overview" ? group.description : t.descriptions[section as keyof typeof t.descriptions]} kicker={group.kicker} />
    {tab === 1 ? <AdminCommerce copy={t} portalCopy={portalCopy} section={section} />
      : tab === 2 ? <AdminWallet copy={t} portalCopy={portalCopy} section={section} />
      : section === "users" ? <AdminUsers copy={t} portalCopy={portalCopy} currentUserId={account.id} />
      : section === "activity" ? <AdminDataTable copy={t} portalCopy={portalCopy} title={t.sections.activity} columns={["actor", "action", "target", "time"]} />
      : section === "security" ? <><article className="admin-account-card"><h2>{t.currentAccount}</h2><p>{t.currentData}</p><dl><div><dt>{t.columns.name}</dt><dd>{account.name || t.unknown}</dd></div><div><dt>{t.columns.email}</dt><dd>{account.email}</dd></div><div><dt>{t.columns.role}</dt><dd>{portalCopy.roles.admin}</dd></div><div><dt>{t.userId}</dt><dd>{account.id}</dd></div></dl></article><PasswordChangeForm copy={portalCopy} /></>
      : <AdminOverview copy={t} />}
  </section>;
}
