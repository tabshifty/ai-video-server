package models

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	TVAppPackageName = "com.chee.videos.tv"

	TVReleaseStatusDraft             = "draft"
	TVReleaseStatusPublishedComplete = "published_complete"
	TVReleaseStatusPublishedMissing  = "published_missing_abi"
	TVReleaseStatusOffline           = "offline"

	TVABIArmV7 = "armeabi-v7a"
	TVABIArm64 = "arm64-v8a"
)

var TVSupportedABIs = []string{TVABIArm64, TVABIArmV7}

type TVAPKDomainError struct {
	Code    string
	Message string
}

func (e *TVAPKDomainError) Error() string {
	return e.Message
}

const (
	TVAPKErrorInvalidPackage        = "invalid_package"
	TVAPKErrorUnsupportedABI        = "unsupported_abi"
	TVAPKErrorDebugAPKRejected      = "debug_apk_rejected"
	TVAPKErrorInvalidVersion        = "invalid_version"
	TVAPKErrorAPKStructureBroken    = "apk_structure_broken"
	TVAPKErrorABIAlreadyExists      = "abi_already_exists"
	TVAPKErrorReplaceNeedsOffline   = "replace_requires_offline"
	TVAPKErrorReleaseNotPublishable = "release_not_publishable"
)

func NewTVAPKDomainError(code, message string) error {
	return &TVAPKDomainError{Code: code, Message: message}
}

type TVAppAPKParsedMetadata struct {
	PackageName  string          `json:"package_name"`
	VersionCode  int64           `json:"version_code"`
	VersionName  string          `json:"version_name"`
	ABI          string          `json:"abi"`
	IsDebuggable bool            `json:"is_debuggable"`
	FileName     string          `json:"file_name"`
	FileSize     int64           `json:"file_size"`
	SHA256       string          `json:"sha256"`
	MIMEType     string          `json:"mime_type"`
	RawManifest  json.RawMessage `json:"raw_manifest,omitempty"`
	ParsedAt     time.Time       `json:"parsed_at"`
}

type TVAppReleaseABIInfo struct {
	ID           int64
	ReleaseID    int64
	ABI          string
	FileName     string
	StoredPath   string
	FileSize     int64
	MIMEType     string
	SHA256       string
	IsDebuggable bool
	UploadUserID *uuid.UUID
	UploadUser   string
	UploadedAt   time.Time
	ReplacedAt   *time.Time
	Metadata     json.RawMessage
}

type TVAppReleaseRecord struct {
	ID                  int64
	PackageName         string
	VersionCode         int64
	VersionName         string
	ReleaseNotes        string
	Remarks             string
	PublishStatus       string
	PublishedAt         *time.Time
	LastStatusChangedAt time.Time
	LatestRecommended   bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ABIItems            []TVAppReleaseABIInfo
}

type TVAppFamilyReleaseABI struct {
	ABI       string    `json:"abi"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MIMEType  string    `json:"mime_type"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TVAppFamilyRelease struct {
	ID                int64                   `json:"id"`
	PackageName       string                  `json:"package_name"`
	VersionCode       int64                   `json:"version_code"`
	VersionName       string                  `json:"version_name"`
	ReleaseNotes      string                  `json:"release_notes"`
	PublishedAt       *time.Time              `json:"published_at,omitempty"`
	LatestRecommended bool                    `json:"latest_recommended"`
	UploadedABIs      []string                `json:"uploaded_abis"`
	MissingABIs       []string                `json:"missing_abis"`
	ABIItems          []TVAppFamilyReleaseABI `json:"abi_items"`
}

func TVNormalizeABI(raw string) string {
	value := strings.TrimSpace(raw)
	switch value {
	case TVABIArm64, TVABIArmV7:
		return value
	default:
		return ""
	}
}

func TVReleaseVisibleToFamily(status string) bool {
	switch strings.TrimSpace(status) {
	case TVReleaseStatusPublishedComplete, TVReleaseStatusPublishedMissing:
		return true
	default:
		return false
	}
}

func TVMissingABIs(uploaded []string) []string {
	if len(uploaded) == 0 {
		return append([]string(nil), TVSupportedABIs...)
	}
	seen := map[string]struct{}{}
	for _, abi := range uploaded {
		if normalized := TVNormalizeABI(abi); normalized != "" {
			seen[normalized] = struct{}{}
		}
	}
	out := make([]string, 0, len(TVSupportedABIs))
	for _, abi := range TVSupportedABIs {
		if _, ok := seen[abi]; ok {
			continue
		}
		out = append(out, abi)
	}
	return out
}

func TVUploadedABIs(uploaded []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(uploaded))
	for _, abi := range uploaded {
		normalized := TVNormalizeABI(abi)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	slices.Sort(out)
	return out
}

func TVReleaseStatusForVisibility(visible bool, uploaded []string) string {
	if !visible {
		return TVReleaseStatusOffline
	}
	if len(TVMissingABIs(TVUploadedABIs(uploaded))) == 0 {
		return TVReleaseStatusPublishedComplete
	}
	return TVReleaseStatusPublishedMissing
}
