package api

import (
	"io/ioutil"
	"log"
	"net/http"

	js "github.com/tolopsy/url-shortener/serializer/json"
	ms "github.com/tolopsy/url-shortener/serializer/msgpack"
	"github.com/tolopsy/url-shortener/shortener"

	"github.com/go-chi/chi"
	errs "github.com/pkg/errors"
)

const contentTypeKey = "Content-Type"

type RedirectHandler interface {
	Get(http.ResponseWriter, *http.Request)
	Post(http.ResponseWriter, *http.Request)
}

type handler struct {
	redirectService shortener.RedirectService
}

func NewHandler(service shortener.RedirectService) RedirectHandler {
	return &handler{redirectService: service}
}

func setupResponse(w http.ResponseWriter, contentType string, body []byte, statusCode int) {
	w.Header().Set(contentTypeKey, contentType)
	w.WriteHeader(statusCode)
	_, err := w.Write(body)
	if err != nil {
		log.Println(err)
	}
}

func (h *handler) serializer(contentType string) shortener.RedirectSerializer {
	var serializerObj shortener.RedirectSerializer

	switch contentType {
	case "application/x-msgpack":
		serializerObj = &ms.RedirectSerializer{}
	case "application/json":
		serializerObj = &js.RedirectSerializer{}
	default:
		serializerObj = &js.RedirectSerializer{}
	}

	return serializerObj
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	redirect, err := h.redirectService.Find(code)
	if err != nil {
		if errs.Cause(err) == shortener.ErrRedirectNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, redirect.URL, http.StatusMovedPermanently)
}

func (h *handler) Post(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get(contentTypeKey)
	redirect, err := h.serializer(contentType).Decode(requestBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = h.redirectService.Store(redirect)
	if err != nil {
		if errs.Cause(err) == shortener.ErrRedirectInvalid {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseBody, err := h.serializer(contentType).Encode(redirect)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	setupResponse(w, contentType, responseBody, http.StatusCreated)
}
