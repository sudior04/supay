package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"supay.local/platform/internal/authn"
	"supay.local/platform/internal/platform"
)

func (a *Catalog) confirmActionPIN(c *gin.Context) {
	if !authn.RequiresActionPIN(c.Request.Method, c.Request.URL.Path) {
		c.Next()
		return
	}
	pin := c.GetHeader("X-PIN")
	c.Request.Header.Del("X-PIN")
	if pin == "" {
		platform.Error(c, 403, "PIN_REQUIRED", "Enter your six-digit PIN to confirm this operation")
		c.Abort()
		return
	}
	valid := len(pin) == 6
	for _, digit := range pin {
		if digit < '0' || digit > '9' {
			valid = false
		}
	}
	if !valid {
		platform.Error(c, 400, "INVALID_PIN_FORMAT", "PIN must contain exactly six ASCII digits")
		c.Abort()
		return
	}
	body, _ := json.Marshal(map[string]string{"pin": pin})
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, a.pinURL, bytes.NewReader(body))
	if err != nil {
		platform.Error(c, 503, "PIN_VERIFICATION_UNAVAILABLE", "PIN verification is unavailable")
		c.Abort()
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", c.GetHeader("Authorization"))
	request.Header.Set("X-Request-ID", c.GetHeader("X-Request-ID"))
	response, err := a.pinClient.Do(request)
	if err != nil {
		platform.Error(c, 503, "PIN_VERIFICATION_UNAVAILABLE", "PIN verification is unavailable")
		c.Abort()
		return
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNoContent {
		c.Next()
		return
	}
	var result struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err = json.NewDecoder(io.LimitReader(response.Body, 16<<10)).Decode(&result)
	if err != nil || result.Error.Code == "" || response.StatusCode < 400 || response.StatusCode >= 500 {
		platform.Error(c, 503, "PIN_VERIFICATION_UNAVAILABLE", "PIN verification is unavailable")
		c.Abort()
		return
	}
	if retry := response.Header.Get("Retry-After"); retry != "" {
		c.Header("Retry-After", retry)
	}
	platform.Error(c, response.StatusCode, result.Error.Code, result.Error.Message)
	c.Abort()
}
