package transformer

import (
	"fmt"
	"net/http"

	kithttptest "github.com/rudderlabs/rudder-go-kit/testhelper/httptest"

	"github.com/google/uuid"
)

const (
	getByVersionIdEndPoint = "/transformation/getByVersionId"
	versionIDKey           = "versionId"
)

func NewTestBackendConfigServer(transformations map[string]string) *kithttptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(getByVersionIdEndPoint, getByVersionIdHandler(transformations))
	return kithttptest.NewServer(mux)
}

func getByVersionIdHandler(transformations map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		transformationVersionId := r.URL.Query().Get(versionIDKey)
		transformationCode, ok := transformations[transformationVersionId]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(fmt.Sprintf(`{
		"id": %q,
		"createdAt": "2023-03-27T11:40:00.894Z",
		"updatedAt": "2023-03-27T11:40:00.894Z",
		"versionId": %q,
		"name": "Add Transformed field",
		"description": "",
		"code": %q,
		"secretsVersion": null,
		"codeVersion": "1",
		"language": "javascript",
		"imports": [],
		"secrets": {}
	}`, uuid.NewString(), transformationVersionId, transformationCode)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = fmt.Errorf("error writing response: %v", err)
			return
		}
	}
}
