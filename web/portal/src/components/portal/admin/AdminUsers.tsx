import type { AdminCopy } from "@/i18n/admin";
import type { PortalCopy } from "../types";
import AdminPasswordReset from "../AdminPasswordReset";
import { AdminDataTable } from "./AdminPrimitives";

export default function AdminUsers({ copy: t, portalCopy, currentUserId }: { copy: AdminCopy; portalCopy: PortalCopy; currentUserId: string }) {
  return <><AdminDataTable copy={t} portalCopy={portalCopy} title={t.sections.users} columns={["name", "email", "role", "status", "created", "action"]} users /><AdminPasswordReset copy={portalCopy} currentUserId={currentUserId} /></>;
}
