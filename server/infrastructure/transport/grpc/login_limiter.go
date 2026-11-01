package grpc

import (
	"sync"
	"time"
)

const (
	// loginMaxFailures is how many failed logins for a username are allowed within loginFailureWindow.
	loginMaxFailures = 5
	// loginFailureWindow is how long failed logins are remembered, and so how long a locked username stays locked.
	loginFailureWindow = time.Minute * 15
	// loginLimiterPruneSize is the amount of tracked usernames above which expired entries are pruned.
	loginLimiterPruneSize = 1024
)

type (
	// loginLimiter locks a username out after too many failed logins, to slow down password guessing.
	loginLimiter struct {
		lock     sync.Mutex
		failures map[string]*loginFailures
		now      func() time.Time
	}

	loginFailures struct {
		count int
		first time.Time
	}
)

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{
		failures: make(map[string]*loginFailures),
		now:      time.Now,
	}
}

// Allowed returns whenever or not the username may attempt to log in.
func (l *loginLimiter) Allowed(username string) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	f, ok := l.failures[username]
	if !ok {
		return true
	}
	if l.expired(f) {
		delete(l.failures, username)

		return true
	}

	return f.count < loginMaxFailures
}

// Failed records a failed login for the username.
func (l *loginLimiter) Failed(username string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if len(l.failures) > loginLimiterPruneSize {
		l.prune()
	}

	f, ok := l.failures[username]
	if !ok || l.expired(f) {
		f = &loginFailures{first: l.now()}
		l.failures[username] = f
	}
	f.count++
}

// Succeeded forgets the failed logins of the username.
func (l *loginLimiter) Succeeded(username string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	delete(l.failures, username)
}

func (l *loginLimiter) expired(f *loginFailures) bool {
	return l.now().Sub(f.first) > loginFailureWindow
}

func (l *loginLimiter) prune() {
	for k, f := range l.failures {
		if l.expired(f) {
			delete(l.failures, k)
		}
	}
}
