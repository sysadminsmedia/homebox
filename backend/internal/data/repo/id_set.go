package repo

import (
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/pkgs/set"
)

// HasID is an interface to entities that have an ID uuid.UUID field and a GetID() method.
// This interface is fulfilled by all entities generated by entgo.io/ent via a custom template
type HasID interface {
	GetID() uuid.UUID
}

func newIDSet[T HasID](entities []T) set.Set[uuid.UUID] {
	uuids := make([]uuid.UUID, 0, len(entities))
	for _, e := range entities {
		uuids = append(uuids, e.GetID())
	}

	return set.New(uuids...)
}
