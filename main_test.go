package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestUser struct {
}

func (u *TestUser) GetID() int {
	return 1
}

type TestRequestUserProvider struct {
}

func (auth *TestRequestUserProvider) GetUser(r *http.Request) (User[int], error) {
	return &TestUser{}, nil
}

type TestRoleChecker struct {
	inRole bool
}

func (auth *TestRoleChecker) UserInRole(user User[int]) (bool, error) {
	return auth.inRole, nil
}

func TestMiddleware(t *testing.T) {
	// Create server.
	var requestUserProvider RequestUserProvider[int] = &TestRequestUserProvider{}
	testRoleChecker := &TestRoleChecker{}
	var roleChecker RoleChecker[int] = testRoleChecker

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	wrapperMux := NewMiddleware(roleChecker, mux, requestUserProvider)

	s := httptest.NewServer(wrapperMux)

	testCases := []struct {
		name     string
		inRole   bool
		expected int
	}{
		{name: "HasPermissions", inRole: true, expected: http.StatusOK},
		{name: "NoPermissions", inRole: false, expected: http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testRoleChecker.inRole = tc.inRole
			res, err := s.Client().Get(s.URL)
			require.NoError(t, err)
			require.Equal(t, tc.expected, res.StatusCode)
		})
	}
}
