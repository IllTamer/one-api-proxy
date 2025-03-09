package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"io"
	"net/http"
)

func SetUpLogger(server *gin.Engine) {
	server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var requestID string
		if param.Keys != nil {
			requestID = param.Keys[helper.RequestIdKey].(string)
		}
		return fmt.Sprintf("[GIN] %s | %s | %3d | %13v | %15s | %7s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			requestID,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))
}

func SetUpRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		requestBody := ""
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				logger.Errorf(ctx, "Request body read error: %v", err)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = string(bodyBytes)
		}
		model.AddRequestRecord(ctx, c.Request, c.GetString(helper.RequestIdKey), requestBody)
		c.Next()
	}
}

// ResponseWriter 自定义ResponseWriter，用于记录响应内容
type customResponseWriter struct {
	gin.ResponseWriter               // 嵌入Gin的ResponseWriter
	body               *bytes.Buffer // 用于存储响应体的缓冲区
}

// Write 重写Write方法，在写入响应的同时记录数据
func (w customResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)                  // 将数据写入缓冲区
	return w.ResponseWriter.Write(b) // 继续写入原始ResponseWriter
}

func SetUpResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &customResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer
		c.Next()
		model.AddResponseRecord(c.Request.Context(), c.GetString(helper.RequestIdKey), c.GetInt(ctxkey.Id), c.GetString(ctxkey.RequestModel), c.Writer.Status(), writer.body.String())
	}
}
