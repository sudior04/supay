package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"supay.local/platform/internal/authn"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/commerce-service/internal/domain"
)

type Catalog struct {
	db        *pgxpool.Pool
	secret    []byte
	pinURL    string
	pinClient *http.Client
}

func Register(r *gin.Engine, db *pgxpool.Pool) error {
	secret := []byte(platform.Env("JWT_SECRET", ""))
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	userURL, err := url.Parse(platform.Env("USER_SERVICE_URL", "http://localhost:8081"))
	if err != nil || userURL.Host == "" || (userURL.Scheme != "http" && userURL.Scheme != "https") {
		return fmt.Errorf("Invalid USER_SERVICE_URL")
	}
	a := &Catalog{db: db, secret: secret, pinURL: userURL.ResolveReference(&url.URL{Path: "/api/v1/auth/pin/verify"}).String(), pinClient: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	r.GET("/api/v1/products/status", func(c *gin.Context) { c.JSON(200, gin.H{"service": "commerce-service", "stage": "catalog"}) })
	g := r.Group("/api/v1", a.authenticate, a.confirmActionPIN)
	g.GET("/categories", a.listCategories)
	g.GET("/categories/:category_id", a.getCategory)
	g.POST("/categories", a.admin, a.createCategory)
	g.PUT("/categories/:category_id", a.admin, a.updateCategory)
	g.DELETE("/categories/:category_id", a.admin, a.deleteCategory)
	g.GET("/products", a.listProducts)
	g.GET("/products/:product_id", a.getProduct)
	g.POST("/products", a.seller, a.createProduct)
	g.PUT("/products/:product_id", a.seller, a.updateProduct)
	g.DELETE("/products/:product_id", a.seller, a.deleteProduct)
	g.POST("/products/:product_id/variants", a.seller, a.createVariant)
	g.PUT("/products/:product_id/variants/:variant_id", a.seller, a.updateVariant)
	g.DELETE("/products/:product_id/variants/:variant_id", a.seller, a.deleteVariant)
	g.PATCH("/products/:product_id/variants/:variant_id/inventory", a.seller, a.updateInventory)
	return nil
}
func (a *Catalog) authenticate(c *gin.Context) {
	fields := strings.Fields(c.GetHeader("Authorization"))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		platform.Error(c, 401, "UNAUTHORIZED", "A valid access token is required")
		c.Abort()
		return
	}
	claims, err := authn.Verify(fields[1], a.secret)
	if err != nil {
		platform.Error(c, 401, "UNAUTHORIZED", "A valid access token is required")
		c.Abort()
		return
	}
	c.Set("catalog_claims", claims)
	c.Next()
}
func caller(c *gin.Context) authn.Claims { return c.MustGet("catalog_claims").(authn.Claims) }
func (a *Catalog) admin(c *gin.Context) {
	if caller(c).Role != "ADMIN" {
		platform.Error(c, 403, "FORBIDDEN", "Administrator role is required")
		c.Abort()
		return
	}
	c.Next()
}
func (a *Catalog) seller(c *gin.Context) {
	role := caller(c).Role
	if role != "ADMIN" && role != "MERCHANT" {
		platform.Error(c, 403, "FORBIDDEN", "Merchant role is required")
		c.Abort()
		return
	}
	c.Next()
}
func bind(c *gin.Context, dst any) bool {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		platform.Error(c, 415, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 128<<10))
	dec.DisallowUnknownFields()
	if err = dec.Decode(dst); err != nil {
		var big *http.MaxBytesError
		if errors.As(err, &big) {
			platform.Error(c, 413, "PAYLOAD_TOO_LARGE", "Request body exceeds 128 KiB")
		} else {
			platform.Error(c, 400, "INVALID_INPUT", "Invalid JSON or unknown field")
		}
		return false
	}
	if err = dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		platform.Error(c, 400, "INVALID_INPUT", "One JSON object is required")
		return false
	}
	return true
}
func validID(c *gin.Context, key string) bool {
	if !domain.UUIDPattern.MatchString(c.Param(key)) {
		platform.Error(c, 400, "INVALID_ID", "Path ID must be a UUID")
		return false
	}
	return true
}
func databaseError(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(c, 404, "NOT_FOUND", "Resource was not found")
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			platform.Error(c, 409, "DUPLICATE_RESOURCE", "Slug, SKU, or variant attributes already exist")
			return
		case "23503":
			platform.Error(c, 409, "RESOURCE_IN_USE", "Referenced resource is missing or still in use")
			return
		case "23514":
			platform.Error(c, 400, "INVALID_INPUT", "Value violates a catalog constraint")
			return
		}
	}
	platform.Error(c, 500, "INTERNAL_ERROR", "Could not complete catalog operation")
}
func pagination(c *gin.Context) (int, int, bool) {
	limit, offset := 20, 0
	var err error
	if raw := c.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			platform.Error(c, 400, "INVALID_PAGINATION", "Invalid limit")
			return 0, 0, false
		}
	}
	if raw := c.Query("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil {
			platform.Error(c, 400, "INVALID_PAGINATION", "Invalid offset")
			return 0, 0, false
		}
	}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		platform.Error(c, 400, "INVALID_PAGINATION", "Limit must be 1–100 and offset 0–100000")
		return 0, 0, false
	}
	return limit, offset, true
}
