package log

import (
  // "time"
	"go.uber.org/zap"
)

var (
  logger *zap.SugaredLogger
)

func Infof(template string, args ...interface{}) {
  logger.Infof(template, args...)
}

func Debugf(template string, args ...interface{}) {
  logger.Debugf(template, args...)
}
 
func init() {
	ctx, _ := zap.NewProduction()
	defer ctx.Sync() // flushes buffer, if any
	logger = ctx.Sugar()
	// sugar.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", url,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// sugar.Infof("Failed to fetch URL: %s", url)
}
