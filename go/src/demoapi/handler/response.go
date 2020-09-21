package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	he "demoapi/httperror"
)

func jsonResponse(w http.ResponseWriter, obj interface{}, err error) {
	writeJSONError := func(jsonErr error) {
		statusCode := he.StatusCodeByError(jsonErr)
		w.WriteHeader(statusCode)

		errStr := fmt.Sprintf("%s", jsonErr)
		jsonErrObj := map[string]string{"error": errStr}
		jsonObj, err := json.Marshal(jsonErrObj)
		if err != nil {
			logrus.Warningf("failed to convert %v to json", jsonErrObj)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(errStr))
			return
		}

		w.Write(jsonObj)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		writeJSONError(err)
		return
	}

	if obj == nil {
		obj = map[string]string{"response": "okay!"}
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		writeJSONError(err)
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		writeJSONError(err)
		return
	}
}
