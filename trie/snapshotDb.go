package trie

import (
	"os"
	"sync"

	"github.com/ElrondNetwork/elrond-go/common"
)

type snapshotDb struct {
	common.DBWriteCacher
	numReferences        uint32
	shouldBeRemoved      bool
	shouldBeDisconnected bool
	path                 string
	mutex                sync.RWMutex
}

// DecreaseNumReferences decreases the num references counter
func (s *snapshotDb) DecreaseNumReferences() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.numReferences > 0 {
		s.numReferences--
	}

	if s.numReferences == 0 && s.shouldBeRemoved {
		removeSnapshot(s.DBWriteCacher, s.path)
		return
	}
	if s.numReferences == 0 && s.shouldBeDisconnected {
		err := disconnectSnapshot(s.DBWriteCacher)
		if err != nil {
			log.Error("snapshotDb: disconnectSnapshot", "error", err.Error())
		}
		return
	}
}

func disconnectSnapshot(db common.DBWriteCacher) error {
	return db.Close()
}

func removeSnapshot(db common.DBWriteCacher, path string) {
	err := disconnectSnapshot(db)
	if err != nil {
		log.Error("trie storage manager: disconnectSnapshot", "error", err.Error())
		return
	}

	log.Debug("remove trie snapshot db", "snapshot path", path)
	go removeDirectory(path)
}

func removeDirectory(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Error(err.Error())
	}
}

// IncreaseNumReferences increases the num references counter
func (s *snapshotDb) IncreaseNumReferences() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.numReferences++
}

// MarkForRemoval marks the current db for removal. When the numReferences buffer reaches 0, the db will be removed
func (s *snapshotDb) MarkForRemoval() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.shouldBeRemoved = true
}

// MarkForDisconnection marks the current db for disconnection. When the numReferences buffer reaches 0, the db will be disconnected
func (s *snapshotDb) MarkForDisconnection() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.shouldBeDisconnected = true
}

// SetPath sets the db path
func (s *snapshotDb) SetPath(path string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.path = path
}

// IsInUse returns true if the numReferences counter is greater than 0
func (s *snapshotDb) IsInUse() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.numReferences > 0
}
