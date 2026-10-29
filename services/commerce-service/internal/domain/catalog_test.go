package domain

import "testing"

func TestVariantRejectsInvalidMoneyStockAndAttributes(t *testing.T) {
	valid := func() VariantInput {
		return VariantInput{VariantFields: VariantFields{SKU: "TEE-M", Name: "Medium", PriceMinor: 100000, Attributes: map[string]string{"size": "M"}}, Quantity: 10}
	}
	v := valid()
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*VariantInput){func(v *VariantInput) { v.PriceMinor = 0 }, func(v *VariantInput) { v.PriceMinor = MaxValue + 1 }, func(v *VariantInput) { v.Quantity = -1 }, func(v *VariantInput) { v.Attributes[""] = "M" }, func(v *VariantInput) { v.SKU = "invalid sku" }} {
		v := valid()
		mutate(&v)
		if v.Validate() == nil {
			t.Fatal("unsafe variant accepted")
		}
	}
}
func TestProductRequiresCategoryAndAtLeastOneVariant(t *testing.T) {
	v := ProductInput{ProductFields: ProductFields{CategoryID: "a12cc857-fec7-4b70-8ec0-9692eb21aadc", Name: "Tee", Status: "ACTIVE"}}
	if v.Validate() == nil {
		t.Fatal("product without variants accepted")
	}
	v.Variants = []VariantInput{{VariantFields: VariantFields{SKU: "TEE-M", Name: "Medium", PriceMinor: 1000}}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	v.CategoryID = "invalid"
	if v.Validate() == nil {
		t.Fatal("invalid category ID accepted")
	}
}

func TestAttributesCanonicalizeNamesAndRejectDuplicateKeys(t *testing.T) {
	v := VariantFields{SKU: "TEE", Name: "Tee", PriceMinor: 100, Attributes: map[string]string{" Size ": " M "}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(v.Attributes) != 1 || v.Attributes["size"] != "M" {
		t.Fatal("attribute normalization failed")
	}
	v.Attributes = map[string]string{" Size ": "M", "size": "L"}
	if v.Validate() == nil {
		t.Fatal("duplicate normalized keys accepted")
	}
}
