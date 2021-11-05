package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

func FindMessage(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	if ops.findUserByIds == nil {
		panic("findUserByIds is empty")
	}
	return func(c *gin.Context) {
		var r req.Message
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		r.ToUserId = u.Id
		ops.addCtx(c)
		list := make([]resp.Message, 0)
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			list = rd.FindUnDeleteMessage(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindUnDeleteMessage(&r)
		}
		userIds := make([]uint, 0)
		for _, item := range list {
			if !utils.ContainsUint(userIds, item.ToUserId) {
				userIds = append(userIds, item.ToUserId)
			}
			if !utils.ContainsUint(userIds, item.FromUserId) {
				userIds = append(userIds, item.FromUserId)
			}
		}
		users := ops.findUserByIds(c, userIds)
		for i, item := range list {
			for _, user := range users {
				if item.ToUserId == user.Id {
					list[i].ToUsername = user.Username
					list[i].ToNickname = user.Nickname
					break
				}
			}
			for _, user := range users {
				if item.FromUserId == user.Id {
					list[i].FromUsername = user.Username
					list[i].FromNickname = user.Nickname
					break
				}
			}
		}
		resp.SuccessWithPageData(list, &[]resp.Message{}, r.Page)
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
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			total, err = rd.GetUnReadMessageCount(u.Id)
		default:
			my := query.NewMySql(ops.dbOps...)
			total, err = my.GetUnReadMessageCount(u.Id)
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
		var r req.PushMessage
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		r.FromUserId = u.Id
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
		err := q.UpdateAllMessageRead(u.Id)
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
		err := q.UpdateAllMessageDeleted(u.Id)
		resp.CheckErr(err)
		resp.Success()
	}
}
