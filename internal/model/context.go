package model

import "time"

type EntryType string

const (
	Progress EntryType = "progress"
	Decision EntryType = "decision"
	Blocker  EntryType = "blocker"
)

type ContextEntry struct {
	ID        string    `json:"id"`
	Repo      string    `json:"repo"`
	Branch    string    `json:"branch"`
	CommitSHA string    `json:"commit_sha"`
	Files     []string  `json:"files"`
	Session   string    `json:"session"`
	Completed string    `json:"completed"`
	Remaining string    `json:"remaining"`
	Blocker   string    `json:"blocker"`
	Type      EntryType `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}
