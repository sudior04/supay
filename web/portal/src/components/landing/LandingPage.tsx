"use client";

import { landingMessages } from "@/i18n/landing";
import { useI18n } from "@/i18n/provider";
import FeaturedProducts from "./FeaturedProducts";
import LandingFooter from "./LandingFooter";
import LandingHeader from "./LandingHeader";
import LandingHero from "./LandingHero";
import MerchantCallout from "./MerchantCallout";
import PlatformOverview from "./PlatformOverview";
import SystemStatus from "./SystemStatus";
import { useServiceStatus } from "./useServiceStatus";

export default function LandingPage() {
  const { language, setLanguage } = useI18n();
  const copy = landingMessages[language];
  const services = useServiceStatus();
  return <div className="landing-page">
    <LandingHeader copy={copy} language={language} setLanguage={setLanguage} />
    <main><LandingHero copy={copy} /><PlatformOverview copy={copy} /><FeaturedProducts copy={copy} language={language} /><MerchantCallout copy={copy} /><SystemStatus copy={copy} services={services} /></main>
    <LandingFooter copy={copy} />
  </div>;
}
