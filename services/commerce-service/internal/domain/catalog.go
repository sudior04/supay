package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const MaxValue int64 = 9007199254740991

var UUIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
var skuPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

type CategoryInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (v *CategoryInput) Validate() error {
	v.Name = strings.TrimSpace(v.Name)
	v.Slug = strings.ToLower(strings.TrimSpace(v.Slug))
	v.Description = strings.TrimSpace(v.Description)
	if len([]rune(v.Name)) < 1 || len([]rune(v.Name)) > 120 || len(v.Slug) > 120 || !slugPattern.MatchString(v.Slug) || len([]rune(v.Description)) > 2000 {
		return fmt.Errorf("Valid category name (1–120), slug (a-z, 0-9, hyphens), and description (up to 2000 characters) are required")
	}
	return nil
}

type Category struct {
	ID string `json:"id"`
	CategoryInput
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type ProductFields struct {
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (v *ProductFields) Validate() error {
	v.Name = strings.TrimSpace(v.Name)
	v.Description = strings.TrimSpace(v.Description)
	v.Status = strings.ToUpper(strings.TrimSpace(v.Status))
	if v.Status == "" {
		v.Status = "DRAFT"
	}
	if !UUIDPattern.MatchString(v.CategoryID) || len([]rune(v.Name)) < 1 || len([]rune(v.Name)) > 150 || len([]rune(v.Description)) > 5000 || (v.Status != "ACTIVE" && v.Status != "DRAFT") {
		return fmt.Errorf("Valid category_id, name (1–150), description (up to 5000), and status ACTIVE or DRAFT are required")
	}
	return nil
}

type VariantFields struct {
	SKU        string            `json:"sku"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
	PriceMinor int64             `json:"price_minor"`
}

func (v *VariantFields) Validate() error {
	v.SKU = strings.TrimSpace(v.SKU)
	v.Name = strings.TrimSpace(v.Name)
	if v.Attributes == nil {
		v.Attributes = map[string]string{}
	}
	if len(v.SKU) > 80 || !skuPattern.MatchString(v.SKU) || len([]rune(v.Name)) < 1 || len([]rune(v.Name)) > 120 || v.PriceMinor < 1 || v.PriceMinor > MaxValue || len(v.Attributes) > 10 {
		return fmt.Errorf("Variant requires SKU (1–80 letters/digits/_.-), name (1–120), positive integer VND price, and up to 10 attributes")
	}
	normalized := make(map[string]string, len(v.Attributes))
	for k, val := range v.Attributes {
		k = strings.ToLower(strings.TrimSpace(k))
		val = strings.TrimSpace(val)
		if k == "" || len([]rune(k)) > 64 || val == "" || len([]rune(val)) > 100 {
			return fmt.Errorf("Attribute keys must be 1–64 and values 1–100 characters")
		}
		if _, exists := normalized[k]; exists {
			return fmt.Errorf("Attribute names must be unique after normalization")
		}
		normalized[k] = val
	}
	v.Attributes = normalized
	return nil
}

type VariantInput struct {
	VariantFields
	Quantity int64 `json:"quantity"`
}

func (v *VariantInput) Validate() error {
	if err := v.VariantFields.Validate(); err != nil {
		return err
	}
	if v.Quantity < 0 || v.Quantity > MaxValue {
		return fmt.Errorf("Quantity must be a nonnegative safe integer")
	}
	return nil
}

type ProductInput struct {
	ProductFields
	Variants []VariantInput `json:"variants"`
}

func (v *ProductInput) Validate() error {
	if err := v.ProductFields.Validate(); err != nil {
		return err
	}
	if len(v.Variants) < 1 || len(v.Variants) > 100 {
		return fmt.Errorf("A product needs between 1 and 100 variants")
	}
	for i := range v.Variants {
		if err := v.Variants[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Inventory struct {
	Quantity  int64 `json:"quantity"`
	Reserved  int64 `json:"reserved"`
	Available int64 `json:"available"`
	Version   int64 `json:"version"`
}
type Variant struct {
	ID string `json:"id"`
	VariantFields
	Currency  string    `json:"currency"`
	Inventory Inventory `json:"inventory"`
}
type Product struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchant_id"`
	ProductFields
	CategoryName string    `json:"category_name"`
	Variants     []Variant `json:"variants"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
