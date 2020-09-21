package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetGroupMembership tests the SetGroupMembership method. It is expected
// to update the membership rows in the db to exactly match the provided list
// of userIDs by either deleting or inserting rows. Intersecting rows should
// remain untouched
func TestSetGroupMembership(test *testing.T) {
	ctx, t := newDBTest(test)
	defer t.cleanup()

	t.newUser(ctx, "user1")
	t.newUser(ctx, "user2")
	t.newGroup(ctx, "group1")
	t.newGroup(ctx, "group2")

	add, del, noop, err := t.db.SetGroupMembership(ctx, "group1",
		[]string{"user1", "user2"})
	assert.NoError(t, err)
	assert.Equal(t, 2, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 0, noop)

	add, del, noop, err = t.db.SetGroupMembership(ctx, "group1",
		[]string{"user1", "user2"})
	assert.NoError(t, err)
	assert.Equal(t, 0, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 2, noop)

	add, del, noop, err = t.db.SetGroupMembership(ctx, "group1",
		[]string{"user1"})
	assert.NoError(t, err)
	assert.Equal(t, 0, add)
	assert.Equal(t, 1, del)
	assert.Equal(t, 1, noop)

	add, del, noop, err = t.db.SetGroupMembership(ctx, "group1",
		[]string{"user1", "user2"})
	assert.NoError(t, err)
	assert.Equal(t, 1, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 1, noop)

	add, del, noop, err = t.db.SetGroupMembership(ctx, "group2",
		[]string{"user1", "user2"})
	assert.NoError(t, err)
	assert.Equal(t, 2, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 0, noop)

	add, del, noop, err = t.db.SetGroupMembership(ctx, "group2", nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, add)
	assert.Equal(t, 2, del)
	assert.Equal(t, 0, noop)
}

// TestSetUserMembership tests the SetUserMembership method. It is expected
// to update the membership rows in the db to exactly match the provided list
// of groupNames by either deleting or inserting rows. Intersecting rows should
// remain untouched
func TestSetUserMembership(test *testing.T) {
	ctx, t := newDBTest(test)
	defer t.cleanup()

	t.newUser(ctx, "user1")
	t.newUser(ctx, "user2")
	t.newGroup(ctx, "group1")
	t.newGroup(ctx, "group2")

	add, del, noop, err := t.db.SetUserMembership(ctx, "user1",
		[]string{"group1", "group2"})
	assert.NoError(t, err)
	assert.Equal(t, 2, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 0, noop)

	add, del, noop, err = t.db.SetUserMembership(ctx, "user1",
		[]string{"group1", "group2"})
	assert.NoError(t, err)
	assert.Equal(t, 0, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 2, noop)

	add, del, noop, err = t.db.SetUserMembership(ctx, "user1",
		[]string{"group1"})
	assert.NoError(t, err)
	assert.Equal(t, 0, add)
	assert.Equal(t, 1, del)
	assert.Equal(t, 1, noop)

	add, del, noop, err = t.db.SetUserMembership(ctx, "user1",
		[]string{"group1", "group2"})
	assert.NoError(t, err)
	assert.Equal(t, 1, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 1, noop)

	add, del, noop, err = t.db.SetUserMembership(ctx, "user2",
		[]string{"group1", "group2"})
	assert.NoError(t, err)
	assert.Equal(t, 2, add)
	assert.Equal(t, 0, del)
	assert.Equal(t, 0, noop)

	add, del, noop, err = t.db.SetUserMembership(ctx, "user2", nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, add)
	assert.Equal(t, 2, del)
	assert.Equal(t, 0, noop)
}
