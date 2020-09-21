package database

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

// This file contains any sql queries that are written by hand because they use
// functionality that is currently unsupported by DBX

// TODO(sam): find a better database layer. dbx is leaving me wanting. it's not
// very extensible.

// SetGroupMembership will remove any membership relationships that exist but
// aren't provided in userIDs, it will add any new membership relationships,
// and the intersection set will be untouched.
func (db *Database) SetGroupMembership(ctx context.Context, groupName string,
	userIDs []string) (int, int, int, error) { // added, removed, unchanged

	added, removed := 0, 0

	err := db.WithTx(ctx, func(ctx context.Context, tx *Tx) error {
		var err error

		// delete all memberships for the groupname that aren't listed in userIDs
		removed, err = db.DeleteMembershipNotListedForGroup(ctx, tx, groupName,
			userIDs)
		if err != nil {
			return err
		}

		// TODO(sam): validate that all of the provided userIDs actually exist
		// in the db

		// insert or ignore the remaining user ids as memberships
		added, err = db.InsertOrIgnoreMembershipToGroup(ctx, tx, groupName,
			userIDs)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logrus.Error(err)
		return 0, 0, 0, dbErr.Wrap(err)
	}

	return added, removed, len(userIDs) - added, nil
}

// DeleteMembershipNotListedForGroup will use a single query to delete all
// membership rows that are not included in the provided slice of userIDs, for
// a groupName. This is all done within the provided transaction.
func (db *Database) DeleteMembershipNotListedForGroup(ctx context.Context,
	tx *Tx, groupName string, userIDs []string) (int, error) {

	values := make([]interface{}, 0, 1+len(userIDs))
	values = append(values, groupName)
	for _, userID := range userIDs {
		values = append(values, userID)
	}

	// TODO(sam): user a string builder
	optJoin, optSuffix := "", ""
	if len(userIDs) > 0 {
		optJoin = "JOIN users ON users.pk = memberships.user_pk "
		optSuffix = "AND users.id NOT IN (?" +
			strings.Repeat(",?", len(userIDs)-1) + ")"
	}

	queryRaw := "DELETE FROM memberships WHERE memberships.pk IN (" +
		"SELECT memberships.pk FROM memberships " +
		"JOIN groups ON groups.pk = memberships.group_pk " +
		optJoin +
		"WHERE groups.name = ? " +
		optSuffix + ")"

	stmt := db.Rebind(queryRaw) // cleans up sql as needed per driver (eg ?->$1)
	logrus.Debugf("stmt: <%s>, values: <%v>", stmt, values)

	result, err := tx.Tx.ExecContext(ctx, stmt, values...)
	if err != nil {
		return 0, dbErr.Wrap(err)
	}

	removed, err := result.RowsAffected()
	if err != nil {
		return 0, dbErr.Wrap(err)
	}
	return int(removed), nil
}

// InsertOrIgnoreMembershipToGroup will insert memberships for all of the
// userIDs with the groupName that do not already exist in one query. This
// is all done within the provided transaction.
func (db *Database) InsertOrIgnoreMembershipToGroup(ctx context.Context,
	tx *Tx, groupName string, userIDs []string) (int, error) {

	if len(userIDs) == 0 {
		// nothing to do
		return 0, nil
	}

	// TODO(sam): use a string builder for all of this
	prefix, suffix := "", ""
	switch db.driver {
	case SqliteDriver:
		prefix = "INSERT OR IGNORE INTO"
	case PostgresDriver:
		prefix = "INSERT INTO"
		suffix = " ON CONFLICT DO NOTHING"
	default:
		return 0, dbErr.New("unsupported driver %q", db.driver)
	}

	parameters := "SELECT ?, (SELECT pk FROM groups WHERE groups.name = ?), " +
		"users.pk FROM users WHERE users.id IN (?" +
		strings.Repeat(",?", len(userIDs)-1) + ")"
	values := make([]interface{}, 0, 2+len(userIDs))
	values = append(values, db.Hooks.Now().UTC().UTC())
	values = append(values, groupName)
	for _, userID := range userIDs {
		values = append(values, userID)
	}

	queryRaw := prefix + " memberships ( created, group_pk, user_pk ) " +
		parameters + suffix
	stmt := db.Rebind(queryRaw) // cleans up sql as needed per driver (eg ?->$1)
	logrus.Debugf("stmt: <%s>, values: <%v>", stmt, values)

	result, err := tx.Tx.ExecContext(ctx, stmt, values...)
	if err != nil {
		return 0, dbErr.Wrap(err)
	}

	added, err := result.RowsAffected()
	if err != nil {
		return 0, dbErr.Wrap(err)
	}
	return int(added), nil
}

// SetUserMembership will remove any membership relationships that exist but
// aren't provided in groupNames, it will add any new membership relationships,
// and the intersection set will be untouched.
func (db *Database) SetUserMembership(ctx context.Context, userID string,
	groupNames []string) (int, int, int, error) { // added, removed, unchanged

	added, removed := 0, 0

	err := db.WithTx(ctx, func(ctx context.Context, tx *Tx) error {
		var err error

		// delete all memberships for the user that aren't listed in groupNames
		removed, err = db.DeleteMembershipNotListedForUser(ctx, tx, userID,
			groupNames)
		if err != nil {
			return err
		}

		// TODO(sam): validate that all of the provided groupNames actually exist
		// in the db

		// insert or ignore the remaining group names as memberships
		added, err = db.InsertOrIgnoreMembershipToUser(ctx, tx, userID,
			groupNames)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logrus.Error(err)
		return 0, 0, 0, dbErr.Wrap(err)
	}

	return added, removed, len(groupNames) - added, nil
}

// DeleteMembershipNotListedForUser will use a single query to delete all
// membership rows that are not included in the provided slice of groupNames,
// for a userID. This is all done within the provided transaction.
func (db *Database) DeleteMembershipNotListedForUser(ctx context.Context,
	tx *Tx, userID string, groupNames []string) (int, error) {

	values := make([]interface{}, 0, 1+len(groupNames))
	values = append(values, userID)
	for _, groupName := range groupNames {
		values = append(values, groupName)
	}

	// TODO(sam): user a string builder
	optJoin, optSuffix := "", ""
	if len(groupNames) > 0 {
		optJoin = "JOIN groups ON groups.pk = memberships.group_pk "
		optSuffix = "AND groups.name NOT IN (?" +
			strings.Repeat(",?", len(groupNames)-1) + ")"
	}

	queryRaw := "DELETE FROM memberships WHERE memberships.pk IN (" +
		"SELECT memberships.pk FROM memberships " +
		"JOIN users ON users.pk = memberships.user_pk " +
		optJoin +
		"WHERE users.id = ? " +
		optSuffix + ")"

	stmt := db.Rebind(queryRaw) // cleans up sql as needed per driver (eg ?->$1)
	logrus.Debugf("stmt: <%s>, values: <%v>", stmt, values)

	result, err := tx.Tx.ExecContext(ctx, stmt, values...)
	if err != nil {
		return 0, dbErr.Wrap(err)
	}

	removed, err := result.RowsAffected()
	if err != nil {
		return 0, dbErr.Wrap(err)
	}
	return int(removed), nil
}

// InsertOrIgnoreMembershipToUser will insert memberships for all of the
// groupNames with the userID that do not already exist in one query. This
// is all done within the provided transaction.
func (db *Database) InsertOrIgnoreMembershipToUser(ctx context.Context,
	tx *Tx, userID string, groupNames []string) (int, error) {

	if len(groupNames) == 0 {
		// nothing to do
		return 0, nil
	}

	// TODO(sam): use a string builder for all of this
	prefix, suffix := "", ""
	switch db.driver {
	case SqliteDriver:
		prefix = "INSERT OR IGNORE INTO"
	case PostgresDriver:
		prefix = "INSERT INTO"
		suffix = " ON CONFLICT DO NOTHING"
	default:
		return 0, dbErr.New("unsupported driver %q", db.driver)
	}

	parameters := "SELECT ?, (SELECT pk FROM users WHERE users.id = ?), " +
		"groups.pk FROM groups WHERE groups.name IN (?" +
		strings.Repeat(",?", len(groupNames)-1) + ")"
	values := make([]interface{}, 0, 2+len(groupNames))
	values = append(values, db.Hooks.Now().UTC().UTC())
	values = append(values, userID)
	for _, groupName := range groupNames {
		values = append(values, groupName)
	}

	queryRaw := prefix + " memberships ( created, user_pk, group_pk ) " +
		parameters + suffix
	stmt := db.Rebind(queryRaw) // cleans up sql as needed per driver (eg ?->$1)
	logrus.Debugf("stmt: <%s>, values: <%v>", stmt, values)

	result, err := tx.Tx.ExecContext(ctx, stmt, values...)
	if err != nil {
		return 0, dbErr.Wrap(err)
	}

	added, err := result.RowsAffected()
	if err != nil {
		return 0, dbErr.Wrap(err)
	}
	return int(added), nil
}
