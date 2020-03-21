package creating

// Service provides note creating operation
type Service struct {
	repository
}

type repository interface {
	GetNoteCounter(context.Context)
}

// CreateNote creates secure note in storage
func (s Service) CreateNote(n Note) (noteID string) {
	secNote, err := newSecureNote(ctx, dbCli, n)
	if err != nil {
		return web.InternalServerError(), fmt.Errorf("new secure note: %w", err)
	}

	if err := save(ctx, dbCli, secNote); err != nil {
		return web.InternalServerError(), fmt.Errorf("save secured note: %w", err)
	}
}

func newSecureNote(ctx context.Context, dbCli *dynamodb.Client, n note) (SecureNote, error) {
	now := time.Now().UTC()
	ttl := now.Add(time.Duration(n.LifeTimeSeconds) * time.Second).Unix()
	saltedHash, err := generateHashWithSalt([]byte(n.Password))
	if err != nil {
		return secureNote{}, fmt.Errorf("generate hash with salt: %w", err)
	}

	incr, err := getNoteCounter(ctx, dbCli)
	if err != nil {
		return secureNote{}, fmt.Errorf("get note counter: %w", err)
	}
	id := generateHumanFriendlyID(incr)

	return secureNote{
		ID:          id,
		Text:        n.Text,
		Hash:        saltedHash,
		TTL:         ttl,
		OneTimeRead: n.OneTimeRead,
	}, nil
}
