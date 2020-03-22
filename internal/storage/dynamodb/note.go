package dynamodb

// Note defines properties of a secured note that is persisted in storage
type Note struct {
	ID          string `dynamodbav:"pk"`
	Text        string `dynamodbav:"text"`
	Hash        string `dynamodbav:"hash"`
	TTL         int64  `dynamodbav:"ttl"`
	OneTimeRead bool   `dynamodbav:"oneTimeRead"`
}
