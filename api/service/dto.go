package service

import (
	"github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/dezswap/dezswap-api/pkg/db/parser"
)

type Pair = parser.Pair

type Pool = indexer.LatestPool

type Token = indexer.Token
