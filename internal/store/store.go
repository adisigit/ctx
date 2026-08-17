package store

import "ctx/internal/model"

type Store interface {
	Save(entry model.ContextEntry) error
	LastByRepo(repo string) (*model.ContextEntry, error)
	AllByRepo(repo string) ([]model.ContextEntry, error)
}
