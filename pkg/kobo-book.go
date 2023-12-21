package pkg

type KoboBook struct {
	ID    string
	Title string
	URL   string

	Finished        bool
	ProgressPercent int

	Parts map[string]KoboBookPart

	// IsBook true for book, false for Pocket
	IsBook bool
}

type KoboBookPart struct {
	WordCount int
}

func (k KoboBook) TotalWords() int {
	result := 0

	for part := range k.Parts {
		result += k.Parts[part].WordCount
	}

	return result
}
