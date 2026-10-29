package transport

import (
	"github.com/gin-gonic/gin"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/commerce-service/internal/domain"
)

func (a *Catalog) listCategories(c *gin.Context) {
	limit, offset, ok := pagination(c)
	if !ok {
		return
	}
	q := c.Query("q")
	if len(q) > 200 {
		platform.Error(c, 400, "INVALID_INPUT", "Search query is too long")
		return
	}
	rows, err := a.db.Query(c.Request.Context(), `SELECT id::text,name,slug,description,created_at,updated_at FROM categories WHERE name ILIKE '%'||$1||'%' ORDER BY name,id LIMIT $2 OFFSET $3`, q, limit+1, offset)
	if err != nil {
		databaseError(c, err)
		return
	}
	defer rows.Close()
	items := make([]domain.Category, 0)
	for rows.Next() {
		var v domain.Category
		if err = rows.Scan(&v.ID, &v.Name, &v.Slug, &v.Description, &v.CreatedAt, &v.UpdatedAt); err != nil {
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
	c.JSON(200, gin.H{"categories": items, "pagination": gin.H{"limit": limit, "offset": offset, "has_more": more}})
}
func (a *Catalog) getCategory(c *gin.Context) {
	if !validID(c, "category_id") {
		return
	}
	a.respondCategory(c, 200, c.Param("category_id"))
}
func (a *Catalog) respondCategory(c *gin.Context, status int, id string) {
	var v domain.Category
	err := a.db.QueryRow(c.Request.Context(), `SELECT id::text,name,slug,description,created_at,updated_at FROM categories WHERE id=$1`, id).Scan(&v.ID, &v.Name, &v.Slug, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		databaseError(c, err)
		return
	}
	c.JSON(status, gin.H{"category": v})
}
func (a *Catalog) createCategory(c *gin.Context) {
	var in domain.CategoryInput
	if !bind(c, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		platform.Error(c, 400, "INVALID_INPUT", err.Error())
		return
	}
	var id string
	err := a.db.QueryRow(c.Request.Context(), `INSERT INTO categories(name,slug,description) VALUES ($1,$2,$3) RETURNING id::text`, in.Name, in.Slug, in.Description).Scan(&id)
	if err != nil {
		databaseError(c, err)
		return
	}
	a.respondCategory(c, 201, id)
}
func (a *Catalog) updateCategory(c *gin.Context) {
	if !validID(c, "category_id") {
		return
	}
	var in domain.CategoryInput
	if !bind(c, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		platform.Error(c, 400, "INVALID_INPUT", err.Error())
		return
	}
	result, err := a.db.Exec(c.Request.Context(), `UPDATE categories SET name=$1,slug=$2,description=$3,updated_at=now() WHERE id=$4`, in.Name, in.Slug, in.Description, c.Param("category_id"))
	if err != nil {
		databaseError(c, err)
		return
	}
	if result.RowsAffected() == 0 {
		platform.Error(c, 404, "NOT_FOUND", "Category was not found")
		return
	}
	a.respondCategory(c, 200, c.Param("category_id"))
}
func (a *Catalog) deleteCategory(c *gin.Context) {
	if !validID(c, "category_id") {
		return
	}
	result, err := a.db.Exec(c.Request.Context(), `DELETE FROM categories WHERE id=$1`, c.Param("category_id"))
	if err != nil {
		databaseError(c, err)
		return
	}
	if result.RowsAffected() == 0 {
		platform.Error(c, 404, "NOT_FOUND", "Category was not found")
		return
	}
	c.Status(204)
}
