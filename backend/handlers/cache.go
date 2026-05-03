package handlers

import (
	"sync"
	"time"

	"github.com/ahmadfarisfs/bajet/models"
)

const cacheTTL = 30 * time.Second

// ── cycles list cache (per user) ─────────────────────────────────────────────

type cyclesEntry struct {
	cycles  []models.Cycle
	expires time.Time
}

var (
	cyclesCache   = map[string]cyclesEntry{}
	cyclesCacheMu sync.RWMutex
)

func cacheGetCycles(uid string) ([]models.Cycle, bool) {
	cyclesCacheMu.RLock()
	defer cyclesCacheMu.RUnlock()
	e, ok := cyclesCache[uid]
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return e.cycles, true
}

func cacheSetCycles(uid string, cycles []models.Cycle) {
	cyclesCacheMu.Lock()
	defer cyclesCacheMu.Unlock()
	cyclesCache[uid] = cyclesEntry{cycles: cycles, expires: time.Now().Add(cacheTTL)}
}

// ── single cycle cache (per user+id) ─────────────────────────────────────────

type cycleEntry struct {
	cycle   models.Cycle
	expires time.Time
}

var (
	cycleCache   = map[string]cycleEntry{}
	cycleCacheMu sync.RWMutex
)

func cacheKey(uid, id string) string { return uid + ":" + id }

func cacheGetCycle(uid, id string) (models.Cycle, bool) {
	cycleCacheMu.RLock()
	defer cycleCacheMu.RUnlock()
	e, ok := cycleCache[cacheKey(uid, id)]
	if !ok || time.Now().After(e.expires) {
		return models.Cycle{}, false
	}
	return e.cycle, true
}

func cacheSetCycle(uid, id string, cycle models.Cycle) {
	cycleCacheMu.Lock()
	defer cycleCacheMu.Unlock()
	cycleCache[cacheKey(uid, id)] = cycleEntry{cycle: cycle, expires: time.Now().Add(cacheTTL)}
}

// invalidateUser drops all cached data for a user (called on any write).
func invalidateUser(uid string) {
	cyclesCacheMu.Lock()
	delete(cyclesCache, uid)
	cyclesCacheMu.Unlock()

	// drop all single-cycle entries for this user
	cycleCacheMu.Lock()
	prefix := uid + ":"
	for k := range cycleCache {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(cycleCache, k)
		}
	}
	cycleCacheMu.Unlock()
}
