package services

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"video-server/internal/models"
)

const (
	TVAPKErrorInvalidPackage     = models.TVAPKErrorInvalidPackage
	TVAPKErrorUnsupportedABI     = models.TVAPKErrorUnsupportedABI
	TVAPKErrorDebugAPKRejected   = models.TVAPKErrorDebugAPKRejected
	TVAPKErrorInvalidVersion     = models.TVAPKErrorInvalidVersion
	TVAPKErrorAPKStructureBroken = models.TVAPKErrorAPKStructureBroken
)

func NewTVAPKDomainError(code, message string) error {
	return models.NewTVAPKDomainError(code, message)
}

func TVReleaseMissingABIs(items []models.AdminTvAppReleaseABIItem) []string {
	uploaded := make([]string, 0, len(items))
	for _, item := range items {
		uploaded = append(uploaded, item.ABI)
	}
	return models.TVMissingABIs(uploaded)
}

func TVReleaseUploadedABIs(items []models.AdminTvAppReleaseABIItem) []string {
	uploaded := make([]string, 0, len(items))
	for _, item := range items {
		uploaded = append(uploaded, item.ABI)
	}
	return models.TVUploadedABIs(uploaded)
}

func TVReleaseABIComplete(items []models.AdminTvAppReleaseABIItem) bool {
	return len(TVReleaseMissingABIs(items)) == 0
}

func NormalizeTVReleaseStatus(visible bool, items []models.AdminTvAppReleaseABIItem) string {
	return models.TVReleaseStatusForVisibility(visible, TVReleaseUploadedABIs(items))
}

func ParseTVAPKMetadata(apkPath string, fileName string) (models.TVAppAPKParsedMetadata, error) {
	info, err := os.Stat(apkPath)
	if err != nil {
		return models.TVAppAPKParsedMetadata{}, fmt.Errorf("stat apk: %w", err)
	}
	if info.Size() <= 0 {
		return models.TVAppAPKParsedMetadata{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "APK 文件为空")
	}

	reader, err := zip.OpenReader(apkPath)
	if err != nil {
		return models.TVAppAPKParsedMetadata{}, fmt.Errorf("open apk zip: %w", err)
	}
	defer reader.Close()

	abi, err := detectABIFromAPK(&reader.Reader)
	if err != nil {
		return models.TVAppAPKParsedMetadata{}, err
	}

	manifest, err := parseBinaryManifest(&reader.Reader)
	if err != nil {
		return models.TVAppAPKParsedMetadata{}, err
	}

	hashValue, err := sha256File(apkPath)
	if err != nil {
		return models.TVAppAPKParsedMetadata{}, fmt.Errorf("calculate apk sha256: %w", err)
	}

	rawManifest, err := json.Marshal(manifest)
	if err != nil {
		return models.TVAppAPKParsedMetadata{}, fmt.Errorf("marshal parsed manifest: %w", err)
	}

	meta := models.TVAppAPKParsedMetadata{
		PackageName:  strings.TrimSpace(manifest.PackageName),
		VersionCode:  manifest.VersionCode,
		VersionName:  strings.TrimSpace(manifest.VersionName),
		ABI:          abi,
		IsDebuggable: manifest.Debuggable,
		FileName:     filepath.Base(strings.TrimSpace(fileName)),
		FileSize:     info.Size(),
		SHA256:       hashValue,
		MIMEType:     "application/vnd.android.package-archive",
		RawManifest:  rawManifest,
		ParsedAt:     time.Now().UTC(),
	}

	if meta.PackageName != models.TVAppPackageName {
		return models.TVAppAPKParsedMetadata{}, NewTVAPKDomainError(TVAPKErrorInvalidPackage, "仅允许上传 com.chee.videos.tv 安装包")
	}
	if meta.VersionCode <= 0 || strings.TrimSpace(meta.VersionName) == "" {
		return models.TVAppAPKParsedMetadata{}, NewTVAPKDomainError(TVAPKErrorInvalidVersion, "无法从 APK 解析出有效版本号")
	}
	if meta.IsDebuggable {
		return models.TVAppAPKParsedMetadata{}, NewTVAPKDomainError(TVAPKErrorDebugAPKRejected, "仅允许上传 Release APK")
	}
	return meta, nil
}

func detectABIFromAPK(reader *zip.Reader) (string, error) {
	abis := map[string]struct{}{}
	for _, file := range reader.File {
		name := strings.TrimSpace(file.Name)
		if !strings.HasPrefix(name, "lib/") {
			continue
		}
		parts := strings.Split(name, "/")
		if len(parts) < 3 {
			continue
		}
		if abi := models.TVNormalizeABI(parts[1]); abi != "" {
			abis[abi] = struct{}{}
		}
	}
	if len(abis) == 1 {
		for abi := range abis {
			return abi, nil
		}
	}
	if len(abis) == 0 {
		return "", NewTVAPKDomainError(TVAPKErrorUnsupportedABI, "APK 未包含受支持的 ARM ABI")
	}
	return "", NewTVAPKDomainError(TVAPKErrorUnsupportedABI, "APK 同时包含多个 ABI，首期不接受 universal 包")
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type parsedBinaryManifest struct {
	PackageName string `json:"package_name"`
	VersionCode int64  `json:"version_code"`
	VersionName string `json:"version_name"`
	Debuggable  bool   `json:"debuggable"`
}

const (
	axmlChunkStringPool = 0x0001
	axmlChunkStartNS    = 0x0100
	axmlChunkEndNS      = 0x0101
	axmlChunkStartTag   = 0x0102

	axmlTypeString     = 0x03
	axmlTypeIntBoolean = 0x12
	axmlTypeFirstInt   = 0x10
	axmlResAndroid     = 0x01010000
	axmlAttrDebuggable = 0x0101000f

	manifestAttrVersionCode = "versionCode"
	manifestAttrVersionName = "versionName"
	manifestTagName         = "manifest"
	applicationTagName      = "application"
)

type binaryXMLStringPool struct {
	strings []string
}

func (p binaryXMLStringPool) Get(index uint32) string {
	if index == 0xffffffff || int(index) >= len(p.strings) {
		return ""
	}
	return p.strings[index]
}

type binaryXMLParser struct {
	data       []byte
	offset     int
	stringPool binaryXMLStringPool
}

func parseBinaryManifest(reader *zip.Reader) (parsedBinaryManifest, error) {
	var manifestBytes []byte
	for _, file := range reader.File {
		if strings.TrimSpace(file.Name) != "AndroidManifest.xml" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return parsedBinaryManifest{}, fmt.Errorf("open manifest: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, 4*1024*1024))
		closeErr := rc.Close()
		if readErr != nil {
			return parsedBinaryManifest{}, fmt.Errorf("read manifest: %w", readErr)
		}
		if closeErr != nil {
			return parsedBinaryManifest{}, fmt.Errorf("close manifest: %w", closeErr)
		}
		manifestBytes = body
		break
	}
	if len(manifestBytes) == 0 {
		return parsedBinaryManifest{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "APK 缺少 AndroidManifest.xml")
	}
	return parseBinaryManifestBytes(manifestBytes)
}

func parseBinaryManifestBytes(data []byte) (parsedBinaryManifest, error) {
	if len(data) < 8 {
		return parsedBinaryManifest{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "AndroidManifest.xml 结构损坏")
	}
	parser := &binaryXMLParser{data: data}
	chunkType := parser.readU16(0)
	headerSize := parser.readU16(2)
	chunkSize := parser.readU32(4)
	if chunkType != 0x0003 || headerSize != 8 || int(chunkSize) > len(data) {
		return parsedBinaryManifest{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "AndroidManifest.xml 结构损坏")
	}
	parser.offset = int(headerSize)

	result := parsedBinaryManifest{}
	for parser.offset+8 <= len(data) {
		chunkType := parser.readU16(parser.offset)
		headerSize := int(parser.readU16(parser.offset + 2))
		chunkSize := int(parser.readU32(parser.offset + 4))
		if headerSize < 8 || chunkSize < headerSize || parser.offset+chunkSize > len(data) {
			return parsedBinaryManifest{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "AndroidManifest.xml chunk 损坏")
		}

		switch chunkType {
		case axmlChunkStringPool:
			pool, err := parseStringPoolChunk(data[parser.offset : parser.offset+chunkSize])
			if err != nil {
				return parsedBinaryManifest{}, err
			}
			parser.stringPool = pool
		case axmlChunkStartTag:
			name, attrs, err := parseStartTagChunk(data[parser.offset:parser.offset+chunkSize], parser.stringPool)
			if err != nil {
				return parsedBinaryManifest{}, err
			}
			switch name {
			case manifestTagName:
				result.PackageName = strings.TrimSpace(attrs["package"])
				if raw := strings.TrimSpace(attrs[manifestAttrVersionCode]); raw != "" {
					versionCode, convErr := strconv.ParseInt(raw, 10, 64)
					if convErr == nil {
						result.VersionCode = versionCode
					}
				}
				result.VersionName = strings.TrimSpace(attrs[manifestAttrVersionName])
			case applicationTagName:
				debugRaw := strings.TrimSpace(attrs["debuggable"])
				result.Debuggable = strings.EqualFold(debugRaw, "true") || debugRaw == "1"
			}
		}

		parser.offset += chunkSize
	}
	return result, nil
}

func parseStringPoolChunk(chunk []byte) (binaryXMLStringPool, error) {
	if len(chunk) < 28 {
		return binaryXMLStringPool{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "字符串池结构损坏")
	}
	stringCount := int(binary.LittleEndian.Uint32(chunk[8:12]))
	styleCount := int(binary.LittleEndian.Uint32(chunk[12:16]))
	flags := binary.LittleEndian.Uint32(chunk[16:20])
	stringsStart := int(binary.LittleEndian.Uint32(chunk[20:24]))
	stylesStart := int(binary.LittleEndian.Uint32(chunk[24:28]))
	isUTF8 := flags&0x00000100 != 0

	offsetTableStart := 28
	offsetTableSize := stringCount * 4
	if offsetTableStart+offsetTableSize > len(chunk) || stringsStart > len(chunk) {
		return binaryXMLStringPool{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "字符串池偏移损坏")
	}

	stringsEnd := len(chunk)
	if styleCount > 0 && stylesStart > 0 && stylesStart < stringsEnd {
		stringsEnd = stylesStart
	}

	out := make([]string, stringCount)
	for i := 0; i < stringCount; i++ {
		entryOffset := int(binary.LittleEndian.Uint32(chunk[offsetTableStart+i*4 : offsetTableStart+(i+1)*4]))
		absolute := stringsStart + entryOffset
		if absolute < stringsStart || absolute >= stringsEnd {
			return binaryXMLStringPool{}, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "字符串池内容损坏")
		}
		value, err := decodeStringPoolEntry(chunk[absolute:stringsEnd], isUTF8)
		if err != nil {
			return binaryXMLStringPool{}, err
		}
		out[i] = value
	}
	return binaryXMLStringPool{strings: out}, nil
}

func decodeStringPoolEntry(data []byte, utf8 bool) (string, error) {
	if utf8 {
		_, n1, ok := decodeLength8(data)
		if !ok {
			return "", NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "UTF-8 字符串长度损坏")
		}
		byteLen, n2, ok := decodeLength8(data[n1:])
		if !ok {
			return "", NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "UTF-8 字节长度损坏")
		}
		start := n1 + n2
		end := start + byteLen
		if end > len(data) {
			return "", NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "UTF-8 字符串内容损坏")
		}
		return string(data[start:end]), nil
	}

	charLen, consumed, ok := decodeLength16(data)
	if !ok {
		return "", NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "UTF-16 字符串长度损坏")
	}
	byteLen := charLen * 2
	end := consumed + byteLen
	if end > len(data) {
		return "", NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "UTF-16 字符串内容损坏")
	}
	runes := make([]uint16, charLen)
	for i := 0; i < charLen; i++ {
		pos := consumed + i*2
		runes[i] = binary.LittleEndian.Uint16(data[pos : pos+2])
	}
	return string(utf16Decode(runes)), nil
}

func decodeLength8(data []byte) (length int, consumed int, ok bool) {
	if len(data) < 1 {
		return 0, 0, false
	}
	length = int(data[0])
	consumed = 1
	if length&0x80 != 0 {
		if len(data) < 2 {
			return 0, 0, false
		}
		length = ((length & 0x7f) << 8) | int(data[1])
		consumed = 2
	}
	return length, consumed, true
}

func decodeLength16(data []byte) (length int, consumed int, ok bool) {
	if len(data) < 2 {
		return 0, 0, false
	}
	first := binary.LittleEndian.Uint16(data[:2])
	consumed = 2
	length = int(first)
	if first&0x8000 != 0 {
		if len(data) < 4 {
			return 0, 0, false
		}
		second := binary.LittleEndian.Uint16(data[2:4])
		length = int(first&0x7fff)<<16 | int(second)
		consumed = 4
	}
	return length, consumed, true
}

func parseStartTagChunk(chunk []byte, pool binaryXMLStringPool) (string, map[string]string, error) {
	if len(chunk) < 36 {
		return "", nil, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "开始标签结构损坏")
	}
	nameIdx := binary.LittleEndian.Uint32(chunk[20:24])
	attrStart := int(binary.LittleEndian.Uint16(chunk[24:26]))
	attrSize := int(binary.LittleEndian.Uint16(chunk[26:28]))
	attrCount := int(binary.LittleEndian.Uint16(chunk[28:30]))
	attrStart += 16
	if attrSize < 20 || attrStart < 16 || attrStart+attrCount*attrSize > len(chunk) {
		return "", nil, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "标签属性结构损坏")
	}

	name := pool.Get(nameIdx)
	attrs := make(map[string]string, attrCount)
	for i := 0; i < attrCount; i++ {
		offset := attrStart + i*attrSize
		raw := chunk[offset : offset+attrSize]
		if len(raw) < 20 {
			return "", nil, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "标签属性长度不足")
		}
		attrName := pool.Get(binary.LittleEndian.Uint32(raw[4:8]))
		rawValueIdx := binary.LittleEndian.Uint32(raw[8:12])
		typedSize := binary.LittleEndian.Uint16(raw[12:14])
		if typedSize < 8 || int(typedSize)+12 > len(raw) {
			return "", nil, NewTVAPKDomainError(TVAPKErrorAPKStructureBroken, "属性 typed value 损坏")
		}
		dataType := raw[15]
		dataValue := binary.LittleEndian.Uint32(raw[16:20])
		value := decodeTypedValue(rawValueIdx, dataType, dataValue, pool)

		if attrName == "" && dataValue == axmlAttrDebuggable {
			attrName = "debuggable"
		}
		if attrName != "" {
			attrs[attrName] = value
		}
	}
	return name, attrs, nil
}

func decodeTypedValue(rawValueIdx uint32, dataType byte, dataValue uint32, pool binaryXMLStringPool) string {
	switch dataType {
	case axmlTypeString:
		if rawValueIdx != 0xffffffff {
			return pool.Get(rawValueIdx)
		}
		return pool.Get(dataValue)
	case axmlTypeIntBoolean:
		if dataValue != 0 {
			return "true"
		}
		return "false"
	default:
		if dataType >= axmlTypeFirstInt {
			return strconv.FormatUint(uint64(dataValue), 10)
		}
		return strconv.FormatUint(uint64(dataValue), 10)
	}
}

func (p *binaryXMLParser) readU16(offset int) uint16 {
	return binary.LittleEndian.Uint16(p.data[offset : offset+2])
}

func (p *binaryXMLParser) readU32(offset int) uint32 {
	return binary.LittleEndian.Uint32(p.data[offset : offset+4])
}

func utf16Decode(input []uint16) []rune {
	out := make([]rune, 0, len(input))
	for i := 0; i < len(input); i++ {
		r := input[i]
		switch {
		case r < 0xd800 || r > 0xdfff:
			out = append(out, rune(r))
		case 0xd800 <= r && r <= 0xdbff && i+1 < len(input):
			r2 := input[i+1]
			if 0xdc00 <= r2 && r2 <= 0xdfff {
				code := (uint32(r-0xd800) << 10) + uint32(r2-0xdc00) + 0x10000
				out = append(out, rune(code))
				i++
				continue
			}
			out = append(out, rune(r))
		default:
			out = append(out, rune(r))
		}
	}
	return out
}
