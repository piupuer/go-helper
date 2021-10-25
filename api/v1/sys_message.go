package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

func FindMessage(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.MessageReq
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		r.ToUserId = u.UserId
		ops.addCtx(c)
		list := make([]resp.MessageResp, 0)
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			list = rd.FindUnDeleteMessage(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindUnDeleteMessage(&r)
		}
		resp.SuccessWithPageData(list, []resp.MessageResp{}, r.Page)
	}
}

func GetUnReadMessageCount(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		var total int64
		var err error
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			total, err = rd.GetUnReadMessageCount(u.UserId)
		default:
			my := query.NewMySql(ops.dbOps...)
			total, err = my.GetUnReadMessageCount(u.UserId)
		}
		resp.CheckErr(err)
		resp.SuccessWithData(total)
	}
}

func PushMessage(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.PushMessageReq
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		r.FromUserId = u.UserId
		err := q.CreateMessage(&r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func BatchUpdateMessageRead(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.BatchUpdateMessageRead(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}

func BatchUpdateMessageDeleted(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.BatchUpdateMessageDeleted(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateAllMessageRead(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		u := ops.getCurrentUser(c)
		err := q.UpdateAllMessageRead(u.UserId)
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateAllMessageDeleted(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		u := ops.getCurrentUser(c)
		err := q.UpdateAllMessageDeleted(u.UserId)
		resp.CheckErr(err)
		resp.Success()
	}
}
