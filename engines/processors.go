package engines

import (
	"context"
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

// RequestProcessor is the general type to deal with http requests
type RequestProcessor func(context *HandlerContext) error

// BuildHandlerFunc links processors to process a request
func BuildHandlerFunc(dao storage.Dao, processors ...RequestProcessor) func(http.ResponseWriter, *http.Request) {
	// no handler, default action
	if len(processors) == 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		sharedContext := HandlerContext{
			context:  context.Background(),
			request:  NewRequestDecorator(r),
			response: AbstractResponse{},
			Dao:      dao,
		}

		for _, processor := range processors {
			if err := processor(&sharedContext); err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			} else if sharedContext.response.ShouldSend() {
				sharedContext.response.Write(w)
			}
		}

		// write answer anyway, but do not assume it is ready
		if !sharedContext.response.ShouldSend() {
			sharedContext.ClearResponse()
			sharedContext.SetResponseStatus(http.StatusPartialContent)
			sharedContext.response.Write(w)
		}
	}
}
