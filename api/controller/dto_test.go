package controller

import (
	"encoding/json"
	"testing"

	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

func Test_dto(t *testing.T) {
	res := PoolRes{
		&dezswap.PoolRes{},
	}

	data, err := json.Marshal(res)
	if err != nil {
		println(err)
	}
	println(string(data))
}
