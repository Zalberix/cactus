package core

import "cactus/internal/storage/db"

func (s *Service) GetMessage() (db.Message, error) {
	return db.Message{}, nil
}
