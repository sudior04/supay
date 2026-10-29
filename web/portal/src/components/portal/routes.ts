import type { PortalRole } from "./types";

export const portalRoutes: Record<PortalRole, readonly string[]> = {
  user: ["/customer", "/customer/shop", "/customer/profile", "/customer/wallet"],
  merchant: ["/merchant", "/merchant/products", "/merchant/orders", "/merchant/store", "/merchant/finance"],
  admin: ["/admin", "/admin/commerce", "/admin/wallet"],
};
