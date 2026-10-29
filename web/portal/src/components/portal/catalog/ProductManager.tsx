"use client";
import { useEffect, useState, type FormEvent } from "react";
import { useI18n } from "@/i18n/provider";
import { allCategories, catalogRequest, parseAttributes, type Category, type Page, type Product, type Variant } from "@/services/catalog";
import "./catalog.css";
import { useActionPIN } from "./useActionPIN";

function VariantFields({ prefix, variant, quantity, vi }: { prefix: string; variant?: Variant; quantity?: boolean; vi: boolean }) {
  return <fieldset><legend>{vi ? "Biến thể" : "Variant"}</legend>
    <label>SKU<input name={`${prefix}sku`} defaultValue={variant?.sku} maxLength={80} pattern="[A-Za-z0-9][A-Za-z0-9_.-]*" required /></label>
    <label>{vi ? "Tên biến thể" : "Variant name"}<input name={`${prefix}name`} defaultValue={variant?.name} maxLength={120} required /></label>
    <label>{vi ? "Giá (VND)" : "Price (VND)"}<input name={`${prefix}price_minor`} type="number" min={1} step={1} max={Number.MAX_SAFE_INTEGER} defaultValue={variant?.price_minor} required /></label>
    {quantity && <label>{vi ? "Tồn kho ban đầu" : "Initial stock"}<input name={`${prefix}quantity`} type="number" min={0} step={1} max={Number.MAX_SAFE_INTEGER} defaultValue={0} required /></label>}
    <label>{vi ? "Thuộc tính (mỗi dòng tên=giá trị)" : "Attributes (one name=value per line)"}<textarea name={`${prefix}attributes`} placeholder={"color=red\nsize=M"} defaultValue={Object.entries(variant?.attributes ?? {}).map(([key, value]) => `${key}=${value}`).join("\n")} /></label>
  </fieldset>;
}

function variantBody(form: FormData, prefix: string, quantity: boolean) {
  return { sku: String(form.get(`${prefix}sku`)), name: String(form.get(`${prefix}name`)), price_minor: Number(form.get(`${prefix}price_minor`)), attributes: parseAttributes(String(form.get(`${prefix}attributes`) ?? "")), ...(quantity ? { quantity: Number(form.get(`${prefix}quantity`)) } : {}) };
}

export default function ProductManager({ admin = false }: { admin?: boolean }) {
  const { language } = useI18n(); const vi = language === "vi";
  const { requestPIN, dialog } = useActionPIN(vi, admin ? "/admin/security" : "/merchant");
  const [items, setItems] = useState<Product[]>([]); const [categories, setCategories] = useState<Category[]>([]);
  const [selected, setSelected] = useState<Product | null>(null); const [creating, setCreating] = useState(false); const [variantCount, setVariantCount] = useState(1);
  const [busy, setBusy] = useState(false); const [message, setMessage] = useState(""); const [offset, setOffset] = useState(0); const [more, setMore] = useState(false); const [revision, setRevision] = useState(0);
  async function load(pageOffset = offset) { const page = await catalogRequest<Page<"products", Product>>(`products?limit=20&offset=${pageOffset}${admin ? "" : "&mine=true"}`); setItems(page.products); setMore(page.pagination.has_more); }
  useEffect(() => { Promise.all([load(offset), allCategories().then(setCategories)]).catch(error => setMessage(String(error.message))); }, [offset, admin]);
  async function reloadEditor(id: string) { await load(); const result = await catalogRequest<{ product: Product }>(`products/${id}`); setSelected(result.product); setRevision(value => value + 1); }
  async function saveProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); const form = new FormData(event.currentTarget); const pin = await requestPIN(); if (pin === null) return; setBusy(true); setMessage("");
    try {
      const fields = { category_id: String(form.get("category_id")), name: String(form.get("name")), description: String(form.get("description")), status: String(form.get("status")) };
      const result = await catalogRequest<{ product: Product }>(creating ? "products" : `products/${selected!.id}`, creating ? "POST" : "PUT", { ...fields, ...(creating ? { variants: Array.from({ length: variantCount }, (_, i) => variantBody(form, `v${i}_`, true)) } : {}) }, pin);
      setCreating(false); await reloadEditor(result.product.id); setMessage(vi ? "Đã lưu sản phẩm." : "Product saved.");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Request failed"); } finally { setBusy(false); }
  }
  async function removeProduct(product: Product) {
    if (!window.confirm(`${vi ? "Xóa sản phẩm" : "Delete product"} ${product.name}?`)) return;
    const pin = await requestPIN(); if (pin === null) return;
    setBusy(true); try { await catalogRequest(`products/${product.id}`, "DELETE", undefined, pin); if (selected?.id === product.id) setSelected(null); await load(); setMessage(vi ? "Đã xóa sản phẩm." : "Product deleted."); } catch (error) { setMessage(error instanceof Error ? error.message : "Request failed"); } finally { setBusy(false); }
  }
  async function mutateVariant(path: string, method: string, body?: object) {
    if (!selected) return; const pin = await requestPIN(); if (pin === null) return; setBusy(true); setMessage("");
    try { await catalogRequest(path, method, body, pin); await reloadEditor(selected.id); setMessage(vi ? "Đã cập nhật biến thể/tồn kho." : "Variant/inventory updated."); }
    catch (error) { setMessage(error instanceof Error ? error.message : "Request failed"); await reloadEditor(selected.id).catch(() => undefined); }
    finally { setBusy(false); }
  }
  function variantSubmit(event: FormEvent<HTMLFormElement>, variant?: Variant) {
    event.preventDefault(); if (!selected) return;
    try { const body = variantBody(new FormData(event.currentTarget), "", !variant); void mutateVariant(`products/${selected.id}/variants${variant ? `/${variant.id}` : ""}`, variant ? "PUT" : "POST", body); } catch (error) { setMessage(error instanceof Error ? error.message : "Invalid attributes"); }
  }
  return <section className="catalog-manager">{dialog}<div className="catalog-actions"><h2>{vi ? "Sản phẩm & tồn kho" : "Products & inventory"}</h2><button disabled={busy || !categories.length} onClick={() => { setCreating(true); setSelected(null); setVariantCount(1); setMessage(""); }}>{vi ? "Thêm sản phẩm" : "Add product"}</button></div>
    {!categories.length && <p>{vi ? "Cần ADMIN tạo danh mục trước khi thêm sản phẩm." : "Ask an administrator to create categories first."}</p>}
    {message && <p className="catalog-message" role="status">{message}</p>}
    <div className="workspace-table-wrap"><table className="workspace-table"><thead><tr><th>{vi ? "Tên" : "Name"}</th><th>{vi ? "Danh mục" : "Category"}</th><th>VND</th><th>{vi ? "Khả dụng" : "Available"}</th><th>{vi ? "Trạng thái" : "Status"}</th><th /></tr></thead><tbody>{items.map(product => <tr key={product.id}><td>{product.name}</td><td>{product.category_name}</td><td>{product.variants.length ? Math.min(...product.variants.map(v => v.price_minor)).toLocaleString() : "—"}</td><td>{product.variants.reduce((sum, v) => sum + v.inventory.available, 0)}</td><td>{product.status}</td><td className="catalog-actions"><button disabled={busy} onClick={() => { setSelected(product); setCreating(false); setRevision(v => v + 1); }}>{vi ? "Sửa" : "Edit"}</button><button disabled={busy} className="danger" onClick={() => removeProduct(product)}>{vi ? "Xóa" : "Delete"}</button></td></tr>)}</tbody></table></div>
    <div className="catalog-actions"><button disabled={offset === 0 || busy} onClick={() => setOffset(Math.max(0, offset - 20))}>←</button><span>{offset + 1}–{offset + items.length}</span><button disabled={!more || busy} onClick={() => setOffset(offset + 20)}>→</button></div>
    {(creating || selected) && <form key={`${selected?.id ?? "new"}-${revision}`} className="catalog-form" onSubmit={saveProduct}>
      <h3>{creating ? (vi ? "Sản phẩm mới" : "New product") : selected?.name}</h3>
      <div className="catalog-inline"><label>{vi ? "Tên sản phẩm" : "Product name"}<input name="name" defaultValue={selected?.name} maxLength={150} required /></label><label>{vi ? "Danh mục" : "Category"}<select name="category_id" defaultValue={selected?.category_id} required>{categories.map(category => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label><label>{vi ? "Trạng thái" : "Status"}<select name="status" defaultValue={selected?.status ?? "DRAFT"}><option value="DRAFT">DRAFT</option><option value="ACTIVE">ACTIVE</option></select></label></div>
      <label>{vi ? "Mô tả" : "Description"}<textarea name="description" defaultValue={selected?.description} maxLength={5000} /></label>
      {creating && <>{Array.from({ length: variantCount }, (_, i) => <VariantFields key={i} prefix={`v${i}_`} quantity vi={vi} />)}<div className="catalog-actions"><button type="button" className="secondary" disabled={variantCount >= 100} onClick={() => setVariantCount(v => v + 1)}>+ {vi ? "Biến thể" : "Variant"}</button><button type="button" className="secondary" disabled={variantCount <= 1} onClick={() => setVariantCount(v => v - 1)}>−</button></div></>}
      <div className="catalog-actions"><button disabled={busy}>{vi ? "Lưu sản phẩm" : "Save product"}</button><button type="button" className="secondary" disabled={busy} onClick={() => { setSelected(null); setCreating(false); }}>{vi ? "Đóng" : "Close"}</button></div>
    </form>}
    {selected && !creating && <>{selected.variants.map(variant => <article className="catalog-variant" key={`${variant.id}-${revision}`}>
      <form className="catalog-form" onSubmit={event => variantSubmit(event, variant)}><VariantFields prefix="" variant={variant} vi={vi} /><div className="catalog-actions"><button disabled={busy}>{vi ? "Lưu biến thể" : "Save variant"}</button><button type="button" className="danger" disabled={busy} onClick={() => { if (window.confirm(vi ? "Xóa biến thể này?" : "Delete this variant?")) void mutateVariant(`products/${selected.id}/variants/${variant.id}`, "DELETE"); }}>{vi ? "Xóa biến thể" : "Delete variant"}</button></div></form>
      <form className="catalog-form" onSubmit={event => { event.preventDefault(); const form = new FormData(event.currentTarget); void mutateVariant(`products/${selected.id}/variants/${variant.id}/inventory`, "PATCH", { quantity: Number(form.get("quantity")), expected_version: variant.inventory.version, reason: String(form.get("reason")) }); }}>
        <p>{vi ? "Đang giữ / Khả dụng" : "Reserved / Available"}: {variant.inventory.reserved} / {variant.inventory.available}</p><div className="catalog-inline"><label>{vi ? "Tổng tồn kho" : "Stock quantity"}<input name="quantity" type="number" step={1} min={variant.inventory.reserved} max={Number.MAX_SAFE_INTEGER} defaultValue={variant.inventory.quantity} required /></label><label>{vi ? "Lý do điều chỉnh" : "Adjustment reason"}<input name="reason" maxLength={500} required /></label></div><button disabled={busy}>{vi ? "Cập nhật tồn kho" : "Update inventory"}</button>
      </form>
    </article>)}<form key={`add-${selected.id}-${revision}`} className="catalog-form" onSubmit={event => variantSubmit(event)}><h3>{vi ? "Thêm biến thể" : "Add variant"}</h3><VariantFields prefix="" quantity vi={vi} /><button disabled={busy}>{vi ? "Thêm biến thể" : "Add variant"}</button></form></>}
  </section>;
}
