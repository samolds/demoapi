package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"demoapi/config"
	"demoapi/database"
	"demoapi/util"
)

// common tools used when testing the server

type testResponse struct {
	RootJSON
	Error    string `json:"error"`
	Health   string `json:"health"`
	Response string `json:"response"`
}

type serverTest struct {
	*testing.T
	server *Server
}

func newServerTest(t *testing.T) (context.Context, *serverTest) {
	// https://www.sqlite.org/inmemorydb.html
	testDBURL, err := url.Parse("sqlite3::memory:")
	assert.NoError(t, err)
	testDB, err := database.Connect(testDBURL, nil)
	assert.NoError(t, err)
	c := &config.Configs{
		InsecureRequestsMode: true,
	}
	return context.Background(), &serverTest{
		T:      t,
		server: New(testDB, c),
	}
}

func (st *serverTest) cleanup() {
	assert.NoError(st, st.server.DB.Close())
}

func (st *serverTest) newUser(ctx context.Context, id string) string {
	user, err := st.server.DB.Create_User(ctx,
		database.User_Uuid(util.MustUUID4()),
		database.User_Id(id),
		database.User_FirstName(id+"first_name"),
		database.User_LastName(id+"last_name"))
	assert.NoError(st, err)
	return user.Id
}

func (st *serverTest) newGroup(ctx context.Context, name string) string {
	group, err := st.server.DB.Create_Group(ctx,
		database.Group_Uuid(util.MustUUID4()),
		database.Group_Name(name))
	assert.NoError(st, err)
	return group.Name
}

func (st *serverTest) newMembership(ctx context.Context, userID,
	groupName string) {

	u, err := st.server.DB.Get_User_By_Id(ctx, database.User_Id(userID))
	assert.NoError(st, err)
	g, err := st.server.DB.Get_Group_By_Name(ctx, database.Group_Name(groupName))
	assert.NoError(st, err)
	_, err = st.server.DB.Create_Membership(ctx, database.Membership_UserPk(u.Pk),
		database.Membership_GroupPk(g.Pk))
	assert.NoError(st, err)
}

func jsonRequest(st *serverTest, method, target string,
	pathParams map[string]string, body interface{}) *http.Request {

	// https://github.com/go-chi/chi/issues/76
	rctx := chi.NewRouteContext()
	for k, v := range pathParams {
		rctx.URLParams.Add(k, v)
	}

	buf, err := json.Marshal(body)
	assert.NoError(st, err)
	r := httptest.NewRequest(method, target, bytes.NewReader(buf))
	r.Header.Set("Content-Type", "application/json")

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
