package antiabuse

import "sync"

//Set defines HashSet structure
type Set[T comparable] struct {
	sync.RWMutex
	m map[T]struct{}
}

// NewSet creates Set with optional key(s)
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{m: make(map[T]struct{})}
}

// Add adds key(s) to Set
func (s *Set[T]) Add(key ...T) {
	s.Lock()
	defer s.Unlock()
	for _, k := range key {
		s.m[k] = struct{}{}
	}
}

// Include asserts key in Set
func (s *Set[T]) Include(key T) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[key]
	return ok
}

// Iter calls f when traversing Set
func (s *Set[T]) Iter(f func(T) error) error {
	s.RLock()
	defer s.RUnlock()
	var err error
	for key := range s.m {
		err = f(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove removes key from Set
func (s *Set[T]) Remove(key T) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

// ToSlice convert Set to slice
func (s *Set[T]) ToSlice() (res []T) {
	s.RLock()
	defer s.RUnlock()
	for key := range s.m {
		res = append(res, key)
	}
	return res
}
