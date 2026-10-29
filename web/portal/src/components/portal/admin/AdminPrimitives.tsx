"use client";

import Link from "next/link";
import { useState } from "react";
import type { AdminCopy } from "@/i18n/admin";
import type { PortalCopy } from "../types";
import { adminSectionHref, type AdminSection } from "./navigation";

export function AdminHeading({ title, description, kicker }: { title: string; description: string; kicker: string }) {
  return <div className="admin-heading"><span className="section-kicker">{kicker}</span><h1>{title}</h1><p>{description}</p></div>;
}

export function AdminMetrics({ copy, labels }: { copy: AdminCopy; labels: readonly string[] }) {
  return <div className="admin-metrics">{labels.map((label) => <article key={label}><span>{label}</span><strong>—</strong><small>{copy.pendingData}</small></article>)}</div>;
}

export function AdminDirectory({ copy: t, tab, sections }: { copy: AdminCopy; tab: number; sections: readonly AdminSection[] }) {
  return <section className="admin-directory"><h2>{t.directory}</h2><div>{sections.filter((section) => section !== "overview").map((section, index) => <Link key={section} href={adminSectionHref(tab, section)} className="admin-directory-card"><span className="admin-card-number">{String(index + 1).padStart(2, "0")}</span><h3>{t.sections[section]}</h3><p>{t.descriptions[section as keyof AdminCopy["descriptions"]]}</p><span className="admin-card-link">{t.open} <span aria-hidden="true">→</span></span></Link>)}</div></section>;
}

export function AdminDataTable({ copy: t, portalCopy, title, columns, users = false, statuses = ["ACTIVE", "DISABLED"] }: { copy: AdminCopy; portalCopy: PortalCopy; title: string; columns: readonly (keyof AdminCopy["columns"])[]; users?: boolean; statuses?: readonly (keyof AdminCopy["states"])[] }) {
  const [query, setQuery] = useState("");
  const [role, setRole] = useState("");
  const [status, setStatus] = useState("");
  const filtered = Boolean(query.trim() || role || status);
  return <section className="admin-table-panel">
    <div className="admin-table-heading"><h2>{title}</h2><span className="admin-data-state">{t.pendingData}</span></div>
    <div className="admin-table-toolbar"><label className="admin-table-search"><span aria-hidden="true">⌕</span><input type="search" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t.search} aria-label={t.search} /></label>{users && <select aria-label={t.columns.role} value={role} onChange={(event) => setRole(event.target.value)}><option value="">{t.allRoles}</option><option value="CUSTOMER">{portalCopy.roles.user}</option><option value="MERCHANT">{portalCopy.roles.merchant}</option><option value="ADMIN">{portalCopy.roles.admin}</option></select>}{columns.includes("status") && <select aria-label={t.columns.status} value={status} onChange={(event) => setStatus(event.target.value)}><option value="">{t.allStatuses}</option>{statuses.map((value) => <option key={value} value={value}>{t.states[value]}</option>)}</select>}</div>
    <div className="admin-table-scroll"><table><thead><tr>{columns.map((column) => <th key={column} scope="col">{t.columns[column]}</th>)}</tr></thead><tbody><tr><td colSpan={columns.length}><div className="admin-empty"><span aria-hidden="true">▤</span><h3>{filtered ? t.noResults : t.emptyTitle}</h3><p>{filtered ? t.noResultsDescription : t.emptyDescription}</p></div></td></tr></tbody></table></div>
  </section>;
}
