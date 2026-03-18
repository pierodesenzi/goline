package tasks

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Create(name string) (map[string]interface{}, error) {
	// minimal: just echo back
	return map[string]interface{}{
		"name": name,
	}, nil
}
