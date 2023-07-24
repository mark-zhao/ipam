package v1

import (
	"ipam/component"
)

var resp component.GokuApiResponse

type Response struct {
	msg  string      `json:"msg"`
	data interface{} `json:"data"`
}

type UriInterface interface {
	GetModel() string
	GetUri() string
}

type Uri struct {
	model string
	uri   string
}

func NewUri(model, uri string) *Uri {
	if len(uri) != 0 && uri[0:1] != "/" {
		uri = ""
	}
	return &Uri{
		model: model,
		uri:   uri,
	}
}

func (u *Uri) GetModel() string {
	return u.model
}

func (u *Uri) GetUri() string {
	return u.uri
}

var APIs = make(map[string]map[UriInterface]interface{})
