package transport

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"supay.local/platform/internal/authn"
	"supay.local/platform/internal/platform"
)

const refreshCookie = "supay_refresh"
const accessTTL = 15 * time.Minute
const refreshTTL = 30 * 24 * time.Hour
const maxAuthBody = 16 << 10

var dummyPasswordHash = []byte("$2a$10$iKMBBXCxBfzDf3opsXWvy.ULzjHdwTk33pdYSlaaCYqL2xV8yduEW")
var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type Auth struct {
	db           *pgxpool.Pool
	secret       []byte
	pinPepper    []byte
	secureCookie bool
}
type account struct{ ID, Name, Email, Role string }

func Register(r *gin.Engine, db *pgxpool.Pool) error {
	secret := []byte(platform.Env("JWT_SECRET", ""))
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	pinPepper := []byte(platform.Env("PIN_PEPPER", ""))
	if len(pinPepper) < 32 {
		return fmt.Errorf("PIN_PEPPER must contain at least 32 bytes")
	}
	secureCookie, err := strconv.ParseBool(platform.Env("COOKIE_SECURE", "true"))
	if err != nil {
		return fmt.Errorf("COOKIE_SECURE must be true or false")
	}
	a := &Auth{db: db, secret: secret, pinPepper: pinPepper, secureCookie: secureCookie}
	r.Use(func(c *gin.Context) { c.Header("Cache-Control", "no-store"); c.Next() })
	r.POST("/api/v1/auth/register", a.register)
	r.POST("/api/v1/auth/login", a.login)
	r.POST("/api/v1/auth/refresh", a.refresh)
	r.POST("/api/v1/auth/logout", a.logout)
	r.POST("/api/v1/auth/pin/verify", a.verifyActionPIN)
	r.GET("/api/v1/users/me/pin", a.pinStatus)
	r.POST("/api/v1/users/me/pin", a.setupPIN)
	r.PUT("/api/v1/users/me/pin", a.changePIN)
	r.POST("/api/v1/users/:user_id/pin/reset", a.adminResetPIN)
	r.GET("/api/v1/users/me", a.me)
	r.PUT("/api/v1/users/me/password", a.changePassword)
	r.POST("/api/v1/users/:user_id/password/reset", a.adminResetPassword)
	r.GET("/api/v1/users/status", func(c *gin.Context) { c.JSON(200, gin.H{"service": "user-service", "stage": "auth"}) })
	return nil
}

func (a *Auth) register(c *gin.Context) {
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		PIN      string `json:"pin"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	if in.PIN != "" && !validPIN(in.PIN) {
		platform.Error(c, 400, "INVALID_PIN_FORMAT", "PIN must contain exactly six ASCII digits")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	parsed, err := mail.ParseAddress(in.Email)
	if in.Name == "" || len([]rune(in.Name)) > 120 || len(in.Email) > 254 || err != nil || parsed.Address != in.Email || len(in.Password) < 10 || len(in.Password) > 72 {
		platform.Error(c, 400, "INVALID_INPUT", "Provide a valid name, email, and password (10–72 bytes)")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	tx, err := a.db.Begin(c)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	defer tx.Rollback(c)
	var u account
	err = tx.QueryRow(c, `INSERT INTO users (name,email,role) VALUES ($1,$2,'CUSTOMER') RETURNING id::text,name,email,role`, in.Name, in.Email).Scan(&u.ID, &u.Name, &u.Email, &u.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
			platform.Error(c, 409, "EMAIL_TAKEN", "An account with this email already exists")
		} else {
			platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	var pinHash *string
	if in.PIN != "" {
		hash, hashErr := a.hashPIN(u.ID, in.PIN)
		if hashErr != nil {
			platform.Error(c, 500, "INTERNAL_ERROR", "Could not hash PIN")
			return
		}
		pinHash = &hash
	}
	if _, err = tx.Exec(c, `INSERT INTO user_credentials(user_id,password_hash,pin_hash) VALUES ($1,$2,$3)`, u.ID, string(hash), pinHash); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if pinHash != nil {
		if _, err = tx.Exec(c, `INSERT INTO pin_audit_events(actor_user_id,target_user_id,action) VALUES ($1,$1,'ENROLL')`, u.ID); err != nil {
			platform.Error(c, 500, "INTERNAL_ERROR", "Could not record PIN enrollment")
			return
		}
	}
	token, refresh, expires, err := a.tokens(u)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	digest := sha256.Sum256([]byte(refresh))
	if _, err = tx.Exec(c, `INSERT INTO refresh_sessions(user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, u.ID, hex.EncodeToString(digest[:]), expires); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if err = tx.Commit(c); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	a.respond(c, u, 201, token, refresh)
}

func (a *Auth) login(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if len(in.Email) > 254 || in.Password == "" || len(in.Password) > 72 {
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Email or password is incorrect")
		return
	}
	var u account
	var hash string
	err := a.db.QueryRow(c, `SELECT u.id::text,u.name,u.email,u.role,cr.password_hash FROM users u JOIN user_credentials cr ON cr.user_id=u.id WHERE u.email=$1 AND u.status='ACTIVE'`, in.Email).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &hash)
	if err == pgx.ErrNoRows {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(in.Password))
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Email or password is incorrect")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Email or password is incorrect")
		return
	}
	a.issuePasswordSession(c, u, hash)
}

func (a *Auth) issuePasswordSession(c *gin.Context, u account, passwordHash string) {
	tx, err := a.db.Begin(c.Request.Context())
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue session")
		return
	}
	defer tx.Rollback(c.Request.Context())
	var matches bool
	err = tx.QueryRow(c.Request.Context(), `SELECT cr.password_hash=$2,u.name,u.email,u.role FROM user_credentials cr JOIN users u ON u.id=cr.user_id WHERE cr.user_id=$1 AND u.status='ACTIVE' FOR UPDATE OF cr`, u.ID, passwordHash).Scan(&matches, &u.Name, &u.Email, &u.Role)
	if err == pgx.ErrNoRows {
		platform.Error(c, 401, "UNAUTHORIZED", "Account unavailable")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue session")
		return
	}
	if !matches {
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Credentials changed; sign in again")
		return
	}
	token, refresh, expires, err := a.tokens(u)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue session")
		return
	}
	digest := sha256.Sum256([]byte(refresh))
	if _, err = tx.Exec(c.Request.Context(), `INSERT INTO refresh_sessions(user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, u.ID, hex.EncodeToString(digest[:]), expires); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue session")
		return
	}
	if tx.Commit(c.Request.Context()) != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue session")
		return
	}
	a.respond(c, u, 200, token, refresh)
}

func (a *Auth) tokens(u account) (string, string, time.Time, error) {
	now := time.Now()
	token, err := authn.Sign(authn.Claims{Subject: u.ID, Email: u.Email, Role: u.Role, Issued: now.Unix(), Expires: now.Add(accessTTL).Unix()}, a.secret)
	if err != nil {
		return "", "", time.Time{}, err
	}
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", time.Time{}, err
	}
	return token, hex.EncodeToString(raw), now.Add(refreshTTL), nil
}

func (a *Auth) respond(c *gin.Context, u account, status int, token, refresh string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookie, refresh, int(refreshTTL.Seconds()), "/api/v1/auth", "", a.secureCookie, true)
	c.JSON(status, gin.H{"access_token": token, "token_type": "Bearer", "expires_in": int(accessTTL.Seconds()), "user": gin.H{"id": u.ID, "name": u.Name, "email": u.Email, "role": u.Role}})
}

func (a *Auth) refresh(c *gin.Context) {
	refresh, err := c.Cookie(refreshCookie)
	if err != nil {
		a.clearCookie(c)
		platform.Error(c, 401, "INVALID_SESSION", "Session is invalid or expired")
		return
	}
	digest := sha256.Sum256([]byte(refresh))
	hash := hex.EncodeToString(digest[:])
	tx, err := a.db.Begin(c)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	defer tx.Rollback(c)
	var u account
	var sessionID string
	var owner string
	err = tx.QueryRow(c, `SELECT cr.user_id::text FROM user_credentials cr JOIN refresh_sessions s ON s.user_id=cr.user_id WHERE s.token_hash=$1 FOR UPDATE OF cr`, hash).Scan(&owner)
	if err == pgx.ErrNoRows {
		a.clearCookie(c)
		platform.Error(c, 401, "INVALID_SESSION", "Session is invalid or expired")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not refresh session")
		return
	}
	err = tx.QueryRow(c, `SELECT s.id::text,u.id::text,u.name,u.email,u.role FROM refresh_sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>now() AND u.status='ACTIVE' FOR UPDATE OF s`, hash).Scan(&sessionID, &u.ID, &u.Name, &u.Email, &u.Role)
	if err == pgx.ErrNoRows {
		a.clearCookie(c)
		platform.Error(c, 401, "INVALID_SESSION", "Session is invalid or expired")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	token, nextRefresh, expires, err := a.tokens(u)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	digestNext := sha256.Sum256([]byte(nextRefresh))
	if _, err = tx.Exec(c, `UPDATE refresh_sessions SET revoked_at=now() WHERE id=$1`, sessionID); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if _, err = tx.Exec(c, `INSERT INTO refresh_sessions(user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, u.ID, hex.EncodeToString(digestNext[:]), expires); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if err = tx.Commit(c); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	a.respond(c, u, 200, token, nextRefresh)
}

func (a *Auth) logout(c *gin.Context) {
	if refresh, err := c.Cookie(refreshCookie); err == nil {
		digest := sha256.Sum256([]byte(refresh))
		if _, err = a.db.Exec(c, `UPDATE refresh_sessions SET revoked_at=now() WHERE token_hash=$1 AND revoked_at IS NULL`, hex.EncodeToString(digest[:])); err != nil {
			platform.Error(c, 500, "INTERNAL_ERROR", "Could not revoke session")
			return
		}
	}
	a.clearCookie(c)
	c.Status(204)
}

func (a *Auth) clearCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookie, "", -1, "/api/v1/auth", "", a.secureCookie, true)
}

func (a *Auth) me(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var u account
	err := a.db.QueryRow(c, `SELECT id::text,name,email,role FROM users WHERE id=$1 AND status='ACTIVE'`, claims.Subject).Scan(&u.ID, &u.Name, &u.Email, &u.Role)
	if err == pgx.ErrNoRows {
		platform.Error(c, 401, "INVALID_TOKEN", "Access token is invalid")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(200, gin.H{"user": gin.H{"id": u.ID, "name": u.Name, "email": u.Email, "role": u.Role}})
}

func (a *Auth) changePassword(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var in struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	if in.CurrentPassword == "" || len(in.CurrentPassword) > 72 || !validNewPassword(in.NewPassword) {
		platform.Error(c, 400, "INVALID_INPUT", "Current password is required; new password must be 10–72 bytes")
		return
	}
	var status, passwordHash string
	err := a.db.QueryRow(c, `SELECT u.status,cr.password_hash FROM users u JOIN user_credentials cr ON cr.user_id=u.id WHERE u.id=$1`, claims.Subject).Scan(&status, &passwordHash)
	if err == pgx.ErrNoRows {
		platform.Error(c, 401, "UNAUTHORIZED", "Account is unavailable")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if status != "ACTIVE" {
		platform.Error(c, 403, "ACCOUNT_DISABLED", "Account is disabled")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.CurrentPassword)) != nil {
		platform.Error(c, 401, "INVALID_CREDENTIALS", "Current password is incorrect")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.NewPassword)) == nil {
		platform.Error(c, 400, "PASSWORD_UNCHANGED", "New password must differ from current password")
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	tx, err := a.db.Begin(c)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	defer tx.Rollback(c)
	result, err := tx.Exec(c, `UPDATE user_credentials SET password_hash=$1,password_changed_at=now() WHERE user_id=$2 AND password_hash=$3`, string(newHash), claims.Subject, passwordHash)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not update password")
		return
	}
	if result.RowsAffected() == 0 {
		platform.Error(c, 409, "CREDENTIALS_CHANGED", "Credentials changed; retry using the current password")
		return
	}
	if _, err = tx.Exec(c, `UPDATE refresh_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, claims.Subject); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not revoke sessions")
		return
	}
	if _, err = tx.Exec(c, `INSERT INTO password_audit_events(actor_user_id,target_user_id,action) VALUES ($1,$1,'SELF_CHANGE')`, claims.Subject); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not record password change")
		return
	}
	if err = tx.Commit(c); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not update password")
		return
	}
	a.clearCookie(c)
	c.Status(204)
}

func (a *Auth) adminResetPassword(c *gin.Context) {
	claims, ok := a.claims(c)
	if !ok {
		return
	}
	var adminRole, adminStatus string
	err := a.db.QueryRow(c, `SELECT role,status FROM users WHERE id=$1`, claims.Subject).Scan(&adminRole, &adminStatus)
	if err == pgx.ErrNoRows {
		platform.Error(c, 401, "UNAUTHORIZED", "Account is unavailable")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if claims.Role != "ADMIN" || adminRole != "ADMIN" {
		platform.Error(c, 403, "FORBIDDEN", "Administrator role is required")
		return
	}
	if adminStatus != "ACTIVE" {
		platform.Error(c, 403, "ACCOUNT_DISABLED", "Administrator account is disabled")
		return
	}
	targetID := c.Param("user_id")
	if !uuidPattern.MatchString(targetID) {
		platform.Error(c, 400, "INVALID_USER_ID", "User ID must be a UUID")
		return
	}
	var in struct {
		NewPassword string `json:"new_password"`
	}
	if !bindAuthJSON(c, &in) {
		return
	}
	if !validNewPassword(in.NewPassword) {
		platform.Error(c, 400, "INVALID_INPUT", "New password must be 10–72 bytes")
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	tx, err := a.db.Begin(c)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		return
	}
	defer tx.Rollback(c)
	result, err := tx.Exec(c, `UPDATE user_credentials SET password_hash=$1,password_changed_at=now() WHERE user_id=$2`, string(newHash), targetID)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not reset password")
		return
	}
	if result.RowsAffected() == 0 {
		platform.Error(c, 404, "USER_NOT_FOUND", "User was not found")
		return
	}
	if _, err = tx.Exec(c, `UPDATE refresh_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, targetID); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not revoke user sessions")
		return
	}
	if _, err = tx.Exec(c, `INSERT INTO password_audit_events(actor_user_id,target_user_id,action) VALUES ($1,$2,'ADMIN_RESET')`, claims.Subject, targetID); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not record password reset")
		return
	}
	if err = tx.Commit(c); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not reset password")
		return
	}
	if targetID == claims.Subject {
		a.clearCookie(c)
	}
	c.Status(204)
}

func validNewPassword(password string) bool {
	return len(password) >= 10 && len(password) <= 72
}

func (a *Auth) claims(c *gin.Context) (authn.Claims, bool) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		platform.Error(c, 401, "UNAUTHORIZED", "A valid access token is required")
		return authn.Claims{}, false
	}
	claims, err := authn.Verify(parts[1], a.secret)
	if err != nil {
		platform.Error(c, 401, "UNAUTHORIZED", "A valid access token is required")
		return claims, false
	}
	return claims, true
}

func bindAuthJSON(c *gin.Context, dst any) bool {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		platform.Error(c, 415, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
		return false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAuthBody)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			platform.Error(c, 413, "PAYLOAD_TOO_LARGE", "Request body is too large")
		} else {
			platform.Error(c, 400, "INVALID_INPUT", "Invalid request body")
		}
		return false
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		platform.Error(c, 400, "INVALID_INPUT", "Request body must contain one JSON object")
		return false
	}
	return true
}
