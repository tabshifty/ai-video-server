package models

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	AppClientTypeAndroidTV    = "android_tv"
	AppClientTypeAndroidPhone = "android_phone"

	TVAppPackageName = "com.chee.videos.tv"
	AppPackageName   = "com.chee.videos"

	TVReleaseStatusDraft             = "draft"
	TVReleaseStatusPublishedComplete = "published_complete"
	TVReleaseStatusPublishedMissing  = "published_missing_abi"
	TVReleaseStatusOffline           = "offline"

	TVABIArmV7 = "armeabi-v7a"
	TVABIArm64 = "arm64-v8a"

	AppAPKSlotSingle = "single"
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
	TVAPKErrorClientTypeMismatch    = "client_type_mismatch"
	TVAPKErrorVersionNameConflict   = "version_name_conflict"
)

func NewTVAPKDomainError(code, message string) error {
	return &TVAPKDomainError{Code: code, Message: message}
}

type TVAppAPKParsedMetadata struct {
	ClientType   string          `json:"client_type"`
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
	ClientType          string
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
	ClientType        string                  `json:"client_type"`
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

func NormalizeAppClientType(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, "-", "_")
	switch value {
	case AppClientTypeAndroidTV, AppClientTypeAndroidPhone:
		return value
	default:
		return ""
	}
}

func AppClientTypeSlug(clientType string) string {
	normalized := NormalizeAppClientType(clientType)
	if normalized == "" {
		return ""
	}
	return strings.ReplaceAll(normalized, "_", "-")
}

func AppClientTypeDisplayName(clientType string) string {
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		return "Android 手机"
	case AppClientTypeAndroidTV:
		return "Android TV"
	default:
		return ""
	}
}

func AppClientTypeShortLabel(clientType string) string {
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		return "手机端"
	case AppClientTypeAndroidTV:
		return "TV 端"
	default:
		return ""
	}
}

func AppPackageNameForClientType(clientType string) string {
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		return AppPackageName
	case AppClientTypeAndroidTV:
		return TVAppPackageName
	default:
		return ""
	}
}

func DetectAppClientTypeByPackageName(packageName string) string {
	switch strings.TrimSpace(packageName) {
	case AppPackageName:
		return AppClientTypeAndroidPhone
	case TVAppPackageName:
		return AppClientTypeAndroidTV
	default:
		return ""
	}
}

func AppClientTypeSupportsABI(clientType string) bool {
	return NormalizeAppClientType(clientType) == AppClientTypeAndroidTV
}

func NormalizeReleaseArtifactSlot(clientType, raw string) string {
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		value := strings.TrimSpace(raw)
		if value == "" || strings.EqualFold(value, AppAPKSlotSingle) {
			return AppAPKSlotSingle
		}
		return ""
	case AppClientTypeAndroidTV:
		return TVNormalizeABI(raw)
	default:
		return ""
	}
}

func ReleaseArtifactComplete(clientType string, uploaded []string) bool {
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		for _, item := range uploaded {
			if NormalizeReleaseArtifactSlot(clientType, item) != "" {
				return true
			}
		}
		return false
	case AppClientTypeAndroidTV:
		return len(TVMissingABIs(TVUploadedABIs(uploaded))) == 0
	default:
		return false
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

func ReleaseStatusForVisibility(clientType string, visible bool, uploaded []string) string {
	if !visible {
		return TVReleaseStatusOffline
	}
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		return TVReleaseStatusPublishedComplete
	case AppClientTypeAndroidTV:
		return TVReleaseStatusForVisibility(true, uploaded)
	default:
		return TVReleaseStatusOffline
	}
}

func ReleaseVisibilityLabel(clientType, status string) string {
	switch NormalizeAppClientType(clientType) {
	case AppClientTypeAndroidPhone:
		switch strings.TrimSpace(status) {
		case TVReleaseStatusPublishedComplete:
			return "已发布"
		case TVReleaseStatusDraft:
			return "草稿"
		case TVReleaseStatusOffline:
			return "已下线"
		default:
			return strings.TrimSpace(status)
		}
	default:
		switch strings.TrimSpace(status) {
		case TVReleaseStatusPublishedComplete:
			return "已发布-完整"
		case TVReleaseStatusPublishedMissing:
			return "已发布-缺少 ABI"
		case TVReleaseStatusDraft:
			return "草稿"
		case TVReleaseStatusOffline:
			return "已下线"
		default:
			return strings.TrimSpace(status)
		}
	}
}
