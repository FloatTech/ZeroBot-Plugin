package antiabuse

import "sync"

//Set defines HashSet structure
type Set struct {
	sync.RWMutex
	m map[string]struct{}
}

var banSet = &Set{m: make(map[string]struct{})}
var wordSet = &Set{m: make(map[string]struct{})}

// Add adds element to Set
func (s *Set) Add(key string) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = struct{}{}
}

// Include asserts element in Set
func (s *Set) Include(key string) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[key]
	return ok
}

// Iter calls f when traversing Set
func (s *Set) Iter(f func(string) error) error {
	s.Lock()
	defer s.Unlock()
	var err error
	for key := range s.m {
		err = f(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove removes element from Set
func (s *Set) Remove(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

// AddMany adds multiple elements to Set
func (s *Set) AddMany(keys []string) {
	s.Lock()
	defer s.Unlock()
	for _, k := range keys {
		s.m[k] = struct{}{}
	}
}
