package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/golang/glog"
	admissionApi "k8s.io/api/admission/v1"
	admissionRegistration "k8s.io/api/admissionregistration/v1"
	k8sCore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const jsonMIME = "application/json"

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
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

	if log.V(8) {
		bodyString := string(body)
		log.Infof("Request body: %v", bodyString)
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
		log.Errorf("Can't encode the response as JSON: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
	} else if _, err := writer.Write(resp); err != nil {
		log.Errorf("Failed to write response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	} else if log.V(10) {
		log.Infof("Sent response: %v", string(resp))
	}
}

func init() {
	_ = admissionApi.AddToScheme(runtimeScheme)
	_ = admissionRegistration.AddToScheme(runtimeScheme)
	_ = k8sCore.AddToScheme(runtimeScheme)
}
