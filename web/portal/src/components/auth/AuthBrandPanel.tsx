import Link from "next/link";
import type { Language } from "@/i18n/provider";

export default function AuthBrandPanel({ language }: { language: Language }) {
  return <div className="auth-decoration"><Link className="brand" href="/"><span className="brand-mark">s</span><span>supay<span className="brand-dot">.</span></span></Link><div className="auth-decoration-copy"><span>Supay · {language === "vi" ? "Thương mại mỗi ngày" : "Everyday commerce"}</span><strong>{language === "vi" ? <>Mọi điều tốt đẹp<br />bắt đầu từ một<br /><em>kết nối nhỏ.</em></> : <>Good things<br />begin with a<br /><em>small connection.</em></>}</strong><p>{language === "vi" ? "Mua sắm, bán hàng và quản lý ví trong cùng một nền tảng." : "Shop, sell, and manage your wallet in one platform."}</p></div><span className="auth-decoration-foot">Shop · Sell · Pay</span></div>;
}
