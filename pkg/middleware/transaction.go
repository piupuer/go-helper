package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/resp"
	"gorm.io/gorm"
	"net/http"
)

func Transaction(options ...func(*TransactionOptions)) gin.HandlerFunc {
	ops := getTransactionOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.dbNoTx == nil {
		panic("dbNoTx is empty")
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		noTransaction := false
		if method == "OPTIONS" || method == "GET" {
			// OPTIONS/GET skip transaction
			noTransaction = true
		}
		defer func() {
			// get db transaction
			tx := getTx(c, *ops)
			if err := recover(); err != nil {
				if rp, ok := err.(resp.Resp); ok {
					if !noTransaction {
						if rp.Code == resp.Ok {
							// commit transaction
							tx.Commit()
						} else {
							// rollback transaction
							tx.Rollback()
						}
					}
					rp.RequestId = c.GetString(ops.requestIdCtxKey)
					c.JSON(http.StatusOK, rp)
					c.Abort()
					return
				}
				if !noTransaction {
					tx.Rollback()
				}
				// throw up exception
				panic(err)
			} else {
				if !noTransaction {
					tx.Commit()
				}
			}
			c.Abort()
		}()
		if !noTransaction {
			tx := ops.dbNoTx.Begin()
			c.Set(ops.txCtxKey, tx)
		}
		c.Next()
	}
}

func getTx(c *gin.Context, ops TransactionOptions) *gorm.DB {
	tx := ops.dbNoTx
	txKey, exists := c.Get(ops.txCtxKey)
	if exists {
		if item, ok := txKey.(*gorm.DB); ok {
			tx = item
		}
	}
	return tx
}
