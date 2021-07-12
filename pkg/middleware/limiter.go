package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	ginLimiter "github.com/julianshen/gin-limiter"
)

// NewRateLimiter ...
func NewRateLimiter(interval time.Duration, cap int64) gin.HandlerFunc {
	return ginLimiter.NewRateLimiter(interval, cap, func(ctx *gin.Context) (string, error) {
		return "", nil
	}).Middleware()
}
