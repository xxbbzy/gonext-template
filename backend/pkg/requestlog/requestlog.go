package requestlog

import (
	"context"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ErrorCodeContextKey stores application error code metadata for request logging.
	ErrorCodeContextKey = "request_log_error_code"
)

var queryValueAllowlist = map[string]func(string) (interface{}, bool){
	"page":      normalizeIntQueryValue,
	"page_size": normalizeIntQueryValue,
	"status":    normalizeStatusQueryValue,
}

var (
	allowedStatusValues = map[string]struct{}{
		"active":   {},
		"inactive": {},
	}
)

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
		if code < 0 {
			return 0, false
		}
		return code, true
	case int32:
		if code < 0 {
			return 0, false
		}
		return int(code), true
	case int64:
		if code < 0 || code > int64(math.MaxInt) {
			return 0, false
		}
		return int(code), true
	case uint:
		if code > uint(math.MaxInt) {
			return 0, false
		}
		return int(code), true
	case uint32:
		if uint64(code) > uint64(math.MaxInt) {
			return 0, false
		}
		return int(code), true
	case uint64:
		if code > uint64(math.MaxInt) {
			return 0, false
		}
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

		if normalizer, allowlisted := queryValueAllowlist[key]; allowlisted {
			summarizeAllowlistedQueryValue(safe, key, vals, normalizer)
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

func summarizeAllowlistedQueryValue(
	safe map[string]interface{},
	key string,
	vals []string,
	normalizer func(string) (interface{}, bool),
) {
	if len(vals) != 1 {
		safe[key+"_len"] = len(vals)
		return
	}

	normalized, ok := normalizer(vals[0])
	if ok {
		safe[key] = normalized
		return
	}

	safe[key+"_present"] = true
	safe[key+"_len"] = len(vals[0])
}

func normalizeIntQueryValue(value string) (interface{}, bool) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}
	return parsed, true
}

func normalizeStatusQueryValue(value string) (interface{}, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if _, ok := allowedStatusValues[value]; !ok {
		return nil, false
	}
	return value, true
}
