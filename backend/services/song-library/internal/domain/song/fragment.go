package song

type FragmentInput struct {
	Title    string
	StartBar int32
	EndBar   int32
}

type Fragment struct {
	ID     string
	SongID string
	FragmentInput
}

func (f FragmentInput) Validate() error {
	if !ValidTitle(f.Title) || f.StartBar < 1 || f.StartBar > f.EndBar {
		return ErrInvalidFragment
	}
	return nil
}
