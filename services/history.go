package services

type HistoryService struct{}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

func (s *HistoryService) List() []map[string]interface{} {
	return []map[string]interface{}{}
}
