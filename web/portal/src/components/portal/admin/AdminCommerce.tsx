import type { AdminCopy } from "@/i18n/admin";
import type { PortalCopy } from "../types";
import { AdminDataTable, AdminDirectory, AdminMetrics } from "./AdminPrimitives";
import { adminSections, type AdminSection } from "./navigation";
import CategoryManager from "../catalog/CategoryManager";
import ProductManager from "../catalog/ProductManager";

const columns = {
  stores: ["store", "merchant", "email", "status", "created"],
  products: ["product", "store", "category", "price", "stock", "status"],
  orders: ["order", "buyer", "store", "amount", "payment", "status", "created"],
  categories: ["category", "parent", "productCount", "status"],
} as const;

export default function AdminCommerce({ copy: t, portalCopy, section }: { copy: AdminCopy; portalCopy: PortalCopy; section: AdminSection }) {
  if (section === "categories") return <CategoryManager />;
  if (section === "products") return <ProductManager admin />;
  if (section === "overview") return <><AdminMetrics copy={t} labels={[t.metrics.stores, t.metrics.products, t.metrics.orders, t.metrics.volume]} /><AdminDirectory copy={t} tab={1} sections={adminSections[1]} /></>;
  const selected = section as keyof typeof columns;
  return <AdminDataTable key={section} copy={t} portalCopy={portalCopy} title={t.sections[selected]} columns={columns[selected]} statuses={selected === "orders" ? ["PENDING_PAYMENT", "PAID", "CANCELLED", "PAYMENT_FAILED", "REFUNDED"] : undefined} />;
}
