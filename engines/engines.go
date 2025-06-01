package engines

import (
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

// ProcessingEngine links url patterns to processors
type ProcessingEngine struct {
	dao storage.Dao
	mux *http.ServeMux
}

// NewProcessingEngine builds a new engine.
// Dao parameter is necessary to put it on each context
func NewProcessingEngine(dao storage.Dao) ProcessingEngine {
	return ProcessingEngine{
		dao: dao,
		mux: http.NewServeMux(),
	}
}

// AddProcessors links a (method + urlpattern) to a set of processors
func (e *ProcessingEngine) AddProcessors(method string, urlPattern string, processors ...RequestProcessor) {
	var allProcessors []RequestProcessor
	allProcessors = append(allProcessors, ValidateQueryProcessor(method))
	allProcessors = append(allProcessors, processors...)
	e.mux.HandleFunc(urlPattern, BuildHandlerFunc(e.dao, allProcessors...))
}

// Launch starts the engine
func (e *ProcessingEngine) Launch(address string) {
	http.ListenAndServe(address, e.mux)
}
