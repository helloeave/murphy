package murphy

import (
	"encoding/json"
	"main/util/errors"
	"net/http"
	"reflect"

	"github.com/golang/glog"
	"github.com/pascallouisperez/reflext"
	go_uuid "github.com/satori/go.uuid"
)

// JsonHandler converts a function capable of being a handler for JSON APIs into
// a net/http compatible handler. To serve JSON API requests, a function may
// take either a struct as arguments (to be converted from JSON passed in the
// body), an http.Request, both or none, and must return a struct pointer, and
// an error.
func JsonHandler(fn interface{}) func(w http.ResponseWriter, r *http.Request) {
	m, err := handlerOf(fn)
	if err != nil {
		panic(err)
	}
	return m.makeHandler()
}

var handlerPattern = reflext.MustCompile(
	"func(%T, *{struct}, *{struct}) error",
	reflect.TypeOf((*HttpContext)(nil)).Elem())

func handlerOf(handler interface{}) (*handlerMaker, error) {
	captures, ok := handlerPattern.FindAll(handler)
	if !ok {
		return nil, errors.New("%#v expected %s, got %T", handlerPattern, handlerPattern, handler)
	}
	return &handlerMaker{
		handlerValue: reflect.ValueOf(handler),
		requestType:  captures[0],
		responseType: captures[1],
	}, nil
}

type handlerMaker struct {
	handlerValue              reflect.Value
	requestType, responseType reflect.Type
}

func (m *handlerMaker) makeHandler() func(http.ResponseWriter, *http.Request) {
	return func(actualW http.ResponseWriter, r *http.Request) {
		w := &responseWriter{actualW, false}

		requestPtr := reflect.New(m.requestType)
		if err := json.NewDecoder(r.Body).Decode(requestPtr.Interface()); err != nil {
			handleBadRequestErr(w, r, BadRequestErrorf("%s", err))
			return
		}

		responsePtr := reflect.New(m.responseType)

		arguments := make([]reflect.Value, 3, 3)
		arguments[0] = reflect.ValueOf(NewHttpContext(w, r))
		arguments[1] = requestPtr
		arguments[2] = responsePtr

		returns := m.handlerValue.Call(arguments)

		if !returns[0].IsNil() {
			if err, ok := returns[0].Interface().(*BadRequestError); ok {
				handleBadRequestErr(w, r, err)
				return
			} else {
				handleErr(w, r, returns[0].Interface().(error))
				return
			}
		}

		if w.skipResponse {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(responsePtr.Interface()); err != nil {
			handleErr(w, r, err)
			return
		}
	}
}

type BadRequestError struct {
	error
}

// Assert that BadRequestError implements the error interface.
var _ error = &BadRequestError{}

func BadRequestErrorf(format string, args ...interface{}) *BadRequestError {
	return &BadRequestError{
		error: errors.New(format, args...),
	}
}

func canShowErr(r *http.Request) bool {
	return r.Host == "localhost"
}

func handleBadRequestErr(w http.ResponseWriter, r *http.Request, theErr *BadRequestError) {
	errId := go_uuid.NewV4().String()
	w.Header().Set("X-Errid", errId)
	w.WriteHeader(http.StatusBadRequest)
	if canShowErr(r) {
		w.Write([]byte(theErr.Error()))
	}
	glog.Errorf("errId=%s, err=%s", errId, theErr)
}

func handleErr(w http.ResponseWriter, r *http.Request, theErr error) {
	errId := go_uuid.NewV4().String()
	w.Header().Set("X-Errid", errId)
	w.WriteHeader(http.StatusInternalServerError)
	if canShowErr(r) {
		w.Write([]byte(theErr.Error()))
	}
	glog.Errorf("errId=%s, err=%s", errId, theErr)
}
