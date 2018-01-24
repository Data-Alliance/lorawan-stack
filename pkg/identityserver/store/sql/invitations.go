// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
)

// InvitationStore implements store.InvitationStore.
type InvitationStore struct {
	storer
}

// NewInvitationStore creates a new invitation store.
func NewInvitationStore(store storer) *InvitationStore {
	return &InvitationStore{
		storer: store,
	}
}

// Save saves the invitation.
func (s *InvitationStore) Save(data *store.InvitationData) error {
	return s.save(s.queryer(), data)
}

func (s *InvitationStore) save(q db.QueryContext, data *store.InvitationData) error {
	_, err := q.NamedExec(
		`INSERT
			INTO invitations (
				token,
				email,
				issued_at,
				expires_at
			)
			VALUES (
				:token,
				:email,
				:issued_at,
				:expires_at
			)
			ON CONFLICT (email)
			DO UPDATE SET
				token = excluded.token,
				issued_at = excluded.issued_at,
				expires_at = excluded.expires_at
		`, data)
	return err
}

// List lists all the saved invitations.
func (s *InvitationStore) List() ([]*store.InvitationData, error) {
	return s.list(s.queryer())
}

func (s *InvitationStore) list(q db.QueryContext) ([]*store.InvitationData, error) {
	var invitations []*store.InvitationData
	err := q.Select(
		&invitations,
		`SELECT
				token,
				email,
				issued_at,
				expires_at
			FROM invitations`)
	if err != nil {
		return nil, err
	}

	if invitations == nil || len(invitations) == 0 {
		return make([]*store.InvitationData, 0), nil
	}

	return invitations, nil
}

// Use deletes an invitation but also takes into account the token binded to it.
func (s *InvitationStore) Use(email, token string) error {
	return s.use(s.queryer(), email, token)
}

func (s *InvitationStore) use(q db.QueryContext, email, token string) error {
	id := ""
	err := q.SelectOne(
		&id,
		`DELETE
			FROM invitations
			WHERE email = $1 AND token = $2
			RETURNING id`,
		email,
		token)
	if db.IsNoRows(err) {
		return ErrInvitationNotFound.New(nil)
	}
	return err
}

// Delete deletes an invitation by its ID.
func (s *InvitationStore) Delete(email string) error {
	return s.delete(s.queryer(), email)
}

func (s *InvitationStore) delete(q db.QueryContext, email string) error {
	_, err := q.Exec(`DELETE FROM invitations WHERE email = $1`, email)
	return err
}