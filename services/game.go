package services

// GameService handles game-related business logic.
type GameService struct{}

// NewGameService creates a new game service instance.
func NewGameService() *GameService {
	return &GameService{}
}

// List returns all available games.
func (s *GameService) List() []map[string]interface{} {
	return []map[string]interface{}{}
}
