package services

type PlayerService struct{}

func NewPlayerService() *PlayerService {
	return &PlayerService{}
}

func (s *PlayerService) List() []map[string]interface{} {
	return []map[string]interface{}{}
}
