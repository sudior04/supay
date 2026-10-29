"use client";

import { useEffect, useState } from "react";
import type { AdminCopy } from "@/i18n/admin";
import { AdminDirectory, AdminMetrics } from "./AdminPrimitives";
import { adminSections } from "./navigation";

export default function AdminOverview({ copy: t }: { copy: AdminCopy }) {
  const [services, setServices] = useState<("checking" | "online" | "offline")[]>(["checking", "checking", "checking"]);
  useEffect(() => {
    let active = true;
    Promise.all(["users", "products", "wallets"].map(async (service) => {
      try { return (await fetch(`/api/v1/${service}/status`, { cache: "no-store" })).ok ? "online" as const : "offline" as const; }
      catch { return "offline" as const; }
    })).then((result) => { if (active) setServices(result); });
    return () => { active = false; };
  }, []);
  return <><AdminMetrics copy={t} labels={[t.metrics.users, t.metrics.merchants, t.metrics.orders, t.metrics.wallets]} /><AdminDirectory copy={t} tab={0} sections={adminSections[0]} /><section className="admin-services"><div><h2>{t.serviceTitle}</h2><p>{t.serviceDescription}</p></div><ul>{services.map((state, index) => <li key={t.services[index]}><span>{t.services[index]}</span><strong className={`admin-service-state ${state}`}>{t[state]}</strong></li>)}</ul></section></>;
}
