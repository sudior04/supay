import type { landingMessages } from "@/i18n/landing";

export type LandingCopy = (typeof landingMessages)[keyof typeof landingMessages];
export type ServiceState = "loading" | "online" | "offline";
