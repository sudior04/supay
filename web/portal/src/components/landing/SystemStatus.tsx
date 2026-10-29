import type { LandingCopy, ServiceState } from "./types";

export default function SystemStatus({ copy: t, services }: { copy: LandingCopy; services: ServiceState[] }) {
  const statusText = (status: ServiceState) => status === "online" ? t.online : status === "offline" ? t.offline : t.checking;
  return <section className="system-status"><div className="status-heading"><span className="status-live-dot"/><div><span className="landing-kicker">{t.system}</span><h2>{t.systemTitle}</h2></div></div><div className="status-services">{t.statusLabels.map((label, index) => <div className="status-service" key={label}><span>{label}</span><strong className={`service-state ${services[index]}`}>{statusText(services[index])}</strong></div>)}</div></section>;
}
