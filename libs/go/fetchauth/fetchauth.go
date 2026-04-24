package fetchauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	HeaderTimestamp = "X-4ks-Auth-Timestamp"
	HeaderNonce     = "X-4ks-Auth-Nonce"
	HeaderBodyHash  = "X-4ks-Auth-Body-SHA256"
	HeaderSignature = "X-4ks-Auth-Signature"
)

type SignatureHeaders struct {
	Timestamp string
	Nonce     string
	BodyHash  string
	Signature string
}

func NewNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func HashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func Sign(secret []byte, method, host, path, bodyHash, timestamp, nonce string) string {
	payload := canonicalPayload(method, host, path, bodyHash, timestamp, nonce)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func BuildHeaders(secret []byte, method, host, path string, body []byte, now time.Time, nonce string) SignatureHeaders {
	timestamp := now.UTC().Format(time.RFC3339)
	bodyHash := HashBody(body)
	signature := Sign(secret, method, host, path, bodyHash, timestamp, nonce)

	return SignatureHeaders{
		Timestamp: timestamp,
		Nonce:     nonce,
		BodyHash:  bodyHash,
		Signature: signature,
	}
}

func ApplyHeaders(req *http.Request, headers SignatureHeaders) {
	req.Header.Set(HeaderTimestamp, headers.Timestamp)
	req.Header.Set(HeaderNonce, headers.Nonce)
	req.Header.Set(HeaderBodyHash, headers.BodyHash)
	req.Header.Set(HeaderSignature, headers.Signature)
}

func Verify(secret []byte, method, host, path, bodyHash, timestamp, nonce, signature string) error {
	expected := Sign(secret, method, host, path, bodyHash, timestamp, nonce)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(strings.ToLower(signature))) != 1 {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func canonicalPayload(method, host, path, bodyHash, timestamp, nonce string) string {
	return strings.Join([]string{
		strings.ToUpper(method),
		strings.ToLower(host),
		path,
		strings.ToLower(bodyHash),
		timestamp,
		nonce,
	}, "\n")
}
