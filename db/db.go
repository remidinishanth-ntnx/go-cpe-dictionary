package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/remidinishanth/go-cpe-dictionary/models"
)

// DB is interface for a database driver
type DB interface {
	Name() string
	CloseDB() error
	InsertCpes([]*models.CategorizedCpe) error
}

// NewDB return DB accessor.
func NewDB(dbType, dbpath string, debugSQL bool) (DB, bool, error) {
	switch dbType {
	case dialectSqlite3, dialectMysql, dialectPostgreSQL:
		return NewRDB(dbType, dbpath, debugSQL)
	}
	return nil, false, fmt.Errorf("Invalid database dialect, %s", dbType)
}

func chunkSlice(l []*models.CategorizedCpe, n int) chan []*models.CategorizedCpe {
	ch := make(chan []*models.CategorizedCpe)
	go func() {
		for i := 0; i < len(l); i += n {
			fromIdx := i
			toIdx := i + n
			if toIdx > len(l) {
				toIdx = len(l)
			}
			ch <- l[fromIdx:toIdx]
		}
		close(ch)
	}()
	return ch
}

// GetByExactTitle Returns the CPE strings which exactly matches the title string
func (r *RDBDriver) GetByExactTitle(title string) ([]models.CategorizedCpe, error) {
	cpes := []models.CategorizedCpe{}
	err := r.conn.Where(&models.CategorizedCpe{Title: title}).Find(&cpes).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return cpes, nil
}

// GetByLikeTitle Returns the CPE strings which matches the title as substring
func (r *RDBDriver) GetByLikeTitle(title string) ([]models.CategorizedCpe, error) {
	cpes := []models.CategorizedCpe{}
	err := r.conn.Where("title LIKE %?%", title).Find(&cpes).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return cpes, nil
}
