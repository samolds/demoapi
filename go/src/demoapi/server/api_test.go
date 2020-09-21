package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"demoapi/database"
)

func TestHealth(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := t.server.Health(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(map[string]string)
	assert.True(t, ok)

	assert.Equal(t, json["health"], "okay!")
}

func TestGetUser(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	t.newUser(ctx, "user1")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, err := t.server.PagedUsers(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(*RootJSON)
	assert.True(t, ok)
	assert.Equal(t, len(json.Users), 1)
	assert.Equal(t, json.Users[0].ID, "user1")
}

func TestCreateUser(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	t.newGroup(ctx, "group1")

	w := httptest.NewRecorder()
	user := User{FirstName: "fn", LastName: "ln", ID: "user1",
		Groups: []Membership{"group1"}}
	r := jsonRequest(t, http.MethodPost, "/users", nil, user)
	resp, err := t.server.CreateUser(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(*RootJSON)
	assert.True(t, ok)
	assert.Equal(t, json.User.ID, "user1")
	assert.Equal(t, len(json.User.Groups), 1)
	assert.Equal(t, json.User.Groups[0], Membership("group1"))
}

func TestBulkJoin(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	t.newUser(ctx, "user1")
	t.newUser(ctx, "user2")
	t.newUser(ctx, "user3")
	t.newGroup(ctx, "group1")

	reqBody := make(map[string][]string)
	pathParams := make(map[string]string)
	pathParams["groupName"] = "group1"

	w := httptest.NewRecorder()
	reqBody["userids"] = []string{"user1", "user2"}
	r := jsonRequest(t, http.MethodPut, "/groups/group1", pathParams, reqBody)
	_, err := t.server.UpdateMembership(ctx, w, r)
	assert.NoError(t, err)

	users, err := t.server.DB.All_User_By_Group_Name(ctx,
		database.Group_Name("group1"))
	assert.NoError(t, err)
	assert.Equal(t, len(users), 2)

	w = httptest.NewRecorder()
	reqBody["userids"] = []string{"user3"}
	r = jsonRequest(t, http.MethodPut, "/groups/group1", pathParams, reqBody)
	_, err = t.server.UpdateMembership(ctx, w, r)
	assert.NoError(t, err)

	users, err = t.server.DB.All_User_By_Group_Name(ctx,
		database.Group_Name("group1"))
	assert.NoError(t, err)
	assert.Equal(t, len(users), 1)
	assert.Equal(t, users[0].Id, "user3")
}
