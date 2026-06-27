//go:build test

package realmtest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sessionTTL             = time.Minute
	eventsBufferSize       = 16
	maxPendingAttempts     = eventsBufferSize
	maxRequestBodyBytes    = 4 << 10
	connectResponseTimeout = 2 * time.Second
	nonceHexLength         = 32
	obfsHexLength          = 64
)

type rendezvousServer struct {
	token          string
	mu             sync.Mutex
	realms         map[string]*realmSession
	sessions       map[string]*realmSession
	realmIDPattern *regexp.Regexp
}

type realmSession struct {
	id       string
	realmID  string
	addresses []string
	expires  time.Time
	events   chan sessionEvent
	done     chan struct{}
	closed   bool
	pending  map[string]chan punchResponsePayload
}

type sessionEvent struct {
	kind string
	data any
}

type punchResponsePayload struct {
	addresses []string
}

type punchEvent struct {
	Addresses []string `json:"addresses"`
	Nonce     string   `json:"nonce"`
	Obfs      string   `json:"obfs"`
}

func StartRendezvous(t *testing.T, token string) (hostPort string, cleanup func()) {
	t.Helper()
	s := &rendezvousServer{
		token:          token,
		realms:         make(map[string]*realmSession),
		sessions:       make(map[string]*realmSession),
		realmIDPattern: regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`),
	}
	srv := &http.Server{
		Handler:           http.HandlerFunc(s.handle),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = srv.Serve(ln) }()
	return ln.Addr().String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
}

func (s *rendezvousServer) handle(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "v1" {
		writeErr(w, http.StatusNotFound, "not_found", "unknown path")
		return
	}
	realmID := parts[1]
	if !s.realmIDPattern.MatchString(realmID) {
		writeErr(w, http.StatusBadRequest, "bad_request", "invalid realm name")
		return
	}
	switch {
	case r.Method == http.MethodPost && len(parts) == 2:
		s.register(w, r, realmID)
	case r.Method == http.MethodDelete && len(parts) == 2:
		s.deregister(w, r, realmID)
	case r.Method == http.MethodGet && len(parts) == 3 && parts[2] == "events":
		s.events(w, r, realmID)
	case r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "heartbeat":
		s.heartbeat(w, r, realmID)
	case r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "connect":
		s.connect(w, r, realmID)
	case r.Method == http.MethodPost && len(parts) == 4 && parts[2] == "connects":
		s.connectResponse(w, r, realmID, parts[3])
	default:
		writeErr(w, http.StatusNotFound, "not_found", "unknown path")
	}
}

func (s *rendezvousServer) register(w http.ResponseWriter, r *http.Request, realmID string) {
	if bearer(r) != s.token {
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid realm token")
		return
	}
	var req struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := validateAddresses(req.Addresses); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	s.mu.Lock()
	if _, exists := s.realms[realmID]; exists {
		s.mu.Unlock()
		writeErr(w, http.StatusConflict, "realm_taken", "realm already registered")
		return
	}
	sess := &realmSession{
		id:        randToken(),
		realmID:   realmID,
		addresses: append([]string(nil), req.Addresses...),
		expires:   time.Now().Add(sessionTTL),
		events:    make(chan sessionEvent, eventsBufferSize),
		done:      make(chan struct{}),
		pending:   make(map[string]chan punchResponsePayload),
	}
	s.realms[realmID] = sess
	s.sessions[sess.id] = sess
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"session_id": sess.id,
		"ttl":        int(sessionTTL.Seconds()),
	})
}

func (s *rendezvousServer) deregister(w http.ResponseWriter, r *http.Request, realmID string) {
	sess := s.getSessionByToken(bearer(r), realmID)
	if sess == nil {
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid session token")
		return
	}
	s.removeSession(sess)
	w.WriteHeader(http.StatusNoContent)
}

func (s *rendezvousServer) heartbeat(w http.ResponseWriter, r *http.Request, realmID string) {
	sess := s.getSessionByToken(bearer(r), realmID)
	if sess == nil {
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid session token")
		return
	}
	var req struct {
		Addresses []string `json:"addresses"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			writeErr(w, http.StatusBadRequest, "bad_request", "invalid json")
			return
		}
	}
	if req.Addresses != nil {
		if err := validateAddresses(req.Addresses); err != nil {
			writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}
	}
	s.mu.Lock()
	if sess.closed {
		s.mu.Unlock()
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid session token")
		return
	}
	sess.expires = time.Now().Add(sessionTTL)
	if req.Addresses != nil {
		sess.addresses = append([]string(nil), req.Addresses...)
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ttl": int(sessionTTL.Seconds())})
}

func (s *rendezvousServer) events(w http.ResponseWriter, r *http.Request, realmID string) {
	sess := s.getSessionByToken(bearer(r), realmID)
	if sess == nil {
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid session token")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "bad_request", "streaming not supported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-sess.done:
			return
		case ev := <-sess.events:
			data, _ := json.Marshal(ev.data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.kind, data)
			flusher.Flush()
		}
	}
}

func (s *rendezvousServer) connect(w http.ResponseWriter, r *http.Request, realmID string) {
	if bearer(r) != s.token {
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid realm token")
		return
	}
	var req struct {
		Addresses []string `json:"addresses"`
		Nonce     string   `json:"nonce"`
		Obfs      string   `json:"obfs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := validateAddresses(req.Addresses); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := validateHexField("nonce", req.Nonce, nonceHexLength); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := validateHexField("obfs", req.Obfs, obfsHexLength); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	s.mu.Lock()
	sess := s.realms[realmID]
	if sess == nil || sess.closed || time.Now().After(sess.expires) {
		s.mu.Unlock()
		writeErr(w, http.StatusNotFound, "realm_not_found", "realm not registered")
		return
	}
	serverAddrs := append([]string(nil), sess.addresses...)
	s.mu.Unlock()

	respCh, ok := s.registerPending(sess, req.Nonce)
	if !ok {
		writeErr(w, http.StatusServiceUnavailable, "rate_limited", "too many in-flight connect attempts")
		return
	}
	defer s.cancelPending(sess, req.Nonce)

	if !s.sendEvent(sess, sessionEvent{kind: "punch", data: punchEvent{
		Addresses: req.Addresses,
		Nonce:     req.Nonce,
		Obfs:      req.Obfs,
	}}) {
		writeErr(w, http.StatusServiceUnavailable, "rate_limited", "server event buffer full")
		return
	}

	timer := time.NewTimer(connectResponseTimeout)
	defer timer.Stop()
	select {
	case payload, ok := <-respCh:
		if !ok {
			writeErr(w, http.StatusNotFound, "realm_not_found", "realm not registered")
			return
		}
		if len(payload.addresses) > 0 {
			serverAddrs = payload.addresses
		}
	case <-timer.C:
	case <-sess.done:
		writeErr(w, http.StatusNotFound, "realm_not_found", "realm not registered")
		return
	case <-r.Context().Done():
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"addresses": serverAddrs,
		"nonce":     req.Nonce,
		"obfs":      req.Obfs,
	})
}

func (s *rendezvousServer) connectResponse(w http.ResponseWriter, r *http.Request, realmID, nonce string) {
	if err := validateHexField("nonce", nonce, nonceHexLength); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	sess := s.getSessionByToken(bearer(r), realmID)
	if sess == nil {
		writeErr(w, http.StatusUnauthorized, "invalid_token", "invalid session token")
		return
	}
	var req struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := validateAddresses(req.Addresses); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if !s.deliverPending(sess, nonce, punchResponsePayload{addresses: req.Addresses}) {
		writeErr(w, http.StatusNotFound, "attempt_not_found", "no pending attempt for nonce")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *rendezvousServer) getSessionByToken(token, realmID string) *realmSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.sessions[token]
	if sess == nil || sess.closed || sess.realmID != realmID || time.Now().After(sess.expires) {
		return nil
	}
	return sess
}

func (s *rendezvousServer) removeSession(sess *realmSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess.closed {
		return
	}
	sess.closed = true
	close(sess.done)
	if s.realms[sess.realmID] == sess {
		delete(s.realms, sess.realmID)
	}
	delete(s.sessions, sess.id)
	for nonce, ch := range sess.pending {
		close(ch)
		delete(sess.pending, nonce)
	}
}

func (s *rendezvousServer) registerPending(sess *realmSession, nonce string) (chan punchResponsePayload, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess.closed || len(sess.pending) >= maxPendingAttempts {
		return nil, false
	}
	if _, exists := sess.pending[nonce]; exists {
		return nil, false
	}
	ch := make(chan punchResponsePayload, 1)
	sess.pending[nonce] = ch
	return ch, true
}

func (s *rendezvousServer) deliverPending(sess *realmSession, nonce string, payload punchResponsePayload) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess.closed {
		return false
	}
	ch, ok := sess.pending[nonce]
	if !ok {
		return false
	}
	delete(sess.pending, nonce)
	select {
	case ch <- payload:
	default:
	}
	return true
}

func (s *rendezvousServer) cancelPending(sess *realmSession, nonce string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(sess.pending, nonce)
}

func (s *rendezvousServer) sendEvent(sess *realmSession, ev sessionEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess.closed {
		return false
	}
	select {
	case sess.events <- ev:
		return true
	default:
		return false
	}
}

func randToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

func validateAddresses(addrs []string) error {
	if len(addrs) == 0 {
		return fmt.Errorf("at least one address required")
	}
	for _, a := range addrs {
		ap, err := netip.ParseAddrPort(a)
		if err != nil || !ap.IsValid() {
			return fmt.Errorf("invalid address: %s", a)
		}
	}
	return nil
}

func validateHexField(name, value string, length int) error {
	if len(value) != length {
		return fmt.Errorf("%s must be %d hex characters", name, length)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s must be valid hex", name)
	}
	return nil
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
