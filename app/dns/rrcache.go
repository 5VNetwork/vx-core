package dns

import (
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/task"
	"github.com/miekg/dns"
)

type rrCache struct {
	sync.Mutex
	cleanupCache *task.Periodic
	cache        map[dns.Question]*rrCacheEntry
	// if not zero, dns responses will be cached for this duration.
	// if zero, the minimum ttl of all answers will be used.
	duration uint32
}

type RrCacheSetting struct {
	Duration uint32
}

func NewRrCache(setting RrCacheSetting) *rrCache {
	c := &rrCache{
		cache:    make(map[dns.Question]*rrCacheEntry),
		duration: setting.Duration,
	}
	cleanInterval := time.Second * 30
	if setting.Duration != 0 {
		cleanInterval = time.Duration(setting.Duration) * time.Second
	}
	c.cleanupCache = &task.Periodic{
		Interval: cleanInterval,
		Execute:  c.cleanCache,
	}
	return c
}

func (w *rrCache) cleanCache() error {
	w.Lock()
	defer w.Unlock()
	for k, v := range w.cache {
		if v.expiredAt.Before(time.Now()) {
			delete(w.cache, k)
		}
	}
	return nil
}

func (ns *rrCache) Start() error {
	ns.cleanupCache.Start()
	return nil
}

func (ns *rrCache) Close() error {
	ns.cleanupCache.Close()
	return nil
}

// msg must have at least one question
func (c *rrCache) Set(msg *dns.Msg) {
	c.Lock()
	defer c.Unlock()
	if len(msg.Question) == 0 {
		return
	}
	existing, ok := c.cache[msg.Question[0]]
	if ok {
		// if msg has no answer, and existing has answer and is valid, skip updating it
		if msg.Answer == nil &&
			existing.Answer != nil && existing.expiredAt.After(time.Now()) {
			return
		}
	}

	// set cache duration
	duration := c.duration
	if duration == 0 {
		// set duration to minimum ttl of all answers
		for _, answer := range msg.Answer {
			if answer.Header().Ttl < duration {
				duration = answer.Header().Ttl
			}
		}
	}

	c.cache[msg.Question[0]] = &rrCacheEntry{
		Msg: msg, expiredAt: time.Now().Add(time.Duration(duration) * time.Second)}
}

func (c *rrCache) Get(question *dns.Question) (*dns.Msg, bool) {
	c.Lock()
	defer c.Unlock()

	msg, ok := c.cache[*question]
	if !ok {
		return nil, false
	}
	// filter rrs
	now := time.Now().Unix()
	if msg.expiredAt.Unix() < now {
		delete(c.cache, *question)
		return nil, false
	}

	return msg.Msg, true
}

type rrCacheEntry struct {
	*dns.Msg
	expiredAt time.Time
}
