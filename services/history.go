package services

// HistoryService handles history-related business logic.
type HistoryService struct{}

// NewHistoryService creates a new history service instance.
func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

// List returns history data for the authenticated user.
func (s *HistoryService) List() []map[string]interface{} {
	return []map[string]interface{}{}
}
