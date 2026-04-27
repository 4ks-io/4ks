package kitchenpasssvc

import (
	"4ks/apps/api/dtos"
	"4ks/libs/go/models"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const (
	tokenPrefix      = "4ks_pass_"
	tokenEntropySize = 32
	minTokenLength   = len(tokenPrefix) + 20
)

var (
	ErrKitchenPassNotFound     = errors.New("kitchen pass not found")
	ErrInvalidKitchenPassToken = errors.New("invalid kitchen pass token")
)

type Service interface {
	GetStatus(context.Context, string) (*dtos.KitchenPassResponse, error)
	CreateOrRotate(context.Context, string) (*dtos.KitchenPassResponse, error)
	Revoke(context.Context, string) error
	ValidateToken(context.Context, string) (*models.PersonalAccessToken, error)
}

type Config struct {
	BaseURL          string
	DigestSecret     string
	EncryptionSecret string
}

type tokenStore interface {
	GetByUserID(context.Context, string) (*models.PersonalAccessToken, error)
	Upsert(context.Context, *models.PersonalAccessToken) error
	FindByDigest(context.Context, string) (*models.PersonalAccessToken, error)
}

type service struct {
	baseURL          string
	digestSecret     []byte
	encryptionSecret []byte
	store            tokenStore
	now              func() time.Time
}

type firestoreStore struct {
	collection *firestore.CollectionRef
}

func New(store *firestore.Client, cfg Config) Service {
	return &service{
		baseURL:          strings.TrimRight(cfg.BaseURL, "/"),
		digestSecret:     []byte(cfg.DigestSecret),
		encryptionSecret: []byte(cfg.EncryptionSecret),
		store: firestoreStore{
			collection: store.Collection("personal_access_tokens"),
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *service) GetStatus(ctx context.Context, userID string) (*dtos.KitchenPassResponse, error) {
	record, err := s.store.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrKitchenPassNotFound) {
			return disabledKitchenPassResponse(), nil
		}
		return nil, err
	}
	if record.RevokedDate != nil {
		return disabledKitchenPassResponse(), nil
	}

	token, err := s.decryptToken(record.EncryptedToken)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(record, token), nil
}

func (s *service) CreateOrRotate(ctx context.Context, userID string) (*dtos.KitchenPassResponse, error) {
	now := s.now()
	existing, err := s.store.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, ErrKitchenPassNotFound) {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	encryptedToken, err := s.encryptToken(token)
	if err != nil {
		return nil, err
	}

	record := &models.PersonalAccessToken{
		UserID:         userID,
		TokenDigest:    s.digestToken(token),
		EncryptedToken: encryptedToken,
		TokenPreview:   previewToken(token),
		CreatedDate:    now,
	}

	if existing != nil {
		record.RotatedDate = &now
		record.LastUsedDate = existing.LastUsedDate
		record.LastUsedAction = existing.LastUsedAction
	}

	if err := s.store.Upsert(ctx, record); err != nil {
		return nil, err
	}

	return s.buildResponse(record, token), nil
}

func (s *service) Revoke(ctx context.Context, userID string) error {
	record, err := s.store.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrKitchenPassNotFound) {
			return nil
		}
		return err
	}

	now := s.now()
	record.RevokedDate = &now
	record.EncryptedToken = ""
	record.TokenPreview = ""

	return s.store.Upsert(ctx, record)
}

func (s *service) ValidateToken(ctx context.Context, token string) (*models.PersonalAccessToken, error) {
	if !IsKitchenPassToken(token) {
		return nil, ErrInvalidKitchenPassToken
	}

	record, err := s.store.FindByDigest(ctx, s.digestToken(token))
	if err != nil {
		return nil, err
	}
	if record.RevokedDate != nil {
		return nil, ErrKitchenPassNotFound
	}

	return record, nil
}

func IsKitchenPassToken(token string) bool {
	return strings.HasPrefix(token, tokenPrefix) && len(token) >= minTokenLength
}

func disabledKitchenPassResponse() *dtos.KitchenPassResponse {
	return &dtos.KitchenPassResponse{
		Enabled:        false,
		SkillURL:       nil,
		CopyText:       nil,
		CreatedDate:    nil,
		LastUsedDate:   nil,
		LastUsedAction: nil,
	}
}

func (s *service) buildResponse(record *models.PersonalAccessToken, token string) *dtos.KitchenPassResponse {
	skillURL := s.baseURL + "/skill.md"
	copyText := fmt.Sprintf(
		"Use this as my 4ks recipe memory:\n\n%s\n\nSecret Authentication header:\nAuthorization: Bearer %s\n\nBefore saving or changing a recipe, search my 4ks recipes first to avoid duplicates.",
		skillURL,
		token,
	)

	return &dtos.KitchenPassResponse{
		Enabled:        true,
		SkillURL:       &skillURL,
		CopyText:       &copyText,
		CreatedDate:    &record.CreatedDate,
		LastUsedDate:   record.LastUsedDate,
		LastUsedAction: record.LastUsedAction,
	}
}

func generateToken() (string, error) {
	buf := make([]byte, tokenEntropySize)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}

	return tokenPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

func previewToken(token string) string {
	if len(token) <= 16 {
		return token
	}

	return token[:12] + "..." + token[len(token)-4:]
}

func (s *service) digestToken(token string) string {
	mac := hmac.New(sha256.New, s.digestSecret)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *service) encryptToken(token string) (string, error) {
	block, err := aes.NewCipher(deriveEncryptionKey(s.encryptionSecret))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (s *service) decryptToken(encrypted string) (string, error) {
	payload, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(deriveEncryptionKey(s.encryptionSecret))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return "", ErrInvalidKitchenPassToken
	}

	nonce, ciphertext := payload[:nonceSize], payload[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func deriveEncryptionKey(secret []byte) []byte {
	sum := sha256.Sum256(secret)
	return sum[:]
}

func (s firestoreStore) GetByUserID(ctx context.Context, userID string) (*models.PersonalAccessToken, error) {
	doc, err := s.collection.Doc(userID).Get(ctx)
	if err != nil {
		return nil, ErrKitchenPassNotFound
	}

	var record models.PersonalAccessToken
	if err := doc.DataTo(&record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (s firestoreStore) Upsert(ctx context.Context, record *models.PersonalAccessToken) error {
	_, err := s.collection.Doc(record.UserID).Set(ctx, record)
	return err
}

func (s firestoreStore) FindByDigest(ctx context.Context, digest string) (*models.PersonalAccessToken, error) {
	iter := s.collection.Where("tokenDigest", "==", digest).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return nil, ErrKitchenPassNotFound
		}
		return nil, err
	}

	var record models.PersonalAccessToken
	if err := doc.DataTo(&record); err != nil {
		return nil, err
	}

	return &record, nil
}
