package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	beeperdesktop "github.com/cameronaaron/beeper-go-sdk"
	"github.com/cameronaaron/beeper-go-sdk/resources"
)

//go:embed ui/*
var uiFS embed.FS

type sessionData struct {
	Client   *beeperdesktop.BeeperDesktop
	UserInfo *resources.UserInfo
	Created  time.Time
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionData
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]*sessionData)}
}

func (s *sessionStore) set(id string, data *sessionData) {
	s.mu.Lock()
	s.sessions[id] = data
	s.mu.Unlock()
}

func (s *sessionStore) get(id string) (*sessionData, bool) {
	s.mu.RLock()
	data, ok := s.sessions[id]
	s.mu.RUnlock()
	return data, ok
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func main() {
	addr := readEnv("PORT", "8080")
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	store := newSessionStore()

	uifs, err := fs.Sub(uiFS, "ui")
	if err != nil {
		log.Fatalf("failed to initialize embedded UI: %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(uifs))))
	mux.HandleFunc("/", serveIndex(uifs))

	mux.HandleFunc("/api/session", withJSON(func(w http.ResponseWriter, r *http.Request) {
		sess, _ := fetchSession(store, r)
		if sess == nil {
			writeJSON(w, http.StatusOK, map[string]any{"authenticated": false})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"user": map[string]any{
				"subject":   sess.UserInfo.Sub,
				"scope":     sess.UserInfo.Scope,
				"token_use": sess.UserInfo.TokenUse,
				"client_id": sess.UserInfo.ClientID,
			},
		})
	}))

	mux.HandleFunc("/api/login", withJSON(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Token   string `json:"token"`
			BaseURL string `json:"baseUrl"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		token := strings.TrimSpace(payload.Token)
		if token == "" {
			writeError(w, http.StatusBadRequest, "access token is required")
			return
		}

		clientOpts := []beeperdesktop.ClientOption{
			beeperdesktop.WithAccessToken(token),
		}

		if base := strings.TrimSpace(payload.BaseURL); base != "" {
			clientOpts = append(clientOpts, beeperdesktop.WithBaseURL(base))
		}

		client, err := beeperdesktop.New(clientOpts...)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		info, err := client.Token.Info(ctx)
		if err != nil {
			writeError(w, http.StatusUnauthorized, fmt.Sprintf("failed to validate token: %v", err))
			return
		}

		sessID := generateSessionID()
		store.set(sessID, &sessionData{Client: client, UserInfo: info, Created: time.Now()})

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   12 * 3600,
		})

		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"user": map[string]any{
				"subject":   info.Sub,
				"scope":     info.Scope,
				"token_use": info.TokenUse,
			},
		})
	}))

	mux.HandleFunc("/api/logout", withJSON(func(w http.ResponseWriter, r *http.Request) {
		sessID := readSessionCookie(r)
		if sessID != "" {
			store.delete(sessID)
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}))

	mux.HandleFunc("/api/chats", withJSON(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := mustSession(store, w, r)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := sess.Client.Chats.Search(ctx, resources.ChatSearchParams{
			Limit: beeperdesktop.IntPtr(50),
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to load chats: %v", err))
			return
		}

		var chats []map[string]any
		for _, chat := range resp.Items {
			title := chat.Title
			if title == "" {
				title = chat.ID
			}
			chats = append(chats, map[string]any{
				"id":           chat.ID,
				"title":        title,
				"network":      chat.Network,
				"unreadCount":  chat.UnreadCount,
				"lastActivity": chat.LastActivity,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{"chats": chats})
	}))

	mux.HandleFunc("/api/messages", withJSON(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := mustSession(store, w, r)
		if !ok {
			return
		}

		chatID := strings.TrimSpace(r.URL.Query().Get("chat_id"))
		if chatID == "" {
			writeError(w, http.StatusBadRequest, "chat_id is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := sess.Client.Messages.Search(ctx, resources.MessageSearchParams{
			ChatIDs:   []string{chatID},
			Limit:     beeperdesktop.IntPtr(50),
			Direction: beeperdesktop.StringPtr("after"),
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to load messages: %v", err))
			return
		}

		var messages []map[string]any
		for _, msg := range resp.Items {
			text := ""
			if msg.Text != nil {
				text = *msg.Text
			}
			senderName := msg.SenderID
			if msg.SenderName != nil && *msg.SenderName != "" {
				senderName = *msg.SenderName
			}
			messages = append(messages, map[string]any{
				"id":         msg.ID,
				"senderID":   msg.SenderID,
				"senderName": senderName,
				"timestamp":  msg.Timestamp.Format(time.RFC3339),
				"text":       text,
				"isSender":   msg.IsSender != nil && *msg.IsSender,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
	}))

	mux.HandleFunc("/api/messages/send", withJSON(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := mustSession(store, w, r)
		if !ok {
			return
		}

		var payload struct {
			ChatID string `json:"chatId"`
			Text   string `json:"text"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		payload.ChatID = strings.TrimSpace(payload.ChatID)
		payload.Text = strings.TrimSpace(payload.Text)

		if payload.ChatID == "" || payload.Text == "" {
			writeError(w, http.StatusBadRequest, "chatId and text are required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := sess.Client.Messages.Send(ctx, resources.MessageSendParams{
			ChatID: payload.ChatID,
			Text:   payload.Text,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to send message: %v", err))
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": resp.Success,
			"message": resp.MessageID,
		})
	}))

	log.Printf("Starting Beeper web chat on %s", addr)
	if err := http.ListenAndServe(addr, loggingMiddleware(mux)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func serveIndex(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

func readEnv(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func generateSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("sess-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error":   true,
		"message": message,
	})
}

func withJSON(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if ct := r.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "application/json") {
				writeError(w, http.StatusUnsupportedMediaType, "expected application/json content type")
				return
			}
		}
		handler(w, r)
	}
}

func readSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func fetchSession(store *sessionStore, r *http.Request) (*sessionData, string) {
	id := readSessionCookie(r)
	if id == "" {
		return nil, ""
	}
	data, ok := store.get(id)
	if !ok {
		return nil, id
	}
	return data, id
}

func mustSession(store *sessionStore, w http.ResponseWriter, r *http.Request) (*sessionData, bool) {
	sess, sessID := fetchSession(store, r)
	if sess == nil {
		if sessID != "" {
			store.delete(sessID)
		}
		writeError(w, http.StatusUnauthorized, "authentication required")
		return nil, false
	}
	return sess, true
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
