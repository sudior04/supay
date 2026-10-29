import type { portalMessages } from "@/i18n/portal";

export type PortalRole = "user" | "merchant" | "admin";
export type PortalCopy = (typeof portalMessages)[keyof typeof portalMessages];
export type ProductId = "tee" | "tote" | "bottle" | "candle" | "audio" | "skincare";
export type CategoryId = "fashion" | "daily" | "home" | "electronics" | "beauty" | "sports";
