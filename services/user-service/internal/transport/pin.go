package transport

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"supay.local/platform/internal/platform"
)

func validPIN(pin string) bool {
	if len(pin) != 6 {
		return false
	}
	for _, digit := range pin {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
func (a *Auth) pinMaterial(userID, pin string) []byte {
	mac := hmac.New(sha256.New, a.pinPepper)
	mac.Write([]byte("supay-pin-v1:" + userID + ":" + pin))
	return []byte(hex.EncodeToString(mac.Sum(nil)))
}
func (a *Auth) hashPIN(userID, pin string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(a.pinMaterial(userID, pin), bcrypt.DefaultCost)
	return string(hash), err
}

type pinCredential struct {
	passwordHash string
	hash         *string
	failures     int
	lockedUntil  *time.Time
}

func (a *Auth) loadPIN(c *gin.Context, tx pgx.Tx, userID string) (pinCredential, bool) {
	var state pinCredential
	err := tx.QueryRow(c.Request.Context(), `SELECT cr.password_hash,cr.pin_hash,cr.pin_failures,cr.pin_locked_until FROM user_credentials cr JOIN users u ON u.id=cr.user_id WHERE u.id=$1 AND u.status='ACTIVE' FOR UPDATE OF cr`, userID).Scan(&state.passwordHash, &state.hash, &state.failures, &state.lockedUntil)
	if err == pgx.ErrNoRows {
		platform.Error(c, 401, "UNAUTHORIZED", "Account unavailable")
		return state, false
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not load PIN settings")
		return state, false
	}
	return state, true
}

// Failure counts are committed even though the business operation is rejected.
// Success only confirms this request; no reusable PIN privilege is issued.
func (a *Auth) checkPIN(c *gin.Context, tx pgx.Tx, userID, pin string, state pinCredential) bool {
	if state.hash == nil {
		platform.Error(c, 409, "PIN_NOT_CONFIGURED", "Set up a six-digit PIN in account security before changing products or orders")
		return false
	}
	if state.lockedUntil != nil && state.lockedUntil.After(time.Now()) {
		c.Header("Retry-After", strconv.Itoa(int(time.Until(*state.lockedUntil).Seconds())+1))
		platform.Error(c, 429, "PIN_LOCKED", "PIN verification is temporarily locked")
		return false
	}
	if state.lockedUntil != nil {
		state.failures = 0
	}
	if bcrypt.CompareHashAndPassword([]byte(*state.hash), a.pinMaterial(userID, pin)) != nil {
		state.failures++
		if _, err := tx.Exec(c.Request.Context(), `UPDATE user_credentials SET pin_failures=$2,pin_locked_until=CASE WHEN $2>=5 THEN now()+interval '15 minutes' ELSE NULL END WHERE user_id=$1`, userID, state.failures); err != nil {
			platform.Error(c, 500, "INTERNAL_ERROR", "Could not record PIN attempt")
			return false
		}
		if tx.Commit(c.Request.Context()) != nil {
			platform.Error(c, 500, "INTERNAL_ERROR", "Could not record PIN attempt")
			return false
		}
		if state.failures >= 5 {
			c.Header("Retry-After", "900")
			platform.Error(c, 429, "PIN_LOCKED", "PIN verification is locked for 15 minutes")
		} else {
			platform.Error(c, 401, "INVALID_PIN", "PIN is incorrect")
		}
		return false
	}
	if _, err := tx.Exec(c.Request.Context(), `UPDATE user_credentials SET pin_failures=0,pin_locked_until=NULL WHERE user_id=$1`, userID); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not verify PIN")
		return false
	}
	return true
}

func (a *Auth) verifyActionPIN(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var in struct {
		PIN string `json:"pin"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	if !validPIN(in.PIN) {
		platform.Error(c, 400, "INVALID_PIN_FORMAT", "PIN must contain exactly six ASCII digits")
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not verify PIN")
		return
	}
	defer tx.Rollback(c.Request.Context())
	state, ok := a.loadPIN(c, tx, claims.Subject)
	if !ok || !a.checkPIN(c, tx, claims.Subject, in.PIN, state) {
		return
	}
	if tx.Commit(c.Request.Context()) != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not verify PIN")
		return
	}
	c.Status(204)
}

func (a *Auth) pinStatus(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var configured bool
	var lockedUntil *time.Time
	err := a.db.QueryRow(c.Request.Context(), `SELECT cr.pin_hash IS NOT NULL,cr.pin_locked_until FROM user_credentials cr JOIN users u ON u.id=cr.user_id WHERE u.id=$1 AND u.status='ACTIVE'`, claims.Subject).Scan(&configured, &lockedUntil)
	if err == pgx.ErrNoRows {
		platform.Error(c, 401, "UNAUTHORIZED", "Account unavailable")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not load PIN settings")
		return
	}
	c.JSON(200, gin.H{"pin_configured": configured, "pin_locked_until": lockedUntil})
}

func (a *Auth) setupPIN(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var in struct {
		Password string `json:"current_password"`
		PIN      string `json:"pin"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	if !validPIN(in.PIN) || in.Password == "" || len(in.Password) > 72 {
		platform.Error(c, 400, "INVALID_INPUT", "Current password and a six-digit PIN are required")
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not set PIN")
		return
	}
	defer tx.Rollback(c.Request.Context())
	state, ok := a.loadPIN(c, tx, claims.Subject)
	if !ok {
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(state.passwordHash), []byte(in.Password)) != nil {
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Current password is incorrect")
		return
	}
	if state.hash != nil {
		platform.Error(c, 409, "PIN_ALREADY_CONFIGURED", "Use change PIN to replace an existing PIN")
		return
	}
	hash, err := a.hashPIN(claims.Subject, in.PIN)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not hash PIN")
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `UPDATE user_credentials SET pin_hash=$2,pin_failures=0,pin_locked_until=NULL WHERE user_id=$1`, claims.Subject, hash); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not set PIN")
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `INSERT INTO pin_audit_events(actor_user_id,target_user_id,action) VALUES ($1,$1,'ENROLL')`, claims.Subject); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not record PIN enrollment")
		return
	}
	if tx.Commit(c.Request.Context()) != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not set PIN")
		return
	}
	c.Status(204)
}

func (a *Auth) changePIN(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var in struct {
		Password   string `json:"current_password"`
		CurrentPIN string `json:"current_pin"`
		NewPIN     string `json:"new_pin"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	if !validPIN(in.CurrentPIN) || !validPIN(in.NewPIN) || in.Password == "" || len(in.Password) > 72 {
		platform.Error(c, 400, "INVALID_INPUT", "Password and two six-digit PINs are required")
		return
	}
	if in.CurrentPIN == in.NewPIN {
		platform.Error(c, 400, "PIN_UNCHANGED", "New PIN must differ from current PIN")
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not change PIN")
		return
	}
	defer tx.Rollback(c.Request.Context())
	state, ok := a.loadPIN(c, tx, claims.Subject)
	if !ok {
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(state.passwordHash), []byte(in.Password)) != nil {
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Current password is incorrect")
		return
	}
	if !a.checkPIN(c, tx, claims.Subject, in.CurrentPIN, state) {
		return
	}
	hash, err := a.hashPIN(claims.Subject, in.NewPIN)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not hash PIN")
		return
	}
	if _, err = tx.Exec(c.Request.Context(), `UPDATE user_credentials SET pin_hash=$2 WHERE user_id=$1`, claims.Subject, hash); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not change PIN")
		return
	}
	if !a.revokePINState(c, tx, claims.Subject, claims.Subject, "CHANGE") {
		return
	}
	if tx.Commit(c.Request.Context()) != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not change PIN")
		return
	}
	a.clearCookie(c)
	c.Status(204)
}

func (a *Auth) adminResetPIN(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var allowed bool
	if err := a.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND role='ADMIN' AND status='ACTIVE')`, claims.Subject).Scan(&allowed); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not authorize operation")
		return
	}
	if !allowed || claims.Role != "ADMIN" {
		platform.Error(c, 403, "FORBIDDEN", "Active administrator role required")
		return
	}
	target := c.Param("user_id")
	if !uuidPattern.MatchString(target) {
		platform.Error(c, 400, "INVALID_USER_ID", "User ID must be a UUID")
		return
	}
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not reset PIN")
		return
	}
	defer tx.Rollback(c.Request.Context())
	result, err := tx.Exec(c.Request.Context(), `UPDATE user_credentials SET pin_hash=NULL,pin_failures=0,pin_locked_until=NULL WHERE user_id=$1`, target)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not reset PIN")
		return
	}
	if result.RowsAffected() == 0 {
		platform.Error(c, 404, "USER_NOT_FOUND", "User was not found")
		return
	}
	if !a.revokePINState(c, tx, claims.Subject, target, "ADMIN_RESET") {
		return
	}
	if tx.Commit(c.Request.Context()) != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not reset PIN")
		return
	}
	if target == claims.Subject {
		a.clearCookie(c)
	}
	c.Status(204)
}

func (a *Auth) revokePINState(c *gin.Context, tx pgx.Tx, actor, target, action string) bool {
	if _, err := tx.Exec(c.Request.Context(), `UPDATE refresh_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, target); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not revoke sessions")
		return false
	}
	if _, err := tx.Exec(c.Request.Context(), `INSERT INTO pin_audit_events(actor_user_id,target_user_id,action) VALUES ($1,$2,$3)`, actor, target, action); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not record PIN event")
		return false
	}
	return true
}
