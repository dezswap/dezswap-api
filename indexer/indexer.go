package indexer

import (
	"github.com/dezswap/dezswap-api/pkg/db"
	"github.com/pkg/errors"
)

type dexIndexer struct {
	repo    Repo
	chainId string
}

var _ Indexer = &dexIndexer{}

func NewDexIndexer(repo Repo, chainId string) Indexer {
	return &dexIndexer{repo, chainId}
}

// UpdatePools implements Indexer
func (d *dexIndexer) UpdateLatestPools() error {
	pairs, err := d.repo.Pairs(db.LastIdLimitCondition{})
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	poolInfos := []PoolInfo{}

	height, err := d.repo.LatestHeightFromNode()
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}

	for _, p := range pairs {
		poolInfo, err := d.repo.PoolFromNode(p.Address, height)
		if err != nil {
			return errors.Wrap(err, "dexIndexer.UpdateLatestPools")
		}
		poolInfos = append(poolInfos, *poolInfo)
	}

	if err := d.repo.SavePools(poolInfos, height); err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateLatestPools")
	}
	return nil
}

// UpdateTokens implements Indexer
func (d *dexIndexer) UpdateTokens() error {
	pairs, err := d.repo.Pairs(db.LastIdLimitCondition{})
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokens, err := d.repo.Tokens(db.LastIdLimitCondition{})
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokenMap := make(map[string]*Token)
	for _, t := range tokens {
		tokenMap[t.Address] = &t
	}

	newTokens := []Token{}
	for _, p := range pairs {
		for _, addr := range []string{p.Asset0, p.Asset1, p.Lp} {
			if _, ok := tokenMap[addr]; ok {
				continue
			}
			t, err := d.repo.TokenFromNode(addr)
			if err != nil {
				return errors.Wrap(err, "dexIndexer.UpdateTokens")
			}
			newTokens = append(newTokens, *t)
			tokenMap[addr] = t
		}
	}

	if err := d.repo.SaveTokens(newTokens); err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}

	return nil
}

// UpdateVerifiedTokens implements Indexer
func (d *dexIndexer) UpdateVerifiedTokens() error {
	tokens, err := d.repo.Tokens(db.LastIdLimitCondition{})
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	verifiedTokens, err := d.repo.VerifiedTokens(d.chainId)
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokenMap := make(map[string]*Token)
	for _, t := range tokens {
		tokenMap[t.Address] = &t
	}

	updatableTokens := []Token{}
	for _, vt := range verifiedTokens {
		t, ok := tokenMap[vt.Address]
		if !ok || !isEqual(t, &vt) {
			updatableTokens = append(updatableTokens, vt)
		}
	}

	if err := d.repo.SaveTokens(updatableTokens); err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateVerifiedTokens")
	}
	return nil
}
