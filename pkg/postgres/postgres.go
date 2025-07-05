package postgres

import (
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"golang-sample/pkg/config"
)

// NewGormDB creates a new gorm postgresql
func NewGormDB(
	pgDSN string,
) (*gorm.DB, func(), error) {

	pgCfg := postgres.Config{
		DSN:                  pgDSN,
		PreferSimpleProtocol: true,
	}

	db, err := gorm.Open(postgres.New(pgCfg), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "gorm.Open")
	}

	if config.ENV.APP.DEBUG {
		db = db.Debug()
	}

	return db, func() {
		s, _ := db.DB()
		_ = s.Close()
	}, nil
}
