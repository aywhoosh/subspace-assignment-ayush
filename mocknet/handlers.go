package mocknet

import (
	"errors"
	"fmt"
	"net/url"
	"net/http"
	"strings"
)

const cookieSession = "mocknet_session"

type baseView struct {
	Title    string
	BaseURL  string
	Authed   bool
	Username string
	Flash    string
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) sessionUsername(r *http.Request) (string, bool) {
	c, err := r.Cookie(cookieSession)
	if err != nil {
		return "", false
	}
	u, ok := s.state.UsernameForSession(c.Value)
	return u, ok
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.sessionUsername(r); !ok {
			to := "/login?next=" + safeRedirectPath(r.URL.Path)
			http.Redirect(w, r, to, http.StatusFound)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	username, authed := s.sessionUsername(r)
	data := struct {
		Base baseView
	}{
		Base: baseView{
			Title:    "MockNet",
			BaseURL:  s.BaseURL(),
			Authed:   authed,
			Username: username,
		},
	}
	if authed {
		http.Redirect(w, r, "/search", http.StatusFound)
		return
	}
	s.render(w, "home.html", data)
}

func (s *Server) handleLoginGet(w http.ResponseWriter, r *http.Request) {
	_, authed := s.sessionUsername(r)
	if authed {
		http.Redirect(w, r, "/search", http.StatusFound)
		return
	}

	data := struct {
		Base baseView
		Next string
		Err  string
	}{
		Base: baseView{Title: "Login", BaseURL: s.BaseURL()},
		Next: safeRedirectPath(r.URL.Query().Get("next")),
	}
	s.render(w, "login.html", data)
}

func (s *Server) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))
	next := safeRedirectPath(r.FormValue("next"))
	if next == "" {
		next = "/search"
	}

	if username != s.cfg.SeedCredentials.Username || password != s.cfg.SeedCredentials.Password {
		data := struct {
			Base baseView
			Next string
			Err  string
		}{
			Base: baseView{Title: "Login", BaseURL: s.BaseURL()},
			Next: next,
			Err:  "Invalid credentials (this is a local mock app).",
		}
		s.render(w, "login.html", data)
		return
	}

	sid, err := s.state.NewSession(username)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieSession,
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if s.cfg.CheckpointEnabled {
		http.Redirect(w, r, "/checkpoint?next="+next, http.StatusFound)
		return
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieSession)
	if err == nil {
		s.state.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: cookieSession, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handleCheckpointGet(w http.ResponseWriter, r *http.Request) {
	username, authed := s.sessionUsername(r)
	if !authed {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := struct {
		Base baseView
		Next string
	}{
		Base: baseView{Title: "Checkpoint", BaseURL: s.BaseURL(), Authed: true, Username: username},
		Next: safeRedirectPath(r.URL.Query().Get("next")),
	}
	s.render(w, "checkpoint.html", data)
}

func (s *Server) handleCheckpointPost(w http.ResponseWriter, r *http.Request) {
	_, authed := s.sessionUsername(r)
	if !authed {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	next := safeRedirectPath(r.FormValue("next"))
	if next == "" {
		next = "/search"
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func (s *Server) handleSearchGet(w http.ResponseWriter, r *http.Request) {
	username, _ := s.sessionUsername(r)

	q := SearchQuery{
		Title:    r.URL.Query().Get("title"),
		Company:  r.URL.Query().Get("company"),
		Location: r.URL.Query().Get("location"),
		Keywords: r.URL.Query().Get("keywords"),
	}
	page := qInt(r, "page", 1)
	per := qInt(r, "per", 10)
	if page < 1 {
		page = 1
	}
	if per < 1 {
		per = 10
	}
	if per > 25 {
		per = 25
	}

	all := s.state.Search(q)
	total := len(all)
	start := (page - 1) * per
	if start > total {
		start = total
	}
	end := start + per
	if end > total {
		end = total
	}
	items := all[start:end]

	hasPrev := page > 1
	hasNext := end < total

	data := struct {
		Base     baseView
		Query    SearchQuery
		Page     int
		Per      int
		Total    int
		HasPrev  bool
		HasNext  bool
		Results  []Profile
		PrevHref string
		NextHref string
	}{
		Base: baseView{Title: "Search", BaseURL: s.BaseURL(), Authed: true, Username: username},
		Query:   q,
		Page:    page,
		Per:     per,
		Total:   total,
		HasPrev: hasPrev,
		HasNext: hasNext,
		Results: items,
	}
	data.PrevHref = s.searchHref(q, page-1, per)
	data.NextHref = s.searchHref(q, page+1, per)

	s.render(w, "search.html", data)
}

func (s *Server) searchHref(q SearchQuery, page, per int) string {
	if page < 1 {
		page = 1
	}
	return fmt.Sprintf("/search?title=%s&company=%s&location=%s&keywords=%s&page=%d&per=%d",
		url.QueryEscape(strings.TrimSpace(q.Title)),
		url.QueryEscape(strings.TrimSpace(q.Company)),
		url.QueryEscape(strings.TrimSpace(q.Location)),
		url.QueryEscape(strings.TrimSpace(q.Keywords)),
		page,
		per,
	)
}

func (s *Server) handleProfileGet(w http.ResponseWriter, r *http.Request) {
	username, _ := s.sessionUsername(r)
	id := r.PathValue("id")
	p, ok := s.state.GetProfile(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	c := s.state.ConnectionFor(id)
	data := struct {
		Base       baseView
		Profile    Profile
		Conn       Connection
		NoteLimit  int
		PostAction string
	}{
		Base:       baseView{Title: "Profile", BaseURL: s.BaseURL(), Authed: true, Username: username},
		Profile:    p,
		Conn:       c,
		NoteLimit:  200,
		PostAction: fmt.Sprintf("/profile/%s/connect", id),
	}
	s.render(w, "profile.html", data)
}

func (s *Server) handleConnectPost(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.state.GetProfile(id); !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	note := strings.TrimSpace(r.FormValue("note"))
	if len(note) > 200 {
		http.Error(w, "note too long", http.StatusBadRequest)
		return
	}
	if note == "" {
		note = "(no note)"
	}

	_, err := s.state.SendConnectionRequest(id, note)
	if err != nil {
		if errors.Is(err, ErrAlreadyRequested) {
			http.Redirect(w, r, "/profile/"+id+"?already=1", http.StatusFound)
			return
		}
		http.Error(w, "failed to send request", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/connections", http.StatusFound)
}

func (s *Server) handleConnectionsGet(w http.ResponseWriter, r *http.Request) {
	username, _ := s.sessionUsername(r)
	pending, accepted := s.state.ListConnections()

	data := struct {
		Base     baseView
		Pending  []Connection
		Accepted []Connection
		Profiles map[string]Profile
	}{
		Base:     baseView{Title: "Connections", BaseURL: s.BaseURL(), Authed: true, Username: username},
		Pending:  pending,
		Accepted: accepted,
		Profiles: s.state.profilesByID, // safe read: map is immutable after init
	}
	s.render(w, "connections.html", data)
}

func (s *Server) handleMessagesGet(w http.ResponseWriter, r *http.Request) {
	username, _ := s.sessionUsername(r)
	threadID := strings.TrimSpace(r.URL.Query().Get("thread"))

	type ThreadView struct {
		ThreadID  string
		ProfileID string
		Name      string
		Company   string
	}
	threadIDs := s.state.ListThreads()
	threads := make([]ThreadView, 0, len(threadIDs))
	for _, tid := range threadIDs {
		pid := strings.TrimPrefix(tid, "thread-")
		p := s.state.profilesByID[pid]
		threads = append(threads, ThreadView{
			ThreadID:  tid,
			ProfileID: pid,
			Name:      strings.TrimSpace(p.First + " " + p.Last),
			Company:   p.Company,
		})
	}
	msgs := []Message{}
	selectedProfileID := ""
	if threadID != "" {
		msgs = s.state.Messages(threadID)
		selectedProfileID = strings.TrimPrefix(threadID, "thread-")
	}

	data := struct {
		Base              baseView
		Threads           []ThreadView
		SelectedThreadID  string
		SelectedProfileID string
		Messages          []Message
		Profiles          map[string]Profile
	}{
		Base:              baseView{Title: "Messages", BaseURL: s.BaseURL(), Authed: true, Username: username},
		Threads:           threads,
		SelectedThreadID:  threadID,
		SelectedProfileID: selectedProfileID,
		Messages:          msgs,
		Profiles:          s.state.profilesByID,
	}
	s.render(w, "messages.html", data)
}

func (s *Server) handleMessagesSendPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	profileID := strings.TrimSpace(r.FormValue("profile_id"))
	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		http.Error(w, "empty message", http.StatusBadRequest)
		return
	}
	if len(body) > 500 {
		http.Error(w, "message too long", http.StatusBadRequest)
		return
	}

	_, err := s.state.SendMessage(profileID, body)
	if err != nil {
		if errors.Is(err, ErrNotConnected) {
			http.Error(w, "not connected", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to send", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/messages?thread="+ThreadID(profileID), http.StatusFound)
}

func (s *Server) handleAdminAcceptPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	profileID := strings.TrimSpace(r.FormValue("profile_id"))
	if profileID == "" {
		http.Error(w, "profile_id required", http.StatusBadRequest)
		return
	}
	if _, ok := s.state.GetProfile(profileID); !ok {
		http.Error(w, "unknown profile_id", http.StatusBadRequest)
		return
	}
	_, _ = s.state.AcceptConnection(profileID)
	back := strings.TrimSpace(r.FormValue("back"))
	if back == "" {
		back = "/connections"
	}
	http.Redirect(w, r, safeRedirectPath(back), http.StatusFound)
}

