package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	logCache = make([]OperationRecord, 0)
	logLock  sync.Mutex
)

type OperationApi struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Desc   string `json:"desc"`
}

type OperationRecord struct {
	CreatedAt  carbon.ToDateTimeString `json:"createdAt"`
	ApiDesc    string                  `json:"apiDesc"`
	Path       string                  `json:"path"`
	Method     string                  `json:"method"`
	Header     string                  `json:"header"`
	Body       string                  `json:"body"`
	Params     string                  `json:"params"`
	Resp       string                  `json:"resp"`
	Status     int                     `json:"status"`
	Username   string                  `json:"username"`
	RoleName   string                  `json:"roleName"`
	Ip         string                  `json:"ip"`
	IpLocation string                  `json:"ipLocation"`
	Latency    time.Duration           `json:"latency"`
	UserAgent  string                  `json:"userAgent"`
}

func OperationLog(options ...func(*OperationLogOptions)) gin.HandlerFunc {
	ops := getOperationLogOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		startTime := carbon.Now()
		// read body
		var body []byte
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			ops.logger.Error(c, "read body err: %v", err)
		} else {
			// write back to gin request body
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		// find request params
		reqParams := c.Request.URL.Query()
		defer func() {
			if ops.skipGetOrOptionsMethod {
				// skip GET/OPTIONS
				if c.Request.Method == http.MethodGet ||
					c.Request.Method == http.MethodOptions {
					return
				}
			}
			// custom skip path
			for _, s := range ops.skipPaths {
				if strings.Contains(c.Request.URL.Path, s) {
					return
				}
			}

			endTime := carbon.ToDateTimeString{
				Carbon: carbon.Now(),
			}

			if len(body) == 0 {
				body = []byte("{}")
			}
			contentType := c.Request.Header.Get("Content-Type")
			// multipart/form-data
			if strings.Contains(contentType, "multipart/form-data") {
				contentTypeArr := strings.Split(contentType, "; ")
				if len(contentTypeArr) == 2 {
					// read boundary
					boundary := strings.TrimPrefix(contentTypeArr[1], "boundary=")
					b := strings.NewReader(string(body))
					r := multipart.NewReader(b, boundary)
					f, _ := r.ReadForm(ops.singleFileMaxSize << 20)
					defer f.RemoveAll()
					params := make(map[string]string, 0)
					for key, val := range f.Value {
						// get first value
						if len(val) > 0 {
							params[key] = val[0]
						}
					}
					params["content-type"] = "multipart/form-data"
					params["file"] = "binary data ignored"
					// save data by json format
					body = []byte(utils.Struct2Json(params))
				}
			}
			// read header
			header := make(map[string]string, 0)
			for k, v := range c.Request.Header {
				header[k] = strings.Join(v, " | ")
			}
			record := OperationRecord{
				Ip:        c.ClientIP(),
				Method:    c.Request.Method,
				Path:      strings.TrimPrefix(c.Request.URL.Path, "/"+ops.urlPrefix),
				Header:    utils.Struct2Json(header),
				Body:      string(body),
				Params:    utils.Struct2Json(reqParams),
				Latency:   endTime.Time.Sub(startTime.Time),
				UserAgent: c.Request.UserAgent(),
			}
			record.CreatedAt = endTime

			username, roleName := ops.getUserInfo(c)

			record.Username = constant.MiddlewareOperationLogNotLogin
			record.RoleName = constant.MiddlewareOperationLogNotLogin
			if username != "" {
				record.Username = username
				record.RoleName = roleName
			}

			record.ApiDesc = getApiDesc(c, record.Method, record.Path, *ops)
			// get ip location
			record.IpLocation = utils.GetIpRealLocation(record.Ip, ops.realIpKey)

			record.Status = c.Writer.Status()
			rp, exists := c.Get(ops.ctxKey)
			var response string
			if exists {
				response = utils.Struct2Json(rp)
				if item, ok := rp.(resp.Resp); ok {
					if item.Code == resp.Unauthorized {
						return
					}
					record.Status = item.Code
				}
			} else {
				response = "no resp"
			}
			record.Resp = response

			// delay to update to db
			logLock.Lock()
			logCache = append(logCache, record)
			if len(logCache) >= ops.maxCountBeforeSave {
				list := make([]OperationRecord, len(logCache))
				copy(list, logCache)
				go ops.save(c, list)
				logCache = make([]OperationRecord, 0)
			}
			logLock.Unlock()
		}()
		c.Next()
	}
}

func getApiDesc(c *gin.Context, method, path string, ops OperationLogOptions) string {
	desc := "no desc"
	if ops.redis != nil {
		oldCache, _ := ops.redis.HGet(c, ops.apiCacheKey, fmt.Sprintf("%s_%s", method, path)).Result()
		if oldCache != "" {
			return oldCache
		}
	}
	apis := ops.findApi(c)
	for _, api := range apis {
		if api.Method == method && api.Path == path {
			desc = api.Desc
			break
		}
	}

	if ops.redis != nil {
		pipe := ops.redis.Pipeline()
		for _, api := range apis {
			pipe.HSet(c, ops.apiCacheKey, fmt.Sprintf("%s_%s", api.Method, api.Path), api.Desc)
		}
		pipe.Exec(c)
	}
	return desc
}
