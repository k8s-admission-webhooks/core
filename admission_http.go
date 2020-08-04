package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/golang/glog"
	admissionApi "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const jsonMIME = "application/json"

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	//defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

// ReadAdmissionReview read an AdmissionReview from a request
func ReadAdmissionReview(request *http.Request) (*admissionApi.AdmissionReview, error) {
	var body []byte
	if request.Body != nil {
		if data, err := ioutil.ReadAll(request.Body); err != nil {
			body = data
		}
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("Empty body")
	}

	contentType := request.Header.Get("Content-Type")
	if contentType != jsonMIME {
		return nil, fmt.Errorf("Invalid content-type. Received: %v, Expected: %v", jsonMIME, contentType)
	}

	ar := admissionApi.AdmissionReview{}
	_, _, err := deserializer.Decode(body, nil, &ar)

	return &ar, err
}

// CreateErrorResponse create an admission error response
func CreateErrorResponse(errorDesc string) *admissionApi.AdmissionResponse {
	return &admissionApi.AdmissionResponse{
		Result: &metav1.Status{
			Status:  "Failure",
			Message: errorDesc,
		},
	}
}

// WriteAdmissionResponse write an AdmissionResponse as a HTTP response
func WriteAdmissionResponse(
	writer http.ResponseWriter,
	ar *admissionApi.AdmissionReview,
	response *admissionApi.AdmissionResponse) {
	responseAR := admissionApi.AdmissionReview{}
	if response != nil {
		responseAR.Response = response
		if ar.Request != nil {
			responseAR.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(responseAR)
	if err != nil {
		errDesc := fmt.Sprintf("Can't encode the response as JSON: %v", err)
		log.Error(errDesc)
		http.Error(writer, errDesc, http.StatusInternalServerError)
	} else if _, err := writer.Write(resp); err != nil {
		errDesc := fmt.Sprintf("Failed to write response: %v", err)
		log.Error(errDesc)
		http.Error(writer, errDesc, http.StatusInternalServerError)
	} else {
		log.Infof("Response successfully sent to the client: %v", resp)
	}
}
