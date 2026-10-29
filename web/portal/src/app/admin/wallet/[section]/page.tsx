import { notFound } from "next/navigation";
import PortalShell from "../../../portal-shell";
import { adminSections } from "@/components/portal/admin/navigation";

export function generateStaticParams() {
  return adminSections[2].filter((section) => section !== "overview").map((section) => ({ section }));
}

export default async function AdminWalletSectionPage({ params }: { params: Promise<{ section: string }> }) {
  const { section } = await params;
  const selected = adminSections[2].find((id) => id === section && id !== "overview");
  if (!selected) notFound();
  return <PortalShell initialRole="admin" initialTab={2} initialSection={selected} />;
}
