package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"../storage"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type response struct {
	Success bool
	Data    interface{}
}

type handler struct {
	Schema  string
	Host    string
	storage storage.Service
}

func (h handler) encode(ctx *fasthttp.RequestCtx) (interface{}, int, error) {
	var input struct {
		URL     string `json:"url"`
		Expires string `json:expires`
	}

	if err := json.Unmarshal(ctx.PostBody(), &input); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Unable to decode post body. Error:%v", err)
	}

	uri, err := url.ParseRequestURI(input.URL)

	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid url. Error:%v", err)
	}

	layoutISO := "2006-01-02 15:04:05"

	expires, err := time.Parse(layoutISO, input.Expires)

	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid expiration date. Error:%v", err)
	}

	c, err := h.storage.Save(uri.String(), expires)

	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Could not create shortlink. Error:%v", err)
	}

	u := url.URL{
		Scheme: h.Schema,
		Host:   h.Host,
		Path:   c,
	}

	return u.String(), http.StatusCreated, nil

}

func (h handler) decode(ctx *fasthttp.RequestCtx) (interface{}, int, error) {
	code := ctx.UserValue("shortlink").(string)

	model, err := h.storage.LoadInfo(code)
	if err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("Error in getting url. Error:%v", err)

	}

	return model, http.StatusOK, nil
}

func (h handler) redirect(ctx *fasthttp.RequestCtx) {
	code := ctx.UserValue("shortlink").(string)

	uri, err := h.storage.Load(code)

	if err != nil {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(http.StatusNotFound)
		return
	}
	ctx.Redirect(uri, http.StatusMovedPermanently)
}

func responseHandler(h func(*fasthttp.RequestCtx) (interface{}, int, error)) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		data, status, err := h(ctx)
		if err != nil {
			data = err.Error()
		}

		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(status)
		err = json.NewEncoder(ctx.Response.BodyWriter()).Encode(response{Data: data, Success: err == nil})
		if err != nil {
			log.Printf("Couldnt encode response. Error:%v", err)
		}

	}

}

func New(schema, host string, storage storage.Service) *router.Router {
	router := router.New()

	h := handler{schema, host, storage}

	router.POST("/encode", responseHandler(h.encode))
	router.GET("/{shortlink}", h.redirect)
	router.GET("/{shortlink}/info", responseHandler(h.decode))
	return router
}
