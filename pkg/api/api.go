package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"wunderbase/pkg/graphiql"

	"github.com/buger/jsonparser"
	"github.com/wundergraph/graphql-go-tools/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/pkg/introspection"
	"go.uber.org/ratelimit"
)

type Handler struct {
	enableSleepMode   bool
	enablePlayground  bool
	queryEngineURL    string
	queryEngineSdlURL string
	sleepAfterSeconds int
	init              sync.Once
	sleepCh           chan struct{}
	client            *http.Client
	readLimit         ratelimit.Limiter
	writeLimit        ratelimit.Limiter
	cancel            func()
}

func NewHandler(enableSleepMode bool, production bool, queryEngineURL string, queryEngineSdlURL string, sleepAfterSeconds, readLimitSeconds, writeLimitSeconds int, cancel func()) *Handler {
	return &Handler{
		enableSleepMode:   enableSleepMode,
		enablePlayground:  !production,
		queryEngineURL:    queryEngineURL,
		queryEngineSdlURL: queryEngineSdlURL,
		sleepCh:           make(chan struct{}),
		sleepAfterSeconds: sleepAfterSeconds,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		readLimit:  ratelimit.New(readLimitSeconds),
		writeLimit: ratelimit.New(writeLimitSeconds),
		cancel:     cancel,
	}
}

type IntrospectionResponse struct {
	Data introspection.Data `json:"data"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.init.Do(func() {
		if h.enableSleepMode {
			go h.runSleepMode()
		}
		for {
			resp, err := http.Get(h.queryEngineURL)
			if err != nil || resp.StatusCode != http.StatusOK {
				time.Sleep(3 * time.Millisecond)
				continue
			}
			break
		}
	})

	if h.enableSleepMode {
		defer func() {
			h.sleepCh <- struct{}{}
		}()
	}

	if h.enablePlayground && r.Header.Get("Content-Type") != "application/json" {
		w.Header().Add("Content-Type", "text/html")
		html := graphiql.GetGraphiqlPlaygroundHTML(r.RequestURI)
		_, _ = w.Write([]byte(html))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// check if body is introspection query
	if bytes.Contains(body, []byte("IntrospectionQuery")) {
		// if so, return the schema
		w.Header().Add("Content-Type", "application/json")
		gen := introspection.NewGenerator()
		// get the schema from the query engine on /sdl endpoint
		resp, err := http.Get(h.queryEngineSdlURL)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		schemaSDL, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		// generate the introspection result from the schema
		doc, report := astparser.ParseGraphqlDocumentBytes(schemaSDL)
		err = asttransform.MergeDefinitionWithBaseSchema(&doc)
		if err != nil {
			log.Fatalln(err)
		}
		var response IntrospectionResponse
		gen.Generate(&doc, &report, &response.Data)
		// marshal the result
		b, err := json.Marshal(response)
		if err != nil {
			log.Fatalln(err)
		}
		_, _ = w.Write(b)
		return
	}
	h.proxyRequestToEngine(body, w, r)
}

func (h *Handler) proxyRequestToEngine(body []byte, w http.ResponseWriter, r *http.Request) {
	variables, _, _, _ := jsonparser.Get(body, "variables")
	if variables == nil {
		// if no variables are set, set an empty object
		body, _ = jsonparser.Set(body, []byte("{}"), "variables")
	}
	operationName, _, _, _ := jsonparser.Get(body, "operationName")
	if operationName == nil {
		// if no operation name is set, set an empty string
		body, _ = jsonparser.Set(body, []byte("null"), "operationName")
	}
	for i := 0; i < 3; i++ {
		if h.sendRequest(body, w, r) {
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *Handler) sendRequest(body []byte, w http.ResponseWriter, r *http.Request) bool {

	if bytes.Contains(body, []byte("mutation")) {
		h.writeLimit.Take()
	}
	h.readLimit.Take()

	newRequest, err := http.NewRequestWithContext(r.Context(), r.Method, h.queryEngineURL, ioutil.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		log.Println(err)
		return false
	}
	// set the content type to application/json
	newRequest.Header.Set("content-type", "application/json")
	resp, err := h.client.Do(newRequest)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	if bytes.HasPrefix(data, []byte("{\"e")) && bytes.Contains(data, []byte("Timed out")) {
		return false
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (h *Handler) runSleepMode() {
	timer := time.NewTimer(time.Duration(h.sleepAfterSeconds) * time.Second)
	defer func() {
		fmt.Println("No requests for", h.sleepAfterSeconds, "seconds, cancelling context")
		h.cancel()
		return
	}()
	for {
		select {
		case <-h.sleepCh:
			done := timer.Reset(time.Duration(h.sleepAfterSeconds) * time.Second)
			if !done {
				return
			}
		case <-timer.C:
			return
		}
	}
}
