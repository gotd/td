package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// SearchResult is Search query result.
type SearchResult struct {
	MyResults []Peer
	Results   []Peer
}

// Search searches peers by given query.
func (m *Manager) Search(ctx context.Context, q string) (SearchResult, error) {
	convert := func(input []tg.PeerClass) ([]Peer, error) {
		resolved := make([]Peer, len(input))
		for i, p := range input {
			r, err := m.ResolvePeer(ctx, p)
			if err != nil {
				return nil, errors.Wrapf(err, "resolve %d (%+v)", i, p)
			}
			resolved[i] = r
		}
		return resolved, nil
	}

	found, err := m.api.ContactsSearch(ctx, &tg.ContactsSearchRequest{
		Q:     q,
		Limit: 10,
	})
	if err != nil {
		return SearchResult{}, errors.Wrap(err, "search")
	}

	if err := m.applyEntities(ctx, found.Users, found.Chats); err != nil {
		return SearchResult{}, err
	}

	var r SearchResult

	r.MyResults, err = convert(found.MyResults)
	if err != nil {
		return SearchResult{}, errors.Wrap(err, "my results")
	}
	r.Results, err = convert(found.Results)
	if err != nil {
		return SearchResult{}, errors.Wrap(err, "results")
	}

	return r, nil
}
