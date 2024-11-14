package statsapi

import "github.com/gorilla/mux"

var (
	handler *mux.Router
)

func init() {
	handler = mux.NewRouter()
}
