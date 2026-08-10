package session

// GetDescribe returns the review header set by `crit describe`.
func (s *Session) GetDescribe() (title, description string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.title, s.description
}
