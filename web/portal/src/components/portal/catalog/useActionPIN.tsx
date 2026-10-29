"use client";
import Link from "next/link";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";

export function useActionPIN(vi: boolean, settingsHref: string) {
  const [open, setOpen] = useState(false); const [pin, setPIN] = useState("");
  const pending = useRef<((pin: string | null) => void) | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);
  const finish = useCallback((value: string | null) => { const resolve = pending.current; pending.current = null; setPIN(""); setOpen(false); resolve?.(value); }, []);
  useEffect(() => () => { pending.current?.(null); pending.current = null; }, []);
  useEffect(() => { const element = dialogRef.current; if (open && element && !element.open) element.showModal(); if (!open && element?.open) element.close(); }, [open]);
  const requestPIN = useCallback(() => { if (pending.current) return Promise.resolve(null); return new Promise<string | null>(resolve => { pending.current = resolve; setPIN(""); setOpen(true); }); }, []);
  function submit(event: FormEvent<HTMLFormElement>) { event.preventDefault(); if (/^[0-9]{6}$/.test(pin)) finish(pin); }
  const dialog = <dialog ref={dialogRef} className="catalog-pin-dialog" onCancel={event => { event.preventDefault(); finish(null); }} aria-label={vi ? "Xác nhận bằng PIN" : "Confirm with PIN"}>
    <form className="catalog-form" onSubmit={submit}><h3>{vi ? "Xác nhận thao tác" : "Confirm this change"}</h3><p>{vi ? "Nhập PIN 6 số để xác nhận thay đổi sản phẩm này." : "Enter your 6-digit PIN to confirm this product change."}</p>
      <label>PIN<input type="password" inputMode="numeric" pattern="[0-9]{6}" minLength={6} maxLength={6} autoComplete="off" autoFocus value={pin} onChange={event => setPIN(event.target.value)} required /></label>
      <div className="catalog-actions"><button type="submit">{vi ? "Xác nhận" : "Confirm"}</button><button type="button" className="secondary" onClick={() => finish(null)}>{vi ? "Hủy" : "Cancel"}</button></div>
      <Link href={settingsHref}>{vi ? "Chưa có PIN? Mở bảo mật tài khoản để thiết lập." : "No PIN yet? Open account security to set it up."}</Link>
    </form>
  </dialog>;
  return { requestPIN, dialog };
}
