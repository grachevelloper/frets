package song

const MaxFileSizeBytes = 16 * 1024 * 1024 // 16 MiB

type Upload struct {
	Title string
	Key   string
	Hash  string
	Data  []byte
}

type UploadMetadata struct {
	Title string
	Key   string
	Hash  string
	Size  int
}
