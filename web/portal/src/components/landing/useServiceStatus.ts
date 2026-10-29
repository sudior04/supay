"use client";

import { useEffect, useState } from "react";
import { serviceEndpoints } from "./data";
import type { ServiceState } from "./types";

export function useServiceStatus() {
  const [services, setServices] = useState<ServiceState[]>(["loading", "loading", "loading"]);
  useEffect(() => {
    let active = true;
    Promise.all(serviceEndpoints.map(async (endpoint) => {
      try {
        const response = await fetch(endpoint, { cache: "no-store" });
        return response.ok ? "online" as const : "offline" as const;
      } catch {
        return "offline" as const;
      }
    })).then((result) => { if (active) setServices(result); });
    return () => { active = false; };
  }, []);
  return services;
}
