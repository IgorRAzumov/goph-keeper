package fake

import (
	"bytes"
	"context"
	"sync"
	"time"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/session/model"
)

// SessionRepo — in-memory SessionRepository для тестов.
type SessionRepo struct {
	mu       sync.Mutex
	sessions map[string]*model.Session
}

// NewSessionRepo создаёт пустой in-memory репозиторий сессий.
func NewSessionRepo() *SessionRepo {
	return &SessionRepo{sessions: map[string]*model.Session{}}
}

func (r *SessionRepo) Save(_ context.Context, session *model.Session) error {
	if session == nil || session.ID == "" {
		return common.ErrInvalidInput
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	copySession := *session
	r.sessions[session.ID] = &copySession
	return nil
}

func (r *SessionRepo) Get(_ context.Context, id string) (*model.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	copySession := *s
	return &copySession, nil
}

func (r *SessionRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[id]; !ok {
		return common.ErrNotFound
	}
	delete(r.sessions, id)
	return nil
}

func (r *SessionRepo) Rotate(_ context.Context, oldSessionID string, newSession *model.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[oldSessionID]; !ok {
		return common.ErrNotFound
	}
	delete(r.sessions, oldSessionID)
	copySession := *newSession
	r.sessions[newSession.ID] = &copySession
	return nil
}

func (r *SessionRepo) FindActiveByRefreshToken(_ context.Context, refreshToken string, now time.Time) (*model.Session, error) {
	hash := model.HashRefreshToken(refreshToken)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.sessions {
		if bytes.Equal(s.RefreshTokenHash, hash) && now.Before(s.RefreshExpiresAt) {
			copySession := *s
			return &copySession, nil
		}
	}
	return nil, common.ErrNotFound
}
