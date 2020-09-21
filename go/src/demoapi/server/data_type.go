package server

import (
	"fmt"
	"strconv"
	"time"
)

type RootJSON struct {
	User     *User        `json:"user,omitempty"`
	Group    *Group       `json:"group,omitempty"`
	Users    []*User      `json:"users,omitempty"`
	Groups   []*Group     `json:"groups,omitempty"`
	Members  []Membership `json:"members,omitempty"`
	NextPage *Page        `json:"next_page,omitempty"`
}

type User struct {
	ID        string       `json:"userid"`
	UUID      string       `json:"uuid,omitempty"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Created   UnixTime     `json:"created"`
	Groups    []Membership `json:"groups,omitempty"`
}

type Group struct {
	Name    string       `json:"name"`
	UUID    string       `json:"uuid,omitempty"`
	Created UnixTime     `json:"created"`
	Users   []Membership `json:"users,omitempty"`
}

// TODO(sam): this could contain additional information, like join time
type Membership string

type Page struct {
	Link  string `json:"link"`
	Token string `json:"token"`
}

type UnixTime struct {
	time.Time
}

func UnixTS(t time.Time) UnixTime { return UnixTime{Time: t} }

func (t UnixTime) MarshalJSON() ([]byte, error) {
	unixTime := t.Time.Unix()
	if t.Time.IsZero() || unixTime == 0 {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprint(unixTime)), nil
}

// UnmarshalJSON expects time to be int64 unix time stamps in seconds
func (t *UnixTime) UnmarshalJSON(ts []byte) (err error) {
	// ignore null, like the main json package
	st := string(ts)
	if st == "null" {
		return nil
	}

	// convert unix time string to time object
	unixSec, err := strconv.ParseInt(st, 10, 64)
	if err != nil {
		return err
	}
	*t = UnixTime{Time: time.Unix(unixSec, 0)}
	return nil
}
