package transport

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/commerce-service/internal/domain"
)

func (a *Catalog) insertVariant(c *gin.Context, tx pgx.Tx, productID string, v domain.VariantInput) error {
	attributes, err := json.Marshal(v.Attributes)
	if err != nil {
		return err
	}
	var id string
	if err = tx.QueryRow(c.Request.Context(), `INSERT INTO product_variants(product_id,sku,name,attributes,price_minor) VALUES ($1,$2,$3,$4::jsonb,$5) RETURNING id::text`, productID, v.SKU, v.Name, string(attributes), v.PriceMinor).Scan(&id); err != nil {
		return err
	}
	if _, err = tx.Exec(c.Request.Context(), `INSERT INTO inventory(variant_id,quantity) VALUES ($1,$2)`, id, v.Quantity); err != nil {
		return err
	}
	_, err = tx.Exec(c.Request.Context(), `INSERT INTO inventory_adjustments(variant_id,actor_id,old_quantity,new_quantity,version,reason) VALUES ($1,$2,0,$3,0,'INITIAL_STOCK')`, id, caller(c).Subject, v.Quantity)
	return err
}
func (a *Catalog) createVariant(c *gin.Context) {
	if !validID(c, "product_id") {
		return
	}
	var in domain.VariantInput
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
	var count int
	if err = tx.QueryRow(c.Request.Context(), `SELECT count(*) FROM product_variants WHERE product_id=$1 AND deleted_at IS NULL`, c.Param("product_id")).Scan(&count); err != nil {
		databaseError(c, err)
		return
	}
	if count >= 100 {
		platform.Error(c, 409, "VARIANT_LIMIT", "A product supports up to 100 variants")
		return
	}
	if err = a.insertVariant(c, tx, c.Param("product_id"), in); err != nil {
		databaseError(c, err)
		return
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	a.respondProduct(c, 201, c.Param("product_id"))
}
func (a *Catalog) updateVariant(c *gin.Context) {
	if !validID(c, "product_id") || !validID(c, "variant_id") {
		return
	}
	var in domain.VariantFields
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
	attributes, _ := json.Marshal(in.Attributes)
	result, err := tx.Exec(c.Request.Context(), `UPDATE product_variants SET sku=$1,name=$2,attributes=$3::jsonb,price_minor=$4,updated_at=now() WHERE id=$5 AND product_id=$6 AND deleted_at IS NULL`, in.SKU, in.Name, string(attributes), in.PriceMinor, c.Param("variant_id"), c.Param("product_id"))
	if err != nil {
		databaseError(c, err)
		return
	}
	if result.RowsAffected() == 0 {
		platform.Error(c, 404, "NOT_FOUND", "Variant was not found")
		return
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	a.respondProduct(c, 200, c.Param("product_id"))
}
func (a *Catalog) deleteVariant(c *gin.Context) {
	if !validID(c, "product_id") || !validID(c, "variant_id") {
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		databaseError(c, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	status, ok := a.lockProduct(c, tx)
	if !ok {
		return
	}
	var exists bool
	if err = tx.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM product_variants WHERE id=$1 AND product_id=$2 AND deleted_at IS NULL)`, c.Param("variant_id"), c.Param("product_id")).Scan(&exists); err != nil {
		databaseError(c, err)
		return
	}
	if !exists {
		platform.Error(c, 404, "NOT_FOUND", "Variant was not found")
		return
	}
	var count int
	if err = tx.QueryRow(c.Request.Context(), `SELECT count(*) FROM product_variants WHERE product_id=$1 AND deleted_at IS NULL`, c.Param("product_id")).Scan(&count); err != nil {
		databaseError(c, err)
		return
	}
	if status == "ACTIVE" && count <= 1 {
		platform.Error(c, 409, "VARIANT_REQUIRED", "Cannot remove the last variant of an active product")
		return
	}
	var reserved int64
	if err = tx.QueryRow(c.Request.Context(), `SELECT reserved FROM inventory WHERE variant_id=$1 FOR UPDATE`, c.Param("variant_id")).Scan(&reserved); err != nil {
		databaseError(c, err)
		return
	}
	if reserved > 0 {
		platform.Error(c, 409, "STOCK_RESERVED", "Variant has reserved inventory")
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `UPDATE product_variants SET deleted_at=now(),updated_at=now() WHERE id=$1`, c.Param("variant_id")); err != nil {
		databaseError(c, err)
		return
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	c.Status(204)
}
func (a *Catalog) updateInventory(c *gin.Context) {
	if !validID(c, "product_id") || !validID(c, "variant_id") {
		return
	}
	var in struct {
		Quantity        *int64 `json:"quantity"`
		ExpectedVersion *int64 `json:"expected_version"`
		Reason          string `json:"reason"`
	}
	if !bind(c, &in) {
		return
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if in.Quantity == nil || *in.Quantity < 0 || *in.Quantity > domain.MaxValue || in.ExpectedVersion == nil || *in.ExpectedVersion < 0 || in.Reason == "" || len([]rune(in.Reason)) > 500 {
		platform.Error(c, 400, "INVALID_INPUT", "Nonnegative quantity, expected_version and a reason (1–500 characters) are required")
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
	var old domain.Inventory
	err = tx.QueryRow(c.Request.Context(), `SELECT i.quantity,i.reserved,i.version FROM inventory i JOIN product_variants v ON v.id=i.variant_id WHERE v.id=$1 AND v.product_id=$2 AND v.deleted_at IS NULL FOR UPDATE OF i`, c.Param("variant_id"), c.Param("product_id")).Scan(&old.Quantity, &old.Reserved, &old.Version)
	if err != nil {
		databaseError(c, err)
		return
	}
	if old.Version != *in.ExpectedVersion {
		platform.Error(c, 409, "INVENTORY_VERSION_CONFLICT", "Reload inventory before applying this change")
		return
	}
	if *in.Quantity < old.Reserved {
		platform.Error(c, 409, "STOCK_RESERVED", "Quantity cannot be lower than reserved inventory")
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `UPDATE inventory SET quantity=$1,version=version+1,updated_at=now() WHERE variant_id=$2`, *in.Quantity, c.Param("variant_id")); err != nil {
		databaseError(c, err)
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `INSERT INTO inventory_adjustments(variant_id,actor_id,old_quantity,new_quantity,version,reason) VALUES ($1,$2,$3,$4,$5,$6)`, c.Param("variant_id"), caller(c).Subject, old.Quantity, *in.Quantity, old.Version+1, in.Reason); err != nil {
		databaseError(c, err)
		return
	}
	if err = tx.Commit(c.Request.Context()); err != nil {
		databaseError(c, err)
		return
	}
	c.JSON(200, gin.H{"variant_id": c.Param("variant_id"), "inventory": domain.Inventory{Quantity: *in.Quantity, Reserved: old.Reserved, Available: *in.Quantity - old.Reserved, Version: old.Version + 1}})
}
