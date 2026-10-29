package transport

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/commerce-service/internal/domain"
)

const productProjection = `SELECT p.id::text,p.merchant_id::text,p.category_id::text,p.name,p.description,p.status,c.name,p.created_at,p.updated_at,
 COALESCE((SELECT jsonb_agg(jsonb_build_object('id',v.id,'sku',v.sku,'name',v.name,'attributes',v.attributes,'price_minor',v.price_minor,'currency',v.currency,'inventory',jsonb_build_object('quantity',i.quantity,'reserved',i.reserved,'available',i.quantity-i.reserved,'version',i.version)) ORDER BY v.created_at,v.id) FROM product_variants v JOIN inventory i ON i.variant_id=v.id WHERE v.product_id=p.id AND v.deleted_at IS NULL),'[]'::jsonb)
 FROM products p JOIN categories c ON c.id=p.category_id `

type scanner interface{ Scan(...any) error }

func scanProduct(row scanner) (domain.Product, error) {
	var p domain.Product
	var variants []byte
	err := row.Scan(&p.ID, &p.MerchantID, &p.CategoryID, &p.Name, &p.Description, &p.Status, &p.CategoryName, &p.CreatedAt, &p.UpdatedAt, &variants)
	if err == nil {
		err = json.Unmarshal(variants, &p.Variants)
	}
	return p, err
}
func (a *Catalog) respondProduct(c *gin.Context, status int, id string) {
	who := caller(c)
	p, err := scanProduct(a.db.QueryRow(c.Request.Context(), productProjection+`WHERE p.id=$1 AND p.deleted_at IS NULL AND (p.status='ACTIVE' OR p.merchant_id=$2 OR $3='ADMIN')`, id, who.Subject, who.Role))
	if err != nil {
		databaseError(c, err)
		return
	}
	c.JSON(status, gin.H{"product": p})
}
func (a *Catalog) getProduct(c *gin.Context) {
	if !validID(c, "product_id") {
		return
	}
	a.respondProduct(c, 200, c.Param("product_id"))
}
func (a *Catalog) listProducts(c *gin.Context) {
	limit, offset, ok := pagination(c)
	if !ok {
		return
	}
	who := caller(c)
	args := []any{who.Subject, who.Role}
	where := `WHERE p.deleted_at IS NULL AND (p.status='ACTIVE' OR p.merchant_id=$1 OR $2='ADMIN')`
	add := func(clause string, value any) { args = append(args, value); where += fmt.Sprintf(clause, len(args)) }
	if raw := c.Query("mine"); raw != "" {
		mine, err := strconv.ParseBool(raw)
		if err != nil {
			platform.Error(c, 400, "INVALID_INPUT", "mine must be true or false")
			return
		}
		if mine {
			if who.Role != "MERCHANT" && who.Role != "ADMIN" {
				platform.Error(c, 403, "FORBIDDEN", "Merchant role is required")
				return
			}
			add(` AND p.merchant_id=$%d`, who.Subject)
		}
	}
	if id := c.Query("category_id"); id != "" {
		if !domain.UUIDPattern.MatchString(id) {
			platform.Error(c, 400, "INVALID_ID", "category_id must be a UUID")
			return
		}
		add(` AND p.category_id=$%d`, id)
	}
	if id := c.Query("merchant_id"); id != "" {
		if !domain.UUIDPattern.MatchString(id) {
			platform.Error(c, 400, "INVALID_ID", "merchant_id must be a UUID")
			return
		}
		add(` AND p.merchant_id=$%d`, id)
	}
	if status := c.Query("status"); status != "" {
		if status != "DRAFT" && status != "ACTIVE" {
			platform.Error(c, 400, "INVALID_INPUT", "status must be ACTIVE or DRAFT")
			return
		}
		add(` AND p.status=$%d`, status)
	}
	if q := c.Query("q"); q != "" {
		if len(q) > 200 {
			platform.Error(c, 400, "INVALID_INPUT", "Search query is too long")
			return
		}
		add(` AND p.name ILIKE '%%'||$%d||'%%'`, q)
	}
	args = append(args, limit+1, offset)
	query := productProjection + where + fmt.Sprintf(` ORDER BY p.created_at DESC,p.id LIMIT $%d OFFSET $%d`, len(args)-1, len(args))
	rows, err := a.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		databaseError(c, err)
		return
	}
	defer rows.Close()
	items := make([]domain.Product, 0)
	for rows.Next() {
		v, err := scanProduct(rows)
		if err != nil {
			databaseError(c, err)
			return
		}
		items = append(items, v)
	}
	if rows.Err() != nil {
		databaseError(c, rows.Err())
		return
	}
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	c.JSON(200, gin.H{"products": items, "pagination": gin.H{"limit": limit, "offset": offset, "has_more": more}})
}
func (a *Catalog) lockProduct(c *gin.Context, tx pgx.Tx) (string, bool) {
	var owner, status string
	err := tx.QueryRow(c.Request.Context(), `SELECT merchant_id::text,status FROM products WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, c.Param("product_id")).Scan(&owner, &status)
	if err != nil {
		databaseError(c, err)
		return "", false
	}
	who := caller(c)
	if owner != who.Subject && who.Role != "ADMIN" {
		platform.Error(c, 403, "FORBIDDEN", "Product belongs to another merchant")
		return "", false
	}
	return status, true
}
func (a *Catalog) createProduct(c *gin.Context) {
	var in domain.ProductInput
	if !bind(c, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		platform.Error(c, 400, "INVALID_INPUT", err.Error())
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		databaseError(c, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	var id string
	err = tx.QueryRow(c.Request.Context(), `INSERT INTO products(merchant_id,category_id,name,description,status) VALUES ($1,$2,$3,$4,$5) RETURNING id::text`, caller(c).Subject, in.CategoryID, in.Name, in.Description, in.Status).Scan(&id)
	if err != nil {
		databaseError(c, err)
		return
	}
	for _, v := range in.Variants {
		if err = a.insertVariant(c, tx, id, v); err != nil {
			databaseError(c, err)
			return
		}
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	a.respondProduct(c, 201, id)
}
func (a *Catalog) updateProduct(c *gin.Context) {
	if !validID(c, "product_id") {
		return
	}
	var in domain.ProductFields
	if !bind(c, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		platform.Error(c, 400, "INVALID_INPUT", err.Error())
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		databaseError(c, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	if _, ok := a.lockProduct(c, tx); !ok {
		return
	}
	if in.Status == "ACTIVE" {
		var count int
		if err = tx.QueryRow(c.Request.Context(), `SELECT count(*) FROM product_variants WHERE product_id=$1 AND deleted_at IS NULL`, c.Param("product_id")).Scan(&count); err != nil {
			databaseError(c, err)
			return
		}
		if count == 0 {
			platform.Error(c, 409, "VARIANT_REQUIRED", "An active product needs at least one variant")
			return
		}
	}
	_, err = tx.Exec(c.Request.Context(), `UPDATE products SET category_id=$1,name=$2,description=$3,status=$4,updated_at=now() WHERE id=$5`, in.CategoryID, in.Name, in.Description, in.Status, c.Param("product_id"))
	if err != nil {
		databaseError(c, err)
		return
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	a.respondProduct(c, 200, c.Param("product_id"))
}
func (a *Catalog) deleteProduct(c *gin.Context) {
	if !validID(c, "product_id") {
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		databaseError(c, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	if _, ok := a.lockProduct(c, tx); !ok {
		return
	}
	var reserved bool
	if err = tx.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM inventory i JOIN product_variants v ON v.id=i.variant_id WHERE v.product_id=$1 AND i.reserved>0)`, c.Param("product_id")).Scan(&reserved); err != nil {
		databaseError(c, err)
		return
	}
	if reserved {
		platform.Error(c, 409, "STOCK_RESERVED", "Product has reserved inventory")
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `UPDATE products SET deleted_at=now(),updated_at=now() WHERE id=$1`, c.Param("product_id")); err != nil {
		databaseError(c, err)
		return
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	c.Status(204)
}
