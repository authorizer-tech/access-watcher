package watcher

import (
	"context"
	"time"
)

type ChangelogDatastore interface {
	GetRelationTupleChanges(ctx context.Context, namespaces []string, timestamp time.Time) (ChangelogIterator, error)
}

// ChangelogIterator is used to iterate over changelog entries as they are yielded.
type ChangelogIterator interface {

	// Next prepares the next changelog entry for reading. It returns true
	// if there is another entry and false if no more entries are available.
	Next() bool

	// Value returns the current most changelog entry that the iterator is
	// iterating over.
	Value() (*ChangelogEntry, error)
}

type ChangelogEntry struct {
	Namespace     string
	Operation     string
	RelationTuple *InternalRelationTuple
	Timestamp     time.Time
}
