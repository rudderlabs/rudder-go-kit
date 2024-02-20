package transformer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	getByVersionIdEndPoint = "/transformation/getByVersionId"
	versionIDKey           = "versionId"
)

type mockHttpServer struct {
	Transformations map[string]string
}

func (m *mockHttpServer) handleGetByVersionId(w http.ResponseWriter, r *http.Request) {
	transformationVersionId := r.URL.Query().Get(versionIDKey)
	transformationCode, ok := m.Transformations[transformationVersionId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	transformationCode = sanitiseTransformationCode(transformationCode)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(`{
		"id": "%s",
		"createdAt": "2023-03-27T11:40:00.894Z",
		"updatedAt": "2023-03-27T11:40:00.894Z",
		"versionId": "%s",
		"name": "Add Transformed field",
		"description": "",
		"code": "%s",
		"secretsVersion": null,
		"codeVersion": "1",
		"language": "javascript",
		"imports": [],
		"secrets": {}
	}`, uuid.NewString(), transformationVersionId, transformationCode)))
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func sanitiseTransformationCode(transformationCode string) string {
	sanitisedTransformationCode := strings.ReplaceAll(transformationCode, "\t", " ")
	sanitisedTransformationCode = strings.ReplaceAll(sanitisedTransformationCode, "\n", "\\n")
	return sanitisedTransformationCode
}
