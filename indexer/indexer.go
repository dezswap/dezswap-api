package indexer

import (
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/db"
	"github.com/pkg/errors"
)

// tokenSaveBatchSize bounds how many freshly fetched tokens are held in memory before
// they are persisted. Keeping it small means an interrupted pass still makes progress.
const tokenSaveBatchSize = 100

type dexIndexer struct {
	pkg.NetworkMetadata
	repo    Repo
	chainId string
}

var _ Indexer = &dexIndexer{}

func NewDexIndexer(networkMetadata pkg.NetworkMetadata, repo Repo, chainId string) Indexer {
	return &dexIndexer{networkMetadata, repo, chainId}
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

	pools, err := d.repo.LatestPools()
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}

	poolMap := make(map[string]PoolInfo)
	for _, p := range pools {
		poolMap[p.Address] = p
	}

	for _, p := range pairs {
		poolInfo, err := d.repo.PoolFromNode(p.Address, height)
		if err != nil {
			return errors.Wrap(err, "dexIndexer.UpdateLatestPools")
		}

		pool, ok := poolMap[p.Address]
		if !ok || !isEqual(&pool, poolInfo) {
			poolInfo.Lp = p.Lp
			poolInfos = append(poolInfos, *poolInfo)
		}
	}

	if err := d.repo.SaveLatestPools(poolInfos, height); err != nil {
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
	tokens, err := d.repo.TokenAddresses(db.LastIdLimitCondition{})
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokensInDB := make(map[string]bool)
	for _, t := range tokens {
		tokensInDB[t] = true
	}

	newTokens := make([]Token, 0, tokenSaveBatchSize)
	flush := func() error {
		if len(newTokens) == 0 {
			return nil
		}
		if err := d.repo.SaveTokens(newTokens); err != nil {
			return errors.Wrap(err, "dexIndexer.UpdateTokens")
		}
		// A fresh buffer, so the batch just handed over is never written through.
		newTokens = make([]Token, 0, tokenSaveBatchSize)
		return nil
	}

	var failedCount uint
	var firstErr error
	var firstFailedAddr string

	for _, p := range pairs {
		for _, addr := range []string{p.Asset0, p.Asset1, p.Lp} {
			if _, ok := tokensInDB[addr]; ok {
				continue
			}
			// Mark before querying so one address is attempted at most once per pass. The
			// map is rebuilt from the DB on the next run, so failures are retried then.
			tokensInDB[addr] = true

			token, err := d.repo.TokenFromNode(addr)
			if err != nil {
				// One unresolvable token must not discard the rest of the pass.
				failedCount++
				if firstErr == nil {
					firstErr, firstFailedAddr = err, addr
				}
				continue
			}

			newTokens = append(newTokens, *token)
			if len(newTokens) >= tokenSaveBatchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		}
	}

	if err := flush(); err != nil {
		return err
	}

	if firstErr != nil {
		return errors.Wrapf(firstErr, "dexIndexer.UpdateTokens: %d token(s) failed, first failure on %s", failedCount, firstFailedAddr)
	}

	return nil
}

// UpdateVerifiedTokens implements Indexer
func (d *dexIndexer) UpdateVerifiedTokens() error {
	tokens, err := d.repo.Tokens(db.LastIdLimitCondition{})
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	newVerifiedTokens, err := d.repo.VerifiedTokens(d.chainId)
	if err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateTokens")
	}
	tokenMap := make(map[string]Token)
	orgVerifiedTokenMap := make(map[string]Token)
	for _, t := range tokens {
		tokenMap[t.Address] = t
		if t.Verified {
			orgVerifiedTokenMap[t.Address] = t
		}
	}

	updatableTokens := []Token{}
	for _, vt := range newVerifiedTokens {
		t, ok := tokenMap[vt.Address]
		if !ok || !isEqual(&t, &vt) {
			vt.ID = t.ID
			updatableTokens = append(updatableTokens, vt)
			tokenMap[vt.Address] = vt
		}
		delete(orgVerifiedTokenMap, vt.Address)
	}

	for _, t := range orgVerifiedTokenMap {
		t.Verified = false
		updatableTokens = append(updatableTokens, t)
	}

	if err := d.repo.SaveTokens(updatableTokens); err != nil {
		return errors.Wrap(err, "dexIndexer.UpdateVerifiedTokens")
	}
	return nil
}
