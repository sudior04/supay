export const featuredItems = [
  { price: 249000, icon: "👕", tone: "mint" },
  { price: 189000, icon: "👜", tone: "peach" },
  { price: 329000, icon: "🧋", tone: "blue" },
  { price: 159000, icon: "🕯️", tone: "lilac" },
] as const;

export const serviceEndpoints = ["/api/v1/users/status", "/api/v1/products/status", "/api/v1/wallets/status"] as const;
