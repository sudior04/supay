import Link from "next/link";
import PasswordChangeForm from "./PasswordChangeForm";
import type { PortalCopy } from "./types";

export default function CustomerProfile({ copy: t, name, email }: { copy: PortalCopy; name: string; email: string }) {
  return <section className="profile-page">
    <div className="dashboard-welcome"><div><span className="section-kicker">{t.accountDetails}</span><h1>{t.profileTitle}</h1><p>{t.profileDesc}</p></div></div>
    <article className="profile-card"><div className="profile-avatar">{name.slice(0, 1).toUpperCase() || "A"}</div><div><span className="section-kicker">{t.roles.user}</span><h2>{name || t.account}</h2><p>{email}</p></div></article>
    <div className="profile-links"><Link href="/customer/shop"><strong>{t.myOrders}</strong><i aria-hidden="true">→</i></Link><Link href="/customer/shop"><strong>{t.savedItems}</strong><i aria-hidden="true">→</i></Link><Link href="/customer/wallet"><strong>{t.openWallet}</strong><i aria-hidden="true">→</i></Link></div>
    <PasswordChangeForm copy={t} />
  </section>;
}
