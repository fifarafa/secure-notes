package creating

// Note defines properties of a note to be created
type Note struct {
	Text            string `json:"text"`
	Password        string `json:"password"`
	LifeTimeSeconds int64  `json:"lifeTimeSeconds"`
	OneTimeRead     bool   `json:"oneTimeRead"`
}

// SecureNote define properties of a note after securing it
type SecureNote struct {
	ID          string `dynamodbav:"pk"`
	Text        string `dynamodbav:"text"`
	Hash        string `dynamodbav:"hash"`
	TTL         int64  `dynamodbav:"ttl"`
	OneTimeRead bool   `dynamodbav:"oneTimeRead"`
}
