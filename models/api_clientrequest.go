package models

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type ClientRequest interface {
	RelativeReference() string
	Method() string
	HasBody() bool
	Body() []byte
	HasContentType() bool
	ContentType() string
	NeedAuthorization() bool
}

// client Request
type clientRequest struct {
	ClientRequest     `json:"-"`
	method            string
	relativeReference string
	body              []byte
	contentType       string
	needAuth          bool
}

func newGetRequest(rr string) ClientRequest {
	return &clientRequest{
		method:            fiber.MethodGet,
		relativeReference: rr,
	}
}

func newPostRequest(rr string) *clientRequest {
	return &clientRequest{
		method:            fiber.MethodPost,
		relativeReference: rr,
	}
}

func (r *clientRequest) RelativeReference() string {
	return r.relativeReference
}

func (r *clientRequest) Method() string {
	return r.method
}

func (r *clientRequest) HasBody() bool {
	return r.body != nil
}

func (r *clientRequest) Body() []byte {
	return r.body
}

func (r *clientRequest) HasContentType() bool {
	return r.contentType != ""
}

func (r *clientRequest) ContentType() string {
	return r.contentType
}

func (r *clientRequest) NeedAuthorization() bool {
	return r.needAuth
}

func (r *clientRequest) SetContentType(contentType string) *clientRequest {
	r.contentType = contentType
	return r
}

func (r *clientRequest) SetContentTypeJSONUTF8() *clientRequest {
	return r.SetContentType(fiber.MIMEApplicationJSONCharsetUTF8)
}

func (r *clientRequest) WithAuth() *clientRequest {
	r.needAuth = true
	return r
}

func (r *clientRequest) SetBodyBytes(body []byte) *clientRequest {
	r.body = body
	return r
}

func (r *clientRequest) SetBodyJSON(body any) *clientRequest {
	var err error
	r.body, err = json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return r
}
