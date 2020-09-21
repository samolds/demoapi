package database

import (
	"context"
	"demoapi/util"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// common tools used when testing the database

type dbTest struct {
	*testing.T
	db *Database
}

func newDBTest(t *testing.T) (context.Context, *dbTest) {
	// https://www.sqlite.org/inmemorydb.html
	testDBURL, err := url.Parse("sqlite3::memory:")
	assert.NoError(t, err)
	testDB, err := Connect(testDBURL, nil)
	assert.NoError(t, err)
	return context.Background(), &dbTest{T: t, db: testDB}
}

func (dbt *dbTest) cleanup() {
	assert.NoError(dbt, dbt.db.Close())
}

func (dbt *dbTest) newUser(ctx context.Context, id string) string {
	user, err := dbt.db.Create_User(ctx,
		User_Uuid(util.MustUUID4()), User_Id(id),
		User_FirstName(id+"first_name"), User_LastName(id+"last_name"))
	assert.NoError(dbt, err)
	return user.Id
}

func (dbt *dbTest) newGroup(ctx context.Context, name string) string {
	group, err := dbt.db.Create_Group(ctx, Group_Uuid(util.MustUUID4()),
		Group_Name(name))
	assert.NoError(dbt, err)
	return group.Name
}

func (dbt *dbTest) newMembership(ctx context.Context, userID,
	groupName string) {

	u, err := dbt.db.Get_User_By_Id(ctx, User_Id(userID))
	assert.NoError(dbt, err)
	g, err := dbt.db.Get_Group_By_Name(ctx, Group_Name(groupName))
	assert.NoError(dbt, err)
	_, err = dbt.db.Create_Membership(ctx, Membership_UserPk(u.Pk),
		Membership_GroupPk(g.Pk))
	assert.NoError(dbt, err)
}
