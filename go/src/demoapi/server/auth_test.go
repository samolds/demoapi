package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	t.newUser(ctx, "user1")

	t.server.Config.InsecureRequestsMode = true
	t.server.router = router(t.server) // remount router with config change

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/users", nil)
	t.server.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := testResponse{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, len(resp.Users), 1)
	assert.Equal(t, resp.Users[0].ID, "user1")

	t.server.Config.InsecureRequestsMode = false
	t.server.router = router(t.server) // remount router with config change

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/users", nil)
	t.server.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp = testResponse{}
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, len(resp.Users), 0)
	assert.Equal(t, resp.Error, "unauthenticated: no authorization header")
}
