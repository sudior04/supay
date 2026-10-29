"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { portalMessages } from "@/i18n/portal";
import { adminMessages } from "@/i18n/admin";
import { useI18n } from "@/i18n/provider";
import CustomerStorefront from "@/components/portal/CustomerStorefront";
import Dashboard from "@/components/portal/Dashboard";
import PortalHeader from "@/components/portal/PortalHeader";
import PortalFooter from "@/components/portal/PortalFooter";
import PortalSidebar from "@/components/portal/PortalSidebar";
import WalletPanel from "@/components/portal/WalletPanel";
import CustomerProfile from "@/components/portal/CustomerProfile";
import MerchantWorkspace from "@/components/portal/MerchantWorkspace";
import AdminWorkspace from "@/components/portal/admin/AdminWorkspace";
import { adminNavigation, type AdminSection } from "@/components/portal/admin/navigation";
import { categories, getSidebarItems } from "@/components/portal/data";
import type { PortalRole, ProductId } from "@/components/portal/types";
import { portalRoutes } from "@/components/portal/routes";
import { refreshSession, signOut as signOutRequest } from "@/services/auth";

export default function PortalShell({ initialRole = "user", initialTab = 0, initialSection = "overview" }: { initialRole?: PortalRole; initialTab?: number; initialSection?: AdminSection }) {
  const router = useRouter();
  const { language, setLanguage } = useI18n();
  const [role, setRole] = useState<PortalRole>(initialRole);
  const [accountName, setAccountName] = useState("");
  const [accountId, setAccountId] = useState("");
  const [accountEmail, setAccountEmail] = useState("");
  const [authReady, setAuthReady] = useState(false);
  const [authError, setAuthError] = useState(false);
  const activeNavIndex = initialTab;
  const [category, setCategory] = useState<string>("all");
  const [saved, setSaved] = useState<ProductId[]>([]);
  const [cart, setCart] = useState(0);
  const t = portalMessages[language];
  const adminCopy = adminMessages[language];
  const adminLinks = adminNavigation(activeNavIndex < 3 ? activeNavIndex : 0, adminCopy);
  const navigation = t.nav[role];
  const activeNav = navigation[activeNavIndex];
  const isCustomer = role === "user";
  const walletPage = (isCustomer && activeNavIndex === 3) || (role === "merchant" && activeNavIndex === 4);
  const sidebarItems = useMemo(() => role === "admin" ? adminNavigation(activeNavIndex, adminCopy).map((item) => item.label) : getSidebarItems(role, activeNavIndex, t), [role, activeNavIndex, t, adminCopy]);
  const money = (amount: number) => new Intl.NumberFormat(language === "vi" ? "vi-VN" : "en-US", { style: "currency", currency: "VND", maximumFractionDigits: 0 }).format(amount);

  useEffect(() => {
    let active = true;
    refreshSession()
      .then(({ response, session }) => {
        if (!response.ok) {
          if (active) router.replace("/auth");
          return;
        }
        const actualRole = session?.user?.role?.toUpperCase();
        const destination = actualRole === "MERCHANT" ? "/merchant" : actualRole === "ADMIN" ? "/admin" : actualRole === "CUSTOMER" ? "/customer" : null;
        if (!destination) { if (active) setAuthError(true); return; }
        const actualPortalRole: PortalRole = actualRole === "MERCHANT" ? "merchant" : actualRole === "ADMIN" ? "admin" : "user";
        if (active && actualPortalRole !== initialRole) { router.replace(destination); return; }
        if (active) { setRole(actualPortalRole); setAccountId(session?.user?.id ?? ""); setAccountName(session?.user?.name ?? ""); setAccountEmail(session?.user?.email ?? ""); setAuthReady(true); }
      })
      .catch(() => { if (active) setAuthError(true); });
    return () => { active = false; };
  }, [initialRole, router]);

  async function signOut() {
    await signOutRequest().catch(() => undefined);
    router.replace("/");
    router.refresh();
  }

  if (authError) return <main className="portal-auth-error"><span className="brand-mark">s</span><h1>{language === "vi" ? "Chưa xác minh được phiên đăng nhập" : "Could not verify your session"}</h1><p>{language === "vi" ? "Kiểm tra kết nối Supay rồi thử lại." : "Check your Supay connection and try again."}</p><button onClick={() => window.location.reload()}>{language === "vi" ? "Thử lại" : "Try again"}</button><Link href="/auth">{language === "vi" ? "Đăng nhập" : "Sign in"}</Link></main>;
  if (!authReady) return <main className="portal-auth-loading"><span className="brand-mark">s</span><p>{language === "vi" ? "Đang xác minh tài khoản…" : "Verifying your account…"}</p></main>;

  const selectedSidebarIndex = role === "admin" ? adminLinks.findIndex((item) => item.id === initialSection) : isCustomer ? category === "all" ? 0 : categories.indexOf(category as (typeof categories)[number]) + 1 : 0;
  return <div className={role === "admin" ? "app-shell admin-shell" : "app-shell"}>
    <PortalHeader role={role} copy={t} language={language} setLanguage={setLanguage} activeNavIndex={activeNavIndex} navHrefs={portalRoutes[role]} cart={cart} accountName={accountName} onSignOut={signOut} />
    <div className="page-frame">
      <PortalSidebar role={role} title={activeNav} items={sidebarItems} hrefs={role === "admin" ? adminLinks.map((item) => item.href) : undefined} selectedIndex={selectedSidebarIndex} copy={t} onSelect={(index) => setCategory(index === 0 ? "all" : categories[index - 1])} />
      <main className="main-content"><div className="breadcrumb"><span>Supay</span><span>/</span><strong>{activeNav}</strong>{role === "admin" && initialSection !== "overview" && <><span>/</span><strong>{adminCopy.sections[initialSection]}</strong></>}</div>
        {role === "admin" ? <AdminWorkspace tab={activeNavIndex} section={initialSection} language={language} portalCopy={t} account={{ id: accountId, name: accountName, email: accountEmail, role: "ADMIN" }} />
          : walletPage ? <WalletPanel copy={t} money={money} merchant={role === "merchant"} />
          : isCustomer && activeNavIndex === 2 ? <CustomerProfile copy={t} name={accountName} email={accountEmail} />
          : isCustomer ? <CustomerStorefront copy={t} language={language} category={category} setCategory={setCategory} saved={saved} setSaved={setSaved} onAddToCart={() => setCart((count) => count + 1)} onNavigateToShop={() => router.push("/customer/shop")} showHero={activeNavIndex === 0} />
          : role === "merchant" && activeNavIndex >= 1 ? <MerchantWorkspace tab={activeNavIndex as 1 | 2 | 3} copy={t} money={money} />
          : <Dashboard role={role} copy={t} language={language} money={money} activeNavIndex={activeNavIndex} currentUserId={accountId} />}
      </main>
    </div>
    {role === "admin" ? <footer className="admin-footer"><span>© 2026 Supay</span><span>{adminCopy.groups.system.kicker}</span></footer> : <PortalFooter copy={t} language={language} />}
  </div>;
}
