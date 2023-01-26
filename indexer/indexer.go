package indexer

import "github.com/pkg/errors"

type dexIndexer struct {
	repo Repo
}

var _ Indexer = &dexIndexer{}

func NewDexIndexer(repo Repo) Indexer {
	return &dexIndexer{repo}
}

// UpdatePools implements Indexer
func (d *dexIndexer) UpdateLatestPools() error {
	panic("unimplemented")

}

// UpdateTokens implements Indexer
func (d *dexIndexer) UpdateTokens() error {
	pairs, err := d.repo.Pairs()
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokens, err := d.repo.Tokens()
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokenMap := make(map[string]*Token)
	for _, t := range tokens {
		tokenMap[t.Address] = &t
	}

	for _, p := range pairs {
		for _, addr := range []string{p.Asset0, p.Asset1, p.Lp} {
			if _, ok := tokenMap[addr]; ok {
				continue
			}

			t, err := d.repo.Token(addr)
			if err != nil {
				return errors.Wrap(err, "dexIndexer.UpdateTokens")
			}
			tokenMap[addr] = t
		}
	}

	return nil
}
