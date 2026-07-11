// Package store persists command aliases in a SQLite database.
package store

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// ErrNotFound is returned when no alias has the requested name.
	ErrNotFound = errors.New("alias not found")
	// ErrExists is returned by Add when the name is already taken.
	ErrExists = errors.New("alias already exists")
)

// Alias is one stored command with its usage statistics.
type Alias struct {
	ID          uint   `gorm:"primarykey"`
	Name        string `gorm:"uniqueIndex;not null"`
	Command     string `gorm:"not null"`
	Description string
	UseCount    int `gorm:"default:0"`
	LastUsedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Store is a SQLite-backed alias repository.
type Store struct {
	db *gorm.DB
}

// Open creates the database (and its parent directory, mode 0700) if
// needed, runs migrations, and returns a ready Store.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Alias{}); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Add(a *Alias) error {
	err := s.db.Create(a).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrExists
	}
	return err
}

func (s *Store) Get(name string) (*Alias, error) {
	var a Alias
	err := s.db.Where("name = ?", name).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// List returns all aliases. sortBy is "name" for alphabetical order;
// anything else sorts by use count (descending), then name.
func (s *Store) List(sortBy string) ([]Alias, error) {
	order := "use_count DESC, name ASC"
	if sortBy == "name" {
		order = "name ASC"
	}
	var out []Alias
	err := s.db.Order(order).Find(&out).Error
	return out, err
}

func (s *Store) Update(a *Alias) error {
	return s.db.Save(a).Error
}

func (s *Store) Delete(name string) error {
	res := s.db.Where("name = ?", name).Delete(&Alias{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Touch records one use of the alias: increments UseCount and stamps
// LastUsedAt.
func (s *Store) Touch(name string) error {
	res := s.db.Model(&Alias{}).Where("name = ?", name).Updates(map[string]any{
		"use_count":    gorm.Expr("use_count + 1"),
		"last_used_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
