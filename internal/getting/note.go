package getting

// SecureNote defines properties of retrieved note from database
type SecureNote struct {
	ID          string `dynamodbav:"pk"`
	Text        string `dynamodbav:"text"`
	Hash        string `dynamodbav:"hash"`
	TTL         int64  `dynamodbav:"ttl"`
	OneTimeRead bool   `dynamodbav:"oneTimeRead"`
}

// Note define properties of successfully decrypted note
type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	TTL  int64  `json:"ttl"`
}
