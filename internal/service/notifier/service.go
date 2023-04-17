package notifier

// Service keep info about price subscriptions
// and notify consumers about price changes
type Service struct {
}

func NewService() *Service {
	return &Service{}
}
