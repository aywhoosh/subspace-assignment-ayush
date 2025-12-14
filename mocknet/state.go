package mocknet

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type Profile struct {
	ID       string `json:"id"`
	First    string `json:"first"`
	Last     string `json:"last"`
	Company  string `json:"company"`
	Title    string `json:"title"`
	Location string `json:"location"`
	Keywords string `json:"keywords"`
}

type ConnectionStatus string

const (
	ConnNone    ConnectionStatus = "none"
	ConnPending ConnectionStatus = "pending"
	ConnAccepted ConnectionStatus = "accepted"
)

type Connection struct {
	ProfileID string
	Status    ConnectionStatus
	Note      string
	SentAt    time.Time
	AcceptedAt *time.Time
}

type Message struct {
	ThreadID  string
	ProfileID string
	FromSelf  bool
	Body      string
	SentAt    time.Time
}

type State struct {
	mu sync.RWMutex

	profiles []Profile
	profilesByID map[string]Profile

	// sessionID -> username
	sessions map[string]string

	// profileID -> connection
	connections map[string]Connection

	// threadID -> messages
	threads map[string][]Message
}

func NewState(seedProfiles []Profile) *State {
	byID := map[string]Profile{}
	for _, p := range seedProfiles {
		byID[p.ID] = p
	}
	return &State{
		profiles: seedProfiles,
		profilesByID: byID,
		sessions: map[string]string{},
		connections: map[string]Connection{},
		threads: map[string][]Message{},
	}
}

func (s *State) NewSession(username string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = username
	return id, nil
}

func (s *State) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *State) UsernameForSession(sessionID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.sessions[sessionID]
	return u, ok
}

type SearchQuery struct {
	Title    string
	Company  string
	Location string
	Keywords string
}

func (s *State) Search(q SearchQuery) []Profile {
	needle := func(hay, want string) bool {
		want = strings.ToLower(strings.TrimSpace(want))
		if want == "" {
			return true
		}
		hay = strings.ToLower(hay)
		return strings.Contains(hay, want)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Profile
	for _, p := range s.profiles {
		if !needle(p.Title, q.Title) {
			continue
		}
		if !needle(p.Company, q.Company) {
			continue
		}
		if !needle(p.Location, q.Location) {
			continue
		}
		if !needle(p.Keywords, q.Keywords) {
			continue
		}
		out = append(out, p)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *State) GetProfile(id string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profilesByID[id]
	return p, ok
}

func (s *State) ConnectionFor(profileID string) Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.connections[profileID]; ok {
		return c
	}
	return Connection{ProfileID: profileID, Status: ConnNone}
}

var ErrAlreadyRequested = errors.New("already requested")

func (s *State) SendConnectionRequest(profileID, note string) (Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.connections[profileID]; ok {
		if c.Status == ConnPending || c.Status == ConnAccepted {
			return c, ErrAlreadyRequested
		}
	}

	c := Connection{ProfileID: profileID, Status: ConnPending, Note: note, SentAt: time.Now().UTC()}
	s.connections[profileID] = c
	return c, nil
}

func (s *State) AcceptConnection(profileID string) (Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.connections[profileID]
	if !ok {
		c = Connection{ProfileID: profileID, Status: ConnNone}
	}
	if c.Status != ConnPending && c.Status != ConnAccepted {
		c.Status = ConnPending
	}
	now := time.Now().UTC()
	c.Status = ConnAccepted
	c.AcceptedAt = &now
	s.connections[profileID] = c

	threadID := ThreadID(profileID)
	if _, ok := s.threads[threadID]; !ok {
		s.threads[threadID] = []Message{}
	}
	return c, nil
}

func ThreadID(profileID string) string {
	return "thread-" + profileID
}

func (s *State) ListConnections() (pending []Connection, accepted []Connection) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.connections {
		switch c.Status {
		case ConnPending:
			pending = append(pending, c)
		case ConnAccepted:
			accepted = append(accepted, c)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].ProfileID < pending[j].ProfileID })
	sort.Slice(accepted, func(i, j int) bool { return accepted[i].ProfileID < accepted[j].ProfileID })
	return pending, accepted
}

func (s *State) ListThreads() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ids []string
	for threadID := range s.threads {
		ids = append(ids, threadID)
	}
	sort.Strings(ids)
	return ids
}

func (s *State) Messages(threadID string) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.threads[threadID]
	out := make([]Message, 0, len(msgs))
	out = append(out, msgs...)
	return out
}

var ErrNotConnected = errors.New("not connected")

func (s *State) SendMessage(profileID, body string) (Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.connections[profileID]
	if !ok || c.Status != ConnAccepted {
		return Message{}, ErrNotConnected
	}
	threadID := ThreadID(profileID)
	m := Message{ThreadID: threadID, ProfileID: profileID, FromSelf: true, Body: body, SentAt: time.Now().UTC()}
	s.threads[threadID] = append(s.threads[threadID], m)
	return m, nil
}
