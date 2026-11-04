package library

import (
	"context"
	"errors"
	"io"
	"time"
)

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

const MaxFileSize = 16 << 20

type Song struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

type FragmentInput struct {
	Title    string `json:"title"`
	StartBar int32  `json:"startBar"`
	EndBar   int32  `json:"endBar"`
}

type Fragment struct {
	ID     string `json:"id"`
	SongID string `json:"songId"`
	FragmentInput
}

type Detail struct {
	Song
	Fragments []Fragment `json:"fragments"`
}

type Upload struct {
	Title, Key, Hash string
	Data             []byte
}

type UploadMetadata struct {
	Title, Key, Hash string
	Size             int
}

type Deletion struct{ SongID, FileKey string }

type Repository interface {
	List(context.Context, string) ([]Song, error)
	Detail(context.Context, string, string) (Detail, error)
	File(context.Context, string, string) (string, int64, error)
	PutFragment(context.Context, string, Fragment) (bool, error)
	Reserve(context.Context, string, UploadMetadata) (Song, string, error)
	Publish(context.Context, string, string) error
	ExpireAttempt(context.Context, string) error
	ExpiredKeys(context.Context) ([]string, error)
	BeginDeletion(context.Context, string, string) (string, error)
	FinishDeletion(context.Context, string) error
	PendingDeletions(context.Context) ([]Deletion, error)
}

type Files interface {
	Put(context.Context, string, []byte) error
	Open(context.Context, string) (io.ReadCloser, error)
	Remove(context.Context, string) error
}

type Service struct {
	repo  Repository
	files Files
}

func New(repo Repository, files Files) *Service {
	return &Service{repo: repo, files: files}
}

func (s *Service) List(ctx context.Context, owner string) ([]Song, error) {
	return s.repo.List(ctx, owner)
}

func (s *Service) Detail(ctx context.Context, owner, id string) (Detail, error) {
	return s.repo.Detail(ctx, owner, id)
}

func (s *Service) Download(ctx context.Context, owner, id string) (io.ReadCloser, int64, error) {
	key, size, err := s.repo.File(ctx, owner, id)
	if err != nil {
		return nil, 0, err
	}

	file, err := s.files.Open(ctx, key)
	return file, size, err
}

func (s *Service) PutFragment(ctx context.Context, owner string, fragment Fragment) (bool, error) {
	if !validTitle(fragment.Title) || fragment.StartBar < 1 || fragment.StartBar > fragment.EndBar {
		return false, ErrInvalidFragment
	}
	return s.repo.PutFragment(ctx, owner, fragment)
}
