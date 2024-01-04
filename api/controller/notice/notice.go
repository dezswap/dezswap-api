package notice

import (
	"net/http"

	"github.com/dezswap/dezswap-api/api/service/notice"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type noticeController struct {
	s      notice.Notice
	logger logging.Logger
	mapper
}

func InitNoticeController(s notice.Notice, route *gin.RouterGroup, logger logging.Logger) *noticeController {
	c := noticeController{
		s, logger, mapper{},
	}
	c.logger.Debug("InitNoticeController")
	c.register(route)
	return &c
}

func (c *noticeController) register(route *gin.RouterGroup) {
	route.GET("", c.Notices)
}

// Dashboard godoc
//
//		@Summary		Notices of the chain
//		@Description	get Notices
//		@Tags			notice
//		@Accept			json
//		@Produce		json
//		@Success		200	{object}	NoticesRes
//		@Failure		400	{object}	httputil.BadRequestError
//		@Failure		500	{object}	httputil.InternalServerError
//
//	 	@Param			chain			query	string	false	"target chain name e.g. (dimension, cube)"
//	 	@Param			after			query	uint	false	"condition to get items after the id"
//	 	@Param			limit			query	uint	false	"the number of items to return (default: 10)"
//	 	@Param			asc				query	bool	false	"order of items to return (default: descending order)"
//		@Router			/notices [get]
func (c *noticeController) Notices(ctx *gin.Context) {
	reqParams := PaginationReq{}.Default()
	if err := ctx.Bind(&reqParams); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, errors.Wrap(err, "bad request"))
		return
	}

	notices, err := c.s.Notices(reqParams.Chain, reqParams.ToCondition())
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.noticesToRes(notices))
}
