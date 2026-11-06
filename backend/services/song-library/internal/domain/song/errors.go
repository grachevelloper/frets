package song

import "errors"

var (
	ErrNotFound         = errors.New("not_found")
	ErrInvalidUpload    = errors.New("invalid_upload")
	ErrInvalidTitle     = errors.New("invalid_title")
	ErrFileTooLarge     = errors.New("file_too_large")
	ErrChecksum         = errors.New("checksum_mismatch")
	ErrUploadConflict   = errors.New("upload_conflict")
	ErrUploadInProgress = errors.New("upload_in_progress")
	ErrUploadExpired    = errors.New("upload_expired")
	ErrInvalidFragment  = errors.New("invalid_fragment")
	ErrFragmentConflict = errors.New("fragment_conflict")
)

var ErrUnauthorized = errors.New("unauthorized")
