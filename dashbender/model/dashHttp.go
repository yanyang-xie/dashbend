package model

import (
	"fmt"
	"net/http"
	"strings"
)

type ReqestModel struct {
	URL     string //url with parameters
	Method  string
	Headers map[string]string
	Body    string
}

func NewReqestModel(url string, method string) *ReqestModel {
	return &ReqestModel{
		URL:    strings.TrimSpace(url),
		Method: strings.TrimSpace(strings.ToLower(method)),
	}
}

func (r *ReqestModel) String() string {
	return fmt.Sprintf("URL: %v, Method: %v, Headers: %v, Body: %v", r.URL, r.Method, r.Headers, r.Body)
}

func (r *ReqestModel) addHeader(key, value string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string, 0)
	}

	r.Headers[key] = value
}

func (r *ReqestModel) setBody(body string) {
	r.Body = body
}

type ReqestResult struct {
	RequestTime int64

	IsError              bool //whether a http request is error
	ResponseCode         int
	ResponseErrorMessage string
}

func (r *ReqestResult) String() string {
	return fmt.Sprintf("RequestTime: %v, ResponseCode: %v, IsError: %v, ResponseErrorMessage: %v", r.RequestTime, r.ResponseCode, r.IsError, r.ResponseErrorMessage)
}

func NewReqestResult() *ReqestResult {
	return &ReqestResult{
		ResponseCode: http.StatusOK,
		IsError:      false,
	}
}

type RespValidationModel struct {
	Req         *http.Request
	ReponseBody string
}

func NewRespValidationModel(req *http.Request, reponseBody string) *RespValidationModel {
	return &RespValidationModel{
		req,
		reponseBody,
	}
}

func (r *RespValidationModel) String() string {
	return fmt.Sprintf("Request: %v, ResponseBody: %v", r.Req, r.ReponseBody)
}
