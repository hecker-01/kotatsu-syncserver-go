package services

type GameService struct{}

func NewGameService() *GameService {
	return &GameService{}
}

func (s *GameService) List() []map[string]interface{} {
	return []map[string]interface{}{}
}
