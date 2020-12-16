package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/korylprince/httputil/jsonapi"
)

const API = "1.0"
const apiPath = "/api/" + API

type CheckRequest struct {
	Email string `json:"email"`
	ID    string `json:"id"`
}

type StatusResponse struct {
	Blocked bool  `json:"blocked"`
	Error   error `json:"error"`
}

func (c *Config) checkFunc(r *http.Request, filter func(string, string) error) (int, interface{}) {
	req := new(CheckRequest)
	if err := jsonapi.ParseJSONBody(r, req); err != nil {
		return http.StatusBadRequest, fmt.Errorf("Invalid body: %w", err)
	}
	if req.ID == "" {
		return http.StatusBadRequest, errors.New("ID is required")
	}
	jsonapi.LogActionID(r, fmt.Sprintf("%s:%s", req.Email, req.ID))

	if _, ok := c.allowedEmails[req.Email]; !ok {
		return http.StatusBadRequest, errors.New("email not allowed")
	}

	err := filter(req.Email, req.ID)
	if _, ok := err.(BlockedError); ok || err == nil {
		return http.StatusOK, &StatusResponse{Blocked: err != nil, Error: err}
	}
	return http.StatusInternalServerError, err
}

func (c *Config) CheckHost(r *http.Request) (int, interface{}) {
	return c.checkFunc(r, c.FilterHost)
}

func (c *Config) CheckVideo(r *http.Request) (int, interface{}) {
	return c.checkFunc(r, c.FilterVideo)
}

func (c *Config) CheckChannel(r *http.Request) (int, interface{}) {
	return c.checkFunc(r, c.FilterChannel)
}

func (c *Config) Router() http.Handler {
	api := jsonapi.New(os.Stdout, nil, nil, nil)
	api.Handle("POST", "/hosts", c.CheckHost, false)
	api.Handle("POST", "/videos", c.CheckVideo, false)
	api.Handle("POST", "/channels", c.CheckChannel, false)

	r := mux.NewRouter()
	r.PathPrefix(apiPath).Handler(http.StripPrefix(apiPath, api))
	return r
}
