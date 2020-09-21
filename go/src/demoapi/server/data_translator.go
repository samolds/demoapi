package server

import (
	"net/url"

	"demoapi/database"
)

func apiUser(m *database.User, groups []*database.Group) *User {
	d := &User{
		ID:        m.Id,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Created:   UnixTS(m.Created),
	}

	if groups != nil && len(groups) > 0 {
		d.Groups = make([]Membership, 0, len(groups))
		for _, g := range groups {
			d.Groups = append(d.Groups, apiMembership(g.Name))
		}
	}

	return d
}

func apiGroup(m *database.Group, users []*database.User) *Group {
	d := &Group{
		Name:    m.Name,
		Created: UnixTS(m.Created),
	}

	if users != nil && len(users) > 0 {
		d.Users = make([]Membership, 0, len(users))
		for _, u := range users {
			d.Users = append(d.Users, apiMembership(u.Id))
		}
	}

	return d
}

func apiMembership(m string) Membership {
	return Membership(m)
}

func apiNextPage(requestedURL *url.URL, token string) *Page {
	if token == "" {
		return nil
	}

	requestedURL.RawQuery = "token=" + token
	link := requestedURL.String()
	return &Page{Link: link, Token: token}
}

func apiUsers(ms []*database.User) []*User {
	s := make([]*User, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiUser(m, nil))
	}
	return s
}

func apiGroups(ms []*database.Group) []*Group {
	s := make([]*Group, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiGroup(m, nil))
	}
	return s
}

func apiMembers(ms []*database.User) []Membership {
	s := make([]Membership, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiMembership(m.Id))
	}
	return s
}

func parseMembership(ms []Membership) []string {
	s := make([]string, 0, len(ms))
	for _, m := range ms {
		s = append(s, string(m))
	}
	return s
}
