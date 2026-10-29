import type { CategoryId, PortalCopy, PortalRole } from "./types";

export const categories: readonly CategoryId[] = ["fashion", "daily", "home", "electronics", "beauty", "sports"];

export const products = [
  { id: "tee", category: "fashion", price: 249000, old: 320000, icon: "👕", tone: "mint", sold: 128 },
  { id: "tote", category: "fashion", price: 189000, old: 240000, icon: "👜", tone: "peach", sold: 86 },
  { id: "bottle", category: "daily", price: 329000, old: 390000, icon: "🧋", tone: "blue", sold: 214 },
  { id: "candle", category: "home", price: 159000, old: 199000, icon: "🕯️", tone: "lilac", sold: 63 },
  { id: "audio", category: "electronics", price: 790000, old: 990000, icon: "🎧", tone: "blue", sold: 142 },
  { id: "skincare", category: "beauty", price: 425000, old: 520000, icon: "🧴", tone: "peach", sold: 97 },
] as const;

export type PortalProduct = (typeof products)[number];

export const productNames = {
  vi: { tee: "Áo thun cotton Essential", tote: "Túi tote Everyday", bottle: "Bình giữ nhiệt 500ml", candle: "Nến thơm thư giãn", audio: "Tai nghe không dây Air", skincare: "Bộ chăm sóc da dịu nhẹ" },
  en: { tee: "Essential cotton tee", tote: "Everyday canvas tote", bottle: "500ml insulated bottle", candle: "Calming scented candle", audio: "Air wireless headphones", skincare: "Gentle daily skincare set" },
} as const;

export function getSidebarItems(role: Exclude<PortalRole, "admin">, activeNavIndex: number, t: PortalCopy) {
  if (role === "user") return [t.all, ...categories.map((category) => t.categories[category])];
  if (role === "merchant") {
    if (activeNavIndex === 1) return [t.catalog, t.addProduct, t.productCategories];
    if (activeNavIndex === 2) return [t.allOrders, t.pending, t.shippingOrders, t.completed];
    if (activeNavIndex === 3) return [t.storeInfo, t.cashflow, t.storeWallet];
    if (activeNavIndex === 4) return [t.walletOverview, t.transactions, t.reconciliation, t.withdraw];
    return [t.overviewNav, t.monthlyRevenue, t.activities];
  }
  return [];
}
