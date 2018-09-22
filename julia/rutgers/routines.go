package rutgers

import (
	"sync"

	"github.com/tevjef/uct-backend/julia/rutgers/topic"
)

type Routines struct {
	mu         sync.RWMutex
	routineMap map[string]*topic.Routine
}

func (r *Routines) Exists(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.routineMap[key] != nil
}

func (r *Routines) Set(key string, tr *topic.Routine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routineMap[key] = tr
}

func (r *Routines) Get(key string) *topic.Routine {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.routineMap[key]
}

func (r *Routines) Remove(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.routineMap, key)
}

func (r *Routines) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.routineMap)
}
