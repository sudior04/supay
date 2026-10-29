"use client";
import { useEffect, useState, type FormEvent } from "react";
import { useI18n } from "@/i18n/provider";
import { catalogRequest, type Category, type Page } from "@/services/catalog";
import "./catalog.css";

export default function CategoryManager() {
  const { language } = useI18n(); const vi = language === "vi";
  const [items, setItems] = useState<Category[]>([]); const [editing, setEditing] = useState<Category | null>(null);
  const [message, setMessage] = useState(""); const [busy, setBusy] = useState(false); const [offset, setOffset] = useState(0); const [more, setMore] = useState(false);
  async function load(pageOffset = offset) { const page = await catalogRequest<Page<"categories", Category>>(`categories?limit=20&offset=${pageOffset}`); setItems(page.categories); setMore(page.pagination.has_more); }
  useEffect(() => { load(offset).catch(error => setMessage(String(error.message))); /* page-dependent request */ }, [offset]);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); const formElement = event.currentTarget; const form = new FormData(formElement); setBusy(true); setMessage("");
    try {
      await catalogRequest(editing ? `categories/${editing.id}` : "categories", editing ? "PUT" : "POST", { name: String(form.get("name")), slug: String(form.get("slug")), description: String(form.get("description")) });
      setEditing(null); formElement.reset(); await load(); setMessage(vi ? "Đã lưu danh mục." : "Category saved.");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Request failed"); } finally { setBusy(false); }
  }
  async function remove(category: Category) {
    if (!window.confirm(`${vi ? "Xóa danh mục" : "Delete category"} ${category.name}?`)) return;
    setBusy(true); try { await catalogRequest(`categories/${category.id}`, "DELETE"); if (editing?.id === category.id) setEditing(null); await load(); setMessage(vi ? "Đã xóa danh mục." : "Category deleted."); } catch (error) { setMessage(error instanceof Error ? error.message : "Request failed"); } finally { setBusy(false); }
  }
  return <div className="catalog-manager"><h2>{vi ? "Quản lý danh mục" : "Categories"}</h2>{message && <p role="status" className="catalog-message">{message}</p>}
    <form key={editing?.id ?? "new"} className="catalog-form" onSubmit={submit}>
      <h3>{editing ? (vi ? "Sửa danh mục" : "Edit category") : (vi ? "Thêm danh mục" : "Add category")}</h3>
      <label>{vi ? "Tên" : "Name"}<input name="name" defaultValue={editing?.name} maxLength={120} required /></label>
      <label>Slug<input name="slug" defaultValue={editing?.slug} pattern="[a-z0-9]+(-[a-z0-9]+)*" maxLength={120} placeholder="ao-thun" required /></label>
      <label>{vi ? "Mô tả" : "Description"}<textarea name="description" defaultValue={editing?.description} maxLength={2000} /></label>
      <div className="catalog-actions"><button disabled={busy}>{vi ? "Lưu danh mục" : "Save category"}</button>{editing && <button type="button" className="secondary" onClick={() => setEditing(null)}>{vi ? "Hủy" : "Cancel"}</button>}</div>
    </form>
    <div className="workspace-table-wrap"><table className="workspace-table"><thead><tr><th>{vi ? "Danh mục" : "Category"}</th><th>Slug</th><th /></tr></thead><tbody>{items.map(category => <tr key={category.id}><td>{category.name}</td><td>{category.slug}</td><td className="catalog-actions"><button disabled={busy} onClick={() => setEditing(category)}>{vi ? "Sửa" : "Edit"}</button><button disabled={busy} className="danger" onClick={() => remove(category)}>{vi ? "Xóa" : "Delete"}</button></td></tr>)}</tbody></table></div>
    <div className="catalog-actions"><button disabled={offset === 0 || busy} onClick={() => setOffset(Math.max(0, offset - 20))}>←</button><span>{offset + 1}–{offset + items.length}</span><button disabled={!more || busy} onClick={() => setOffset(offset + 20)}>→</button></div>
  </div>;
}
