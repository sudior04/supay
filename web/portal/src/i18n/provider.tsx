"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";

export type Language = "vi" | "en";
type LanguageContextValue = { language: Language; setLanguage: (language: Language) => void };
const LanguageContext = createContext<LanguageContextValue | null>(null);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [language, setCurrentLanguage] = useState<Language>("vi");

  useEffect(() => {
    const saved = window.localStorage.getItem("supay-language");
    if (saved === "vi" || saved === "en") setCurrentLanguage(saved);
  }, []);

  function setLanguage(next: Language) {
    setCurrentLanguage(next);
    document.documentElement.lang = next;
    try { window.localStorage.setItem("supay-language", next); } catch { /* Language switching still works when storage is blocked. */ }
  }

  useEffect(() => { document.documentElement.lang = language; }, [language]);

  return <LanguageContext.Provider value={{ language, setLanguage }}>{children}</LanguageContext.Provider>;
}

export function useI18n() {
  const context = useContext(LanguageContext);
  if (!context) throw new Error("useI18n must be used inside I18nProvider");
  return context;
}
