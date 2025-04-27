package storage

import (
	"fmt"

	"go.uber.org/zap"

	"golang-sample/pkg/models"
)

func (s *Storage) IsExistBy(field string, condition string) (bool, error) {
	// Check if the username exists in the database
	var count int64
	query := fmt.Sprintf("%s = ?", field)
	if err := s.db.Model(&models.User{}).Where(query, condition).Count(&count).Error; err != nil {
		s.log.Errorf("Failed to check if %s exists, err: %#v", field, zap.Error(err))
		return false, err
	}
	return count > 0, nil
}
