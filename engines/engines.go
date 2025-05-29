package engines

import (
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

type ProcessingEngine struct {
	dao storage.Dao
	mux *http.ServeMux
}

func NewProcessingEngine(dao storage.Dao) ProcessingEngine {
	return ProcessingEngine{
		dao: dao,
		mux: http.NewServeMux(),
	}
}

func (e *ProcessingEngine) AddProcessors(method string, urlPattern string, processors ...RequestProcessor) {
	var allProcessors []RequestProcessor
	allProcessors = append(allProcessors, ValidateQueryProcessor(method))
	allProcessors = append(allProcessors, processors...)
	e.mux.HandleFunc(urlPattern, BuildHandlerFunc(e.dao, allProcessors...))
}

func (e *ProcessingEngine) Launch(address string) {
	http.ListenAndServe(address, e.mux)
}
