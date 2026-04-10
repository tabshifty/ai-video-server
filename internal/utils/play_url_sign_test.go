package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestSignAndVerifyVideoSource(t *testing.T) {
	secret := "test-secret"
	videoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	exp := int64(1710000000)

	sig := SignVideoSource(secret, videoID, exp)
	if sig == "" {
		t.Fatalf("expected non-empty signature")
	}
	if !VerifyVideoSourceSign(secret, videoID, exp, sig) {
		t.Fatalf("expected signature verification success")
	}
}

func TestVerifyVideoSourceSignRejectsTamperedValues(t *testing.T) {
	secret := "test-secret"
	videoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	exp := int64(1710000000)

	sig := SignVideoSource(secret, videoID, exp)
	if VerifyVideoSourceSign(secret, uuid.MustParse("ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"), exp, sig) {
		t.Fatalf("expected signature mismatch on video id")
	}
	if VerifyVideoSourceSign(secret, videoID, exp+1, sig) {
		t.Fatalf("expected signature mismatch on exp")
	}
	if VerifyVideoSourceSign(secret, videoID, exp, sig+"00") {
		t.Fatalf("expected signature mismatch on tampered sig")
	}
}
