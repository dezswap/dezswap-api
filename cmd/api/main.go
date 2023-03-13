package main

import (
	"github.com/dezswap/dezswap-api/api"
	"github.com/dezswap/dezswap-api/configs"
)

func main() {
	c := configs.New()
	api.RunServer(c)
}
