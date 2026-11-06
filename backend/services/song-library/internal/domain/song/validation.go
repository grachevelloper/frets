package song

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

func ValidTitle(title string) bool {
	return utf8.ValidString(title) && !strings.ContainsRune(title, 0) &&
		strings.TrimSpace(title) != "" && utf8.RuneCountInString(title) <= 200
}

func ValidateUpload(in Upload) (UploadMetadata, error) {
	switch {
	case in.Key == "" || len(in.Key) > 128 || !utf8.ValidString(in.Key) || strings.ContainsRune(in.Key, 0):
		return UploadMetadata{}, ErrInvalidUpload
	case !ValidTitle(in.Title):
		return UploadMetadata{}, ErrInvalidTitle
	case len(in.Data) == 0:
		return UploadMetadata{}, ErrInvalidUpload
	case len(in.Data) > MaxFileSizeBytes:
		return UploadMetadata{}, ErrFileTooLarge
	}

	actual := sha256.Sum256(in.Data)
	hash := hex.EncodeToString(actual[:])
	if !strings.EqualFold(hash, in.Hash) {
		return UploadMetadata{}, ErrChecksum
	}

	return UploadMetadata{
		Title: in.Title,
		Key:   in.Key,
		Hash:  hash,
		Size:  len(in.Data),
	}, nil
}
