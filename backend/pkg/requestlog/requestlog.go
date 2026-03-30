package requestlog

import (
	"context"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ErrorCodeContextKey stores application error code metadata for request logging.
	ErrorCodeContextKey = "request_log_error_code"
)

var queryValueAllowlist = map[string]struct{}{
	"page":      {},
	"page_size": {},
	"status":    {},
	"sort":      {},
	"order":     {},
}

// QuerySummary is a safe query representation for request logs.
type QuerySummary struct {
	Keys []string
	Safe map[string]interface{}
}

// SetErrorCode stores request-log error metadata in the current gin context.
func SetErrorCode(c *gin.Context, code int) {
	if c == nil {
		return
	}
	c.Set(ErrorCodeContextKey, code)
}

// SetErrorCodeFromContext stores request-log error metadata when ctx carries gin.Context.
func SetErrorCodeFromContext(ctx context.Context, code int) {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return
	}
	SetErrorCode(ginCtx, code)
}

// GetErrorCode returns request-log error metadata from gin context.
func GetErrorCode(c *gin.Context) (int, bool) {
	if c == nil {
		return 0, false
	}

	value, exists := c.Get(ErrorCodeContextKey)
	if !exists {
		return 0, false
	}

	switch code := value.(type) {
	case int:
		return code, true
	case int32:
		return int(code), true
	case int64:
		return int(code), true
	case uint:
		return int(code), true
	case uint32:
		return int(code), true
	case uint64:
		return int(code), true
	default:
		return 0, false
	}
}

// RouteAndPath returns route template and concrete path.
func RouteAndPath(c *gin.Context) (route string, path string) {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return "", ""
	}

	path = c.Request.URL.Path
	route = strings.TrimSpace(c.FullPath())
	if route == "" {
		route = path
	}

	return route, path
}

// SummarizeContextQuery builds a safe query summary from gin context.
func SummarizeContextQuery(c *gin.Context) QuerySummary {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return QuerySummary{
			Keys: []string{},
			Safe: map[string]interface{}{},
		}
	}
	return SummarizeQuery(c.Request.URL.Query())
}

// SummarizeQuery builds a safe query summary that never logs free-text values.
func SummarizeQuery(values url.Values) QuerySummary {
	keys := make([]string, 0, len(values))
	safe := make(map[string]interface{}, len(values))

	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		vals := values[key]
		if len(vals) == 0 {
			safe[key+"_present"] = true
			safe[key+"_len"] = 0
			continue
		}

		if _, allowlisted := queryValueAllowlist[key]; allowlisted {
			if len(vals) == 1 {
				safe[key] = vals[0]
			} else {
				safe[key] = vals
			}
			continue
		}

		totalLen := 0
		for _, v := range vals {
			totalLen += len(v)
		}
		safe[key+"_present"] = true
		safe[key+"_len"] = totalLen
	}

	return QuerySummary{
		Keys: keys,
		Safe: safe,
	}
}
