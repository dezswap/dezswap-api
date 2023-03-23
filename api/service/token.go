package service

import (
	"github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type tokenService struct {
	chainId string
	*gorm.DB
}

var _ Getter[Token] = &tokenService{}

func NewTokenService(chainId string, db *gorm.DB) Getter[Token] {
	return &tokenService{chainId, db}
}

// Get implements Getter
func (s *tokenService) Get(key string) (*Token, error) {
	token := &indexer.Token{}
	if err := s.Model(&indexer.Token{}).Where("chain_id = ? and address = ?", s.chainId, key).Omit("id,created_at,updated_at,deleted_at").Find(token).Error; err != nil {
		return nil, errors.Wrap(err, "TokenService.Get")
	}

	if token.Address != key {
		token = nil
	}

	return token, nil
}

// GetAll implements Getter
func (s *tokenService) GetAll() ([]Token, error) {
	tokens := []indexer.Token{}
	if err := s.Model(&indexer.Token{}).Where("chain_id = ?", s.chainId).Omit("id,created_at,updated_at,deleted_at").Find(&tokens).Error; err != nil {
		return nil, errors.Wrap(err, "TokenService.GetAll")
	}

	return tokens, nil
}
