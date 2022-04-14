package main

import (
	"errors"
	"net/http"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type User[T any] interface {
	GetID() T
}

type RoleChecker[T any] interface {
	UserInRole(user User[T]) (bool, error)
}

type ErrorHandler func(err error, w http.ResponseWriter, r *http.Request)

func DefaultErrorHandler(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusUnauthorized)
}

type RequestUserProvider[T any] interface {
	GetUser(r *http.Request) (User[T], error)
}

type Middleware[T any] struct {
	errorHandler        ErrorHandler
	handler             http.Handler
	roleChecker         RoleChecker[T]
	requestUserProvider RequestUserProvider[T]
}

func NewMiddleware[T any](
	roleChecker RoleChecker[T],
	handlerToWrap http.Handler,
	requestUserProvider RequestUserProvider[T],
) *Middleware[T] {
	return &Middleware[T]{
		handler:             handlerToWrap,
		roleChecker:         roleChecker,
		requestUserProvider: requestUserProvider,
	}
}

func (m *Middleware[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := m.requestUserProvider.GetUser(r)
	if err != nil {
		m.handleError(err, w, r)
		return
	}

	hasAccess, err := m.roleChecker.UserInRole(user)
	if err != nil {
		m.handleError(err, w, r)
		return
	}

	if !hasAccess {
		m.handleError(ErrUnauthorized, w, r)
		return
	}

	m.handler.ServeHTTP(w, r)
}

func (m *Middleware[T]) SetErrorHandler(handler ErrorHandler) {
	m.errorHandler = handler
}

func (m *Middleware[T]) handleError(err error, w http.ResponseWriter, r *http.Request) {
	if m.errorHandler != nil {
		m.errorHandler(err, w, r)
	} else {
		DefaultErrorHandler(err, w, r)
	}
}
