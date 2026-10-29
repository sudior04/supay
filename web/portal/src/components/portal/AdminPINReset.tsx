"use client";
import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { useI18n } from "@/i18n/provider";
import { apiFetch, clearAccessToken } from "@/services/auth";

export default function AdminPINReset({ currentUserId }: { currentUserId: string }) {
  const { language } = useI18n(); const vi = language === "vi"; const router = useRouter();
  const [busy, setBusy] = useState(false); const [message, setMessage] = useState("");
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); const id = String(new FormData(event.currentTarget).get("user_id") ?? "").trim();
    if (!window.confirm(vi ? "Thu hồi các phiên đăng nhập và yêu cầu user thiết lập PIN mới?" : "Revoke sessions and require this user to set up a new PIN?")) return;
    setBusy(true); setMessage("");
    try {
      const response = await apiFetch(`/api/v1/users/${encodeURIComponent(id)}/pin/reset`, { method: "POST" });
      if (!response.ok) { const result = await response.json(); setMessage(result.error?.message ?? "Could not reset PIN."); return; }
      if (id.toLowerCase() === currentUserId.toLowerCase()) { clearAccessToken(); router.replace("/auth"); return; }
      setMessage(vi ? "Đã xóa PIN. User có thể đăng nhập bình thường và thiết lập lại PIN trong bảo mật tài khoản." : "PIN cleared. The user can sign in normally and set a new PIN in account security.");
    } catch { setMessage(vi ? "Không kết nối được dịch vụ." : "Could not connect to the service."); } finally { setBusy(false); }
  }
  return <section className="password-panel"><div className="password-panel-heading"><h2>{vi ? "Khôi phục mã PIN" : "Reset user PIN"}</h2><p>{vi ? "Chỉ dành cho ADMIN. User vẫn cần mật khẩu để tạo PIN mới." : "ADMIN only. The user still needs their password to enroll a new PIN."}</p></div><form className="password-form" onSubmit={submit}><label>User ID<input name="user_id" pattern="[0-9a-fA-F-]{36}" required /></label>{message && <p role="status">{message}</p>}<button disabled={busy}>{busy ? "…" : vi ? "Reset PIN" : "Reset PIN"}</button></form></section>;
}
