package services

// PlayerService handles player-related business logic.
type PlayerService struct{}

// NewPlayerService creates a new player service instance.
func NewPlayerService() *PlayerService {
	return &PlayerService{}
}

// List returns player data for the authenticated user.
func (s *PlayerService) List() []map[string]interface{} {
	return []map[string]interface{}{}
}
