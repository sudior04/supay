"use client";
import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { useI18n } from "@/i18n/provider";
import { apiFetch, clearAccessToken } from "@/services/auth";

export default function PINSettings() {
  const { language } = useI18n(); const vi = language === "vi"; const router = useRouter();
  const [configured, setConfigured] = useState<boolean | null>(null); const [busy, setBusy] = useState(false); const [message, setMessage] = useState("");
  useEffect(() => { apiFetch("/api/v1/users/me/pin").then(async response => { if (!response.ok) throw new Error("Could not load PIN settings."); const result = await response.json(); setConfigured(result.pin_configured); }).catch(() => setMessage(vi ? "Chưa tải được thiết lập PIN." : "Could not load PIN settings.")); }, [vi]);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); const form = new FormData(event.currentTarget);
    const currentPIN = String(form.get("current_pin") ?? ""); const newPIN = String(form.get("new_pin") ?? "");
    if (!/^[0-9]{6}$/.test(newPIN) || (configured && !/^[0-9]{6}$/.test(currentPIN)) || newPIN !== form.get("confirm_pin")) { setMessage(vi ? "Nhập PIN 6 chữ số và xác nhận trùng khớp." : "Enter matching 6-digit PINs."); return; }
    setBusy(true); setMessage("");
    try {
      const password = String(form.get("password"));
      const response = await apiFetch("/api/v1/users/me/pin", { method: configured ? "PUT" : "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(configured ? { current_password: password, current_pin: currentPIN, new_pin: newPIN } : { current_password: password, pin: newPIN }) });
      if (response.ok) {
        if (configured) { clearAccessToken(); router.replace("/auth?pinChanged=1"); return; }
        setConfigured(true); setMessage(vi ? "Đã thiết lập PIN xác nhận thao tác." : "Your action confirmation PIN is configured."); return;
      }
      const result = await response.json(); setMessage(result.error?.message ?? "Could not save PIN.");
    } catch { setMessage(vi ? "Không kết nối được dịch vụ tài khoản." : "Could not connect to the account service."); } finally { setBusy(false); }
  }
  return <section className="password-panel"><div className="password-panel-heading"><h2>{vi ? "Mã PIN xác nhận thao tác" : "Action confirmation PIN"}</h2><p>{vi ? "PIN chỉ được yêu cầu khi thêm, sửa, xóa sản phẩm hoặc đơn hàng. Đăng nhập chỉ dùng mật khẩu." : "PIN confirms product/order changes. Sign-in only needs your password."}</p></div>
    <form className="password-form" key={String(configured)} onSubmit={submit}>
      <label>{vi ? "Mật khẩu hiện tại" : "Current password"}<input name="password" type="password" autoComplete="current-password" maxLength={72} required /></label>
      {configured && <label>{vi ? "PIN hiện tại" : "Current PIN"}<input name="current_pin" type="password" inputMode="numeric" pattern="[0-9]{6}" minLength={6} maxLength={6} autoComplete="off" required /></label>}
      {[["new_pin", vi ? "PIN mới" : "New PIN"], ["confirm_pin", vi ? "Xác nhận PIN mới" : "Confirm new PIN"]].map(([name, label]) => <label key={name}>{label}<input name={name} type="password" inputMode="numeric" pattern="[0-9]{6}" minLength={6} maxLength={6} autoComplete="off" required /></label>)}
      {message && <p role="status" className="password-message">{message}</p>}<button disabled={busy || configured === null}>{busy ? "…" : configured ? (vi ? "Đổi PIN" : "Change PIN") : (vi ? "Thiết lập PIN" : "Set up PIN")}</button>
    </form>
  </section>;
}
