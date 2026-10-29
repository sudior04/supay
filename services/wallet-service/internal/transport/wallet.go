package transport

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"supay.local/platform/internal/authn"
	"supay.local/platform/internal/platform"
)

type Wallets struct {
	db     *pgxpool.Pool
	secret []byte
}

func Register(r *gin.Engine, db *pgxpool.Pool) error {
	secret := []byte(platform.Env("JWT_SECRET", ""))
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	w := &Wallets{db: db, secret: secret}
	r.GET("/api/v1/wallets/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "wallet-service", "stage": "wallet-onboarding"})
	})
	r.GET("/api/v1/wallets/me", w.list)
	r.POST("/api/v1/wallets", w.create)
	r.POST("/api/v1/wallets/:wallet_id/cards", w.issueVirtualCard)
	return nil
}

type identity struct{ userID string }

func (w *Wallets) identity(c *gin.Context) (identity, bool) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		platform.Error(c, 401, "UNAUTHORIZED", "A valid access token is required")
		return identity{}, false
	}
	claims, err := authn.Verify(parts[1], w.secret)
	if err != nil {
		platform.Error(c, 401, "UNAUTHORIZED", "A valid access token is required")
		return identity{}, false
	}
	return identity{userID: claims.Subject}, true
}

func (w *Wallets) create(c *gin.Context) {
	who, ok := w.identity(c)
	if !ok {
		return
	}
	var id, status string
	err := w.db.QueryRow(c, `INSERT INTO wallet_accounts (owner_id) VALUES ($1) ON CONFLICT (owner_id,currency,account_type) DO NOTHING RETURNING id::text,status`, who.userID).Scan(&id, &status)
	created := err == nil
	if err == pgx.ErrNoRows {
		err = w.db.QueryRow(c, `SELECT id::text,status FROM wallet_accounts WHERE owner_id=$1 AND currency='VND' AND account_type='PERSONAL'`, who.userID).Scan(&id, &status)
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not create wallet")
		return
	}
	code := 200
	if created {
		code = 201
	}
	c.JSON(code, gin.H{"wallet": gin.H{"id": id, "currency": "VND", "balance_minor": 0, "status": status}})
}

type walletView struct {
	ID            string `json:"id"`
	Currency      string `json:"currency"`
	BalanceMinor  int64  `json:"balance_minor"`
	Status        string `json:"status"`
	CardReference string `json:"card_reference,omitempty"`
	CardStatus    string `json:"card_status,omitempty"`
}

func (w *Wallets) list(c *gin.Context) {
	who, ok := w.identity(c)
	if !ok {
		return
	}
	rows, err := w.db.Query(c, `SELECT wa.id::text,wa.currency,wa.balance_minor,wa.status,COALESCE(vc.public_reference,''),COALESCE(vc.status,'') FROM wallet_accounts wa LEFT JOIN virtual_cards vc ON vc.wallet_id=wa.id WHERE wa.owner_id=$1 ORDER BY wa.created_at`, who.userID)
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not load wallets")
		return
	}
	defer rows.Close()
	items := make([]walletView, 0)
	for rows.Next() {
		var v walletView
		if err = rows.Scan(&v.ID, &v.Currency, &v.BalanceMinor, &v.Status, &v.CardReference, &v.CardStatus); err != nil {
			platform.Error(c, 500, "INTERNAL_ERROR", "Could not load wallets")
			return
		}
		items = append(items, v)
	}
	if rows.Err() != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not load wallets")
		return
	}
	c.JSON(200, gin.H{"wallets": items})
}

func (w *Wallets) issueVirtualCard(c *gin.Context) {
	who, ok := w.identity(c)
	if !ok {
		return
	}
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue virtual card")
		return
	}
	reference := "SUPAY-" + strings.ToUpper(hex.EncodeToString(bytes))
	var id, status string
	err := w.db.QueryRow(c, `INSERT INTO virtual_cards(wallet_id,public_reference) SELECT wa.id,$1 FROM wallet_accounts wa WHERE wa.id=$2 AND wa.owner_id=$3 AND wa.status='ACTIVE' ON CONFLICT (wallet_id) DO NOTHING RETURNING id::text,status`, reference, c.Param("wallet_id"), who.userID).Scan(&id, &status)
	created := err == nil
	if err == pgx.ErrNoRows {
		err = w.db.QueryRow(c, `SELECT vc.id::text,vc.public_reference,vc.status FROM virtual_cards vc JOIN wallet_accounts wa ON wa.id=vc.wallet_id WHERE vc.wallet_id=$1 AND wa.owner_id=$2`, c.Param("wallet_id"), who.userID).Scan(&id, &reference, &status)
	}
	if err == pgx.ErrNoRows {
		platform.Error(c, 404, "WALLET_NOT_FOUND", "Wallet was not found")
		return
	}
	if err != nil {
		platform.Error(c, 500, "INTERNAL_ERROR", "Could not issue virtual card")
		return
	}
	code := 200
	if created {
		code = 201
	}
	c.JSON(code, gin.H{"card": gin.H{"id": id, "type": "VIRTUAL", "reference": reference, "status": status, "usable_for_payments": false}})
}
