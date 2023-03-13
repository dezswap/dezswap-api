package controller

import "github.com/gin-gonic/gin"

type PairController interface {
	Pairs(ctx *gin.Context)
	Pair(ctx *gin.Context)
}

type PoolController interface {
	Pools(ctx *gin.Context)
	Pool(ctx *gin.Context)
}

type TokenController interface {
	Tokens(ctx *gin.Context)
	Token(ctx *gin.Context)
}
