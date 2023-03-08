package controller

import "github.com/dezswap/dezswap-api/api/service"

type responsible interface{}

type mapServiceDtoToRes[T service.Gettable, K responsible] interface {
	toRes(T) K
}
