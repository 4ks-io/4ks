package fetcher

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

const (
	headerTimestamp = "X-4ks-Auth-Timestamp"
	headerNonce     = "X-4ks-Auth-Nonce"
	headerBodyHash  = "X-4ks-Auth-Body-SHA256"
	headerSignature = "X-4ks-Auth-Signature"
)

type signatureHeaders struct {
	Timestamp string
	Nonce     string
	BodyHash  string
	Signature string
}

func newNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func sign(secret []byte, method, host, path, bodyHash, timestamp, nonce string) string {
	payload := strings.Join([]string{
		strings.ToUpper(method),
		strings.ToLower(host),
		path,
		strings.ToLower(bodyHash),
		timestamp,
		nonce,
	}, "\n")

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func buildSignatureHeaders(secret []byte, method, host, path string, body []byte, now time.Time, nonce string) signatureHeaders {
	timestamp := now.UTC().Format(time.RFC3339)
	bodyHash := hashBody(body)
	return signatureHeaders{
		Timestamp: timestamp,
		Nonce:     nonce,
		BodyHash:  bodyHash,
		Signature: sign(secret, method, host, path, bodyHash, timestamp, nonce),
	}
}

func applySignatureHeaders(req *http.Request, headers signatureHeaders) {
	req.Header.Set(headerTimestamp, headers.Timestamp)
	req.Header.Set(headerNonce, headers.Nonce)
	req.Header.Set(headerBodyHash, headers.BodyHash)
	req.Header.Set(headerSignature, headers.Signature)
}
