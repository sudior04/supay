import type { authMessages } from "@/i18n/auth";

export type AuthCopy = (typeof authMessages)[keyof typeof authMessages];
export type AuthMode = "login" | "register";
