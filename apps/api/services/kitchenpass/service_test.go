package kitchenpasssvc

import (
	"4ks/libs/go/models"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type memoryStore struct {
	records map[string]*models.PersonalAccessToken
}

func (s *memoryStore) GetByUserID(_ context.Context, userID string) (*models.PersonalAccessToken, error) {
	record, ok := s.records[userID]
	if !ok {
		return nil, ErrKitchenPassNotFound
	}
	copy := *record
	return &copy, nil
}

func (s *memoryStore) Upsert(_ context.Context, record *models.PersonalAccessToken) error {
	copy := *record
	s.records[record.UserID] = &copy
	return nil
}

func (s *memoryStore) FindByDigest(_ context.Context, digest string) (*models.PersonalAccessToken, error) {
	for _, record := range s.records {
		if record.TokenDigest == digest {
			copy := *record
			return &copy, nil
		}
	}
	return nil, ErrKitchenPassNotFound
}

func TestKitchenPassLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	store := &memoryStore{records: map[string]*models.PersonalAccessToken{}}
	service := &service{
		baseURL:          "https://www.4ks.io",
		digestSecret:     []byte("01234567890123456789012345678901"),
		encryptionSecret: []byte("abcdefghijklmnopqrstuvwxyz012345"),
		store:            store,
		now:              func() time.Time { return now },
	}

	created, err := service.CreateOrRotate(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("CreateOrRotate returned error: %v", err)
	}
	if !created.Enabled || created.SkillURL == nil || created.CopyText == nil {
		t.Fatalf("expected enabled response, got %+v", created)
	}

	record := store.records["user-1"]
	if record == nil {
		t.Fatal("expected record to be stored")
	}
	if record.TokenDigest == "" || record.EncryptedToken == "" {
		t.Fatalf("expected token digest and ciphertext to be stored, got %+v", record)
	}
	if record.EncryptedToken == *created.SkillURL {
		t.Fatal("expected encrypted token storage, not plaintext skill URL")
	}
	if *created.SkillURL != "https://www.4ks.io/skill.md" {
		t.Fatalf("unexpected skill URL %q", *created.SkillURL)
	}
	if !strings.Contains(*created.CopyText, "https://www.4ks.io/skill.md") {
		t.Fatalf("expected copy text to include static skill URL, got %q", *created.CopyText)
	}

	token := strings.TrimPrefix(strings.Split(strings.Split(*created.CopyText, "Authorization: Bearer ")[1], "\n")[0], "Bearer ")
	validated, err := service.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if validated.UserID != "user-1" {
		t.Fatalf("expected validated owner user-1, got %q", validated.UserID)
	}

	rotatedAt := now.Add(5 * time.Minute)
	service.now = func() time.Time { return rotatedAt }
	rotated, err := service.CreateOrRotate(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("rotate returned error: %v", err)
	}
	if rotated.SkillURL == nil || *rotated.SkillURL != "https://www.4ks.io/skill.md" {
		t.Fatalf("expected static skill URL after rotation, got %v", rotated.SkillURL)
	}
	if rotated.CopyText == nil || *rotated.CopyText == *created.CopyText {
		t.Fatalf("expected rotation to produce new tokenized copy text")
	}

	if _, err := service.ValidateToken(context.Background(), token); !errors.Is(err, ErrKitchenPassNotFound) {
		t.Fatalf("expected old token to be invalidated, got %v", err)
	}

	service.now = func() time.Time { return rotatedAt.Add(5 * time.Minute) }
	if err := service.Revoke(context.Background(), "user-1"); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}

	status, err := service.GetStatus(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if status.Enabled {
		t.Fatalf("expected disabled response after revoke, got %+v", status)
	}

	newToken := strings.TrimPrefix(strings.Split(strings.Split(*rotated.CopyText, "Authorization: Bearer ")[1], "\n")[0], "Bearer ")
	if _, err := service.ValidateToken(context.Background(), newToken); !errors.Is(err, ErrKitchenPassNotFound) {
		t.Fatalf("expected revoked token to be rejected, got %v", err)
	}
}

func TestIsKitchenPassToken(t *testing.T) {
	t.Parallel()

	if !IsKitchenPassToken("4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789") {
		t.Fatal("expected prefixed token to be accepted")
	}
	if IsKitchenPassToken("not-a-kitchen-pass") {
		t.Fatal("expected unrelated token to be rejected")
	}
}
