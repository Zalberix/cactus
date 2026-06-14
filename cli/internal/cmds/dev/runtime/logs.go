package runtime

import "sync"

type LogBuffer struct {
	mu       sync.Mutex
	capacity int
	entries  []LogEntry
}

func NewLogBuffer(capacity int) *LogBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &LogBuffer{capacity: capacity}
}

func (b *LogBuffer) Append(entry LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == b.capacity {
		copy(b.entries, b.entries[1:])
		b.entries[len(b.entries)-1] = entry
		return
	}
	b.entries = append(b.entries, entry)
}

func (b *LogBuffer) Entries() []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]LogEntry, len(b.entries))
	copy(out, b.entries)
	return out
}

type LogStore struct {
	mu       sync.Mutex
	capacity int
	buffers  map[string]*LogBuffer
}

func NewLogStore(capacity int) *LogStore {
	if capacity <= 0 {
		capacity = 1
	}
	return &LogStore{
		capacity: capacity,
		buffers:  make(map[string]*LogBuffer),
	}
}

func (s *LogStore) Append(entry LogEntry) {
	s.buffer(entry.TargetID).Append(entry)
}

func (s *LogStore) Entries(targetID string) []LogEntry {
	return s.buffer(targetID).Entries()
}

func (s *LogStore) buffer(targetID string) *LogBuffer {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf, ok := s.buffers[targetID]
	if !ok {
		buf = NewLogBuffer(s.capacity)
		s.buffers[targetID] = buf
	}
	return buf
}
