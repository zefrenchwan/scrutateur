package engines

import (
	"net/http"
	"strings"
)

// ValidateQueryProcessor returns a processor that validates the query method type
func ValidateQueryProcessor(requestMethod string) RequestProcessor {
	return func(context *HandlerContext) error {
		if strings.ToLower(requestMethod) != strings.ToLower(context.GetRequestMethod()) {
			context.Build(http.StatusMethodNotAllowed, "", nil)
		}

		return nil
	}
}
