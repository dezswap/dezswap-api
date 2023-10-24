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

type TickerController interface {
	Tickers(ctx *gin.Context)
	Ticker(ctx *gin.Context)
}

type StatController interface {
	Stats(ctx *gin.Context)
	Stat(ctx *gin.Context)
}

type DashboardController interface{}
