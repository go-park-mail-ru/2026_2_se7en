package csrf

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

const MinSecretLength = 32

type HashToken struct {
	secret []byte
}

func NewHMACHashToken(secret string) (*HashToken, error) {
	if len(secret) < MinSecretLength {
		return nil, errors.New("CSRF_SECRET must contain at least 32 bytes")
	}
	return &HashToken{secret: []byte(secret)}, nil
}

func (t *HashToken) Create(session *models.Session, expiresAt int64) string {
	return hex.EncodeToString(t.signature(session, expiresAt)) + ":" + strconv.FormatInt(expiresAt, 10)
}

func (t *HashToken) Check(session *models.Session, token string) error {
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return errors.New("invalid CSRF token format")
	}
	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return errors.New("invalid CSRF token expiration")
	}
	if expiresAt <= time.Now().Unix() {
		return errors.New("CSRF token expired")
	}
	signature, err := hex.DecodeString(parts[0])
	if err != nil || len(signature) != sha256.Size || !hmac.Equal(signature, t.signature(session, expiresAt)) {
		return errors.New("invalid CSRF token signature")
	}
	return nil
}

func (t *HashToken) signature(session *models.Session, expiresAt int64) []byte {
	mac := hmac.New(sha256.New, t.secret)
	data := fmt.Sprintf("%s:%s:%d", session.ID, session.UserID, expiresAt)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}
