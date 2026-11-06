package song

type Status string

const (
	Uploading Status = "uploading"
	Ready     Status = "ready"
	Deleting  Status = "deleting"
	Deleted   Status = "deleted"
)

func (s Status) CanPublish() bool {
	return s == Uploading
}

func (s Status) Accessible() bool {
	return s == Ready
}
