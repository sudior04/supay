import type { AdminCopy } from "@/i18n/admin";

export type AdminSection = keyof AdminCopy["sections"];
export const adminSections: readonly (readonly AdminSection[])[] = [
  ["overview", "users", "activity", "security"],
  ["overview", "stores", "products", "orders", "categories"],
  ["overview", "wallets", "transactions", "reconciliation", "fees"],
];
const bases = ["/admin", "/admin/commerce", "/admin/wallet"];

export function adminSectionHref(tab: number, section: AdminSection) {
  return section === "overview" ? bases[tab] : `${bases[tab]}/${section}`;
}

export function adminNavigation(tab: number, copy: AdminCopy) {
  return adminSections[tab].map((id) => ({ id, label: copy.sections[id], href: adminSectionHref(tab, id) }));
}
