package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"demoapi/database"
	he "demoapi/httperror"
	monitor "demoapi/prometheus"
	"demoapi/util"
)

const (
	PaginationLimit = 50
)

// Health is a simple endpoint that can be used to help determine server health
// `GET /`
func (s *Server) Health(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {
	return map[string]string{"health": "okay!"}, nil
}

// GetUser returns the matching user record or 404 if none exist.
// `GET /users/<userID>`
func (s *Server) GetUser(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return nil, he.BadRequest.New("incomplete path. missing userID")
	}

	user, err := s.DB.Find_User_By_Id(ctx, database.User_Id(userID))
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, he.NotFound.New("userID %q doesn't exist", userID)
	}

	groups, err := s.DB.All_Group_By_User_Id(ctx, database.User_Id(userID))
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		User: apiUser(user, groups),
	}

	return resp, nil
}

// CreateUser creates a new user record. The body of the request should be a
// valid user record.
// `POST /users`
func (s *Server) CreateUser(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	userJSON := User{}
	err := json.NewDecoder(r.Body).Decode(&userJSON)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	if userJSON.FirstName == "" || userJSON.LastName == "" || userJSON.ID == "" {
		return nil, he.BadRequest.New("required fields missing")
	}

	user, err := s.DB.Create_User(ctx,
		database.User_Uuid(util.MustUUID4()),
		database.User_Id(userJSON.ID),
		database.User_FirstName(userJSON.FirstName),
		database.User_LastName(userJSON.LastName))
	if err != nil {
		// TODO(sam): catch unique constaint violations and return
		// 400 instead of 500
		return nil, err
	}

	monitor.UserGauge.Inc()

	// TODO(sam): do this transactionally with the previous query
	if len(userJSON.Groups) > 0 {
		added, removed, unchanged, err := s.DB.SetUserMembership(ctx, user.Id,
			parseMembership(userJSON.Groups))
		if err != nil {
			return nil, err
		}

		logrus.Debugf("memberships - added: %d, removed: %d, unchanged: %d", added,
			removed, unchanged)

		monitor.MembershipGauge.Add(float64(added))
		monitor.MembershipGauge.Sub(float64(removed))
	}

	groups, err := s.DB.All_Group_By_User_Id(ctx, database.User_Id(user.Id))
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		User: apiUser(user, groups),
	}

	return resp, nil
}

// DeleteUser deletes a user record. Returns 404 if the user doesn't exist.
// `DELETE /users/<userID>`
func (s *Server) DeleteUser(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return nil, he.BadRequest.New("incomplete path. missing userID")
	}

	deleted, err := s.DB.Delete_User_By_Id(ctx, database.User_Id(userID))
	if err != nil {
		return nil, err
	}

	if !deleted {
		return nil, he.NotFound.New("userID %q doesn't exist", userID)
	}

	monitor.UserGauge.Dec()

	return nil, nil
}

// UpdateUser updates an existing user record. The body of the request should
// be a valid user record. PUTs to a non-existent user should return a 404.
// `PUT /users/<userID>`
func (s *Server) UpdateUser(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	userJSON := User{}
	err := json.NewDecoder(r.Body).Decode(&userJSON)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	updateUserData, updateMemberships := false, len(userJSON.Groups) > 0
	userUpdates := database.User_Update_Fields{}

	if userJSON.ID != "" {
		userUpdates.Id = database.User_Id(userJSON.ID)
		updateUserData = true
	}

	if userJSON.FirstName != "" {
		userUpdates.FirstName = database.User_FirstName(userJSON.FirstName)
		updateUserData = true
	}

	if userJSON.LastName != "" {
		userUpdates.LastName = database.User_LastName(userJSON.LastName)
		updateUserData = true
	}

	if !updateUserData && !updateMemberships {
		return nil, he.BadRequest.New("no updates in request")
	}

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return nil, he.BadRequest.New("incomplete path. missing userID")
	}

	user, err := s.DB.Find_User_By_Id(ctx, database.User_Id(userID))
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, he.NotFound.New("userID %q doesn't exist", userID)
	}

	// TODO(sam): figure out how to extract "emptyUpdateError" from orm so that
	// a separate DB call isn't necessary. don't use dbx?
	if updateUserData {
		user, err = s.DB.Update_User_By_Id(ctx, database.User_Id(userID),
			userUpdates)
		if err != nil {
			return nil, err
		}
	}

	// TODO(sam): do this transactionally with the previous query
	added, removed, unchanged, err := s.DB.SetUserMembership(ctx, user.Id,
		parseMembership(userJSON.Groups))
	if err != nil {
		return nil, err
	}

	logrus.Debugf("memberships - added: %d, removed: %d, unchanged: %d", added,
		removed, unchanged)

	monitor.MembershipGauge.Add(float64(added))
	monitor.MembershipGauge.Sub(float64(removed))

	groups, err := s.DB.All_Group_By_User_Id(ctx, database.User_Id(user.Id))
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		User: apiUser(user, groups),
	}

	return resp, nil
}

// GetMemberships returns a JSON list of user ids containing the members of
// that group. Should return a 404 if the group doesn't exist.
// `GET /groups/<groupName>`
func (s *Server) GetMemberships(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	groupName := chi.URLParam(r, "groupName")
	if groupName == "" {
		return nil, he.BadRequest.New("incomplete path. missing groupName")
	}

	groupExists, err := s.DB.Has_Group_By_Name(ctx,
		database.Group_Name(groupName))
	if err != nil {
		return nil, err
	}

	if !groupExists {
		return nil, he.NotFound.New("groupName %q doesn't exist", groupName)
	}

	// this could return an empty set
	users, err := s.DB.All_User_By_Group_Name(ctx, database.Group_Name(groupName))
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		Members: apiMembers(users),
	}

	return resp, nil
}

// CreateGroup creates an empty group. POSTs to an existing group should be
// treated as errors and flagged with the appropriate HTTP status code. The
// body should contain a `name` parameter
// `POST /groups`
func (s *Server) CreateGroup(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	groupJSON := Group{}
	err := json.NewDecoder(r.Body).Decode(&groupJSON)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	if groupJSON.Name == "" {
		return nil, he.BadRequest.New("required fields missing")
	}

	// database enforces uniqueness constraint on group name
	group, err := s.DB.Create_Group(ctx,
		database.Group_Uuid(util.MustUUID4()),
		database.Group_Name(groupJSON.Name))
	if err != nil {
		// TODO(sam): catch unique constaint violations and return
		// 400 instead of 500
		return nil, err
	}

	monitor.GroupGauge.Inc()

	resp := &RootJSON{
		Group: apiGroup(group, nil),
	}

	return resp, nil
}

// UpdateMembership updates the membership list for the group. The body of the
// request should be a JSON list describing the group's members.
// `PUT /groups/<groupName>`
func (s *Server) UpdateMembership(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	groupName := chi.URLParam(r, "groupName")
	if groupName == "" {
		return nil, he.BadRequest.New("incomplete path. missing groupName")
	}

	type members struct {
		UserIDs []string `json:"userids"`
	}

	membersJSON := members{}
	err := json.NewDecoder(r.Body).Decode(&membersJSON)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	added, removed, unchanged, err := s.DB.SetGroupMembership(ctx, groupName,
		membersJSON.UserIDs)
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	logrus.Debugf("memberships - added: %d, removed: %d, unchanged: %d", added,
		removed, unchanged)

	monitor.MembershipGauge.Add(float64(added))
	monitor.MembershipGauge.Sub(float64(removed))

	return nil, nil
}

// DeleteGroup deletes a group.
// `DELETE /groups/<groupName>`
func (s *Server) DeleteGroup(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {
	return nil, nil

	groupName := chi.URLParam(r, "groupName")
	if groupName == "" {
		return nil, he.BadRequest.New("incomplete path. missing groupName")
	}

	deleted, err := s.DB.Delete_Group_By_Name(ctx,
		database.Group_Name(groupName))
	if err != nil {
		return nil, err
	}

	if !deleted {
		return nil, he.NotFound.New("groupName %q doesn't exist", groupName)
	}

	monitor.GroupGauge.Dec()

	return nil, nil
}

// getPaginationLimit will get the limit provided in the query parameter and
// use that, unless it's more than the Limit const hardcoded above.
func getPaginationLimit(queryParams url.Values, queryKey string) (int, error) {
	pageLimitStr := queryParams.Get(queryKey)
	limit := PaginationLimit
	if pageLimitStr != "" {
		pageLimit, err := strconv.Atoi(pageLimitStr)
		if err != nil {
			return 0, err
		}

		limit = util.Min(limit, pageLimit)
	}
	return limit, nil
}

// PagedUsers returns all possible users with pagination
// `GET /users?token=231&limit=20`
func (s *Server) PagedUsers(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	queryParams := r.URL.Query()
	token := queryParams.Get("token")
	limit, err := getPaginationLimit(queryParams, "limit")
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	users, nextToken, err := s.DB.Paged_User(ctx, limit, token)
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		Users:    apiUsers(users),
		NextPage: apiNextPage(r.URL, nextToken),
	}

	return resp, nil
}

// PagedGroups returns all possible groups with pagination
// `GET /groups?token=231&limit=20`
func (s *Server) PagedGroups(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	queryParams := r.URL.Query()
	token := queryParams.Get("token")
	limit, err := getPaginationLimit(queryParams, "limit")
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	groups, nextToken, err := s.DB.Paged_Group(ctx, limit, token)
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		Groups:   apiGroups(groups),
		NextPage: apiNextPage(r.URL, nextToken),
	}

	return resp, nil
}
