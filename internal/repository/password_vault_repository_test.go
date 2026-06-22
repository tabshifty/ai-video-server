package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestNormalizePasswordVaultInputPreservesPasswordSpacing(t *testing.T) {
	t.Parallel()

	got := normalizePasswordVaultInput(models.AdminPasswordVaultInput{
		Name:     "  家用   NAS  ",
		Account:  "  admin  ",
		Password: "  secret  ",
		URL:      "  https://nas.local  ",
		Note:     "  备注  ",
	})

	if got.Name != "家用 NAS" {
		t.Fatalf("unexpected normalized name: %q", got.Name)
	}
	if got.Account != "admin" {
		t.Fatalf("unexpected normalized account: %q", got.Account)
	}
	if got.Password != "  secret  " {
		t.Fatalf("unexpected normalized password: %q", got.Password)
	}
	if got.URL != "https://nas.local" {
		t.Fatalf("unexpected normalized url: %q", got.URL)
	}
	if got.Note != "备注" {
		t.Fatalf("unexpected normalized note: %q", got.Note)
	}
}

func TestScanPasswordVaultEntryReadsStoredValues(t *testing.T) {
	t.Parallel()

	entryID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	now := time.Unix(1710000000, 0).UTC()

	got, ciphertext, err := scanPasswordVaultEntry(stubRowScanner{
		values: []any{
			entryID,
			"家庭 NAS",
			"admin",
			"ciphertext",
			"https://nas.local",
			"备忘",
			now,
			now,
		},
	})
	if err != nil {
		t.Fatalf("scanPasswordVaultEntry returned error: %v", err)
	}

	if got.ID != entryID {
		t.Fatalf("unexpected id: %s", got.ID)
	}
	if got.Name != "家庭 NAS" || got.Account != "admin" || got.URL != "https://nas.local" || got.Note != "备忘" {
		t.Fatalf("unexpected entry: %+v", got)
	}
	if ciphertext != "ciphertext" {
		t.Fatalf("unexpected ciphertext: %q", ciphertext)
	}
}
