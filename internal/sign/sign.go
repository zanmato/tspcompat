package sign

// NewSign represents the new sign format.
type NewSign struct {
	ID          string `json:"id"`
	Word        string `json:"word"`
	Description string `json:"description"`
	LastUpdate  string `json:"last_update"`
	// CategoryID   int64   `json:"category_id"`
	// CategorySlug string  `json:"category_slug"`
	Category  string  `json:"category"`
	Glosa     string  `json:"glosa"`
	Frequency *string `json:"frequency,omitempty"`
	// Type         TypeEnum    `json:"type"`
	// Region       interface{} `json:"region"`
	// HandType        HandType      `json:"hand_type"`
	// RightHandform   RightHandform `json:"right_handform"`
	// RightPosition   string        `json:"right_position"`
	// RightAttitude   RightAttitude `json:"right_attitude"`
	// LeftAttitude    string `json:"left_attitude"`
	// LeftHandform    string `json:"left_handform"`
	// Genuine         int64  `json:"genuine"`
	Movie string `json:"movie"`
	// MovieImage      string        `json:"movie_image"`
	Transcription   string `json:"transcription"`
	AlsoMeans       string `json:"also_means"`
	HiddenAlsoMeans string `json:"hidden_also_means"`
	Phrases         []struct {
		Phrase string
		Movie  string
		// MovieImage string
	} `json:"phrases"`
	Categories []struct {
		Slug string
		Name string
	} `json:"categories"`
}
