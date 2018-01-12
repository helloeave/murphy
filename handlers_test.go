package murphy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pascallouisperez/goutil/errors"
	"github.com/pascallouisperez/goutil/httpstub"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MurphySuite struct{}

var _ = Suite(&MurphySuite{})

type sampleRequest struct {
	Foo string `json:"foo"`
}

type sampleResponse struct{}

func correct(ctx HttpContext, request *sampleRequest, response *sampleResponse) error {
	return nil
}

func correctEmptyRequest(ctx HttpContext, _ *struct{}, response *sampleResponse) error {
	return nil
}

func fails(ctx HttpContext, request *sampleRequest, response *sampleResponse) error {
	return errors.New("a b c")
}

func badRequestWithinHandler(ctx HttpContext, request *sampleRequest, response *sampleResponse) error {
	return BadRequestErrorf("%s %s %s", "a", "b", "c")
}

func (_ *MurphySuite) TestHandlerOf(c *C) {
	_, err := handlerOf(correct)
	c.Assert(err, IsNil)
}

func (_ *MurphySuite) TestJsonHandler_badRequest(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader(`{"foo": not-a-string}`))

	JsonHandler(correct)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusBadRequest)
	c.Assert(w.RecordedBody, Equals, "{\"err\":\"unable to parse request\"}\n")
}

func (_ *MurphySuite) TestJsonHandler_good(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader("{}"))

	JsonHandler(correct)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusOK)
	c.Assert(w.RecordedBody, Equals, "{}\n")
}

func (_ *MurphySuite) TestJsonHandler_expectedEmptyRequest(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader(""))

	JsonHandler(correctEmptyRequest)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusOK)
	c.Assert(w.RecordedBody, Equals, "{}\n")
}

func (_ *MurphySuite) TestJsonHandler_unexpectedEmptyRequest(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader(""))

	JsonHandler(correct)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusBadRequest)
	c.Assert(w.RecordedBody, Equals, "{\"err\":\"unable to parse request; did not expect empty body\"}\n")
}

func (_ *MurphySuite) TestJsonHandler_fails(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader("{}"))

	JsonHandler(fails)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusInternalServerError)
	c.Assert(w.RecordedBody, Equals, "")
}

func (_ *MurphySuite) TestJsonHandler_badRequestWithinHandler(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader("{}"))

	JsonHandler(badRequestWithinHandler)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusBadRequest)
	c.Assert(w.RecordedBody, Equals, "{\"err\":\"a b c\"}\n")
}

func markAsUnauthorized(ctx HttpContext, request *sampleRequest, response *sampleResponse) error {
	ctx.W().WriteHeader(http.StatusUnauthorized)
	return nil
}

func (_ *MurphySuite) TestJsonHandler_handlerMarksRequestAsUnauthorized(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader("{}"))

	JsonHandler(markAsUnauthorized)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusUnauthorized)
	c.Assert(w.RecordedBody, Equals, "")
}

func (_ *MurphySuite) TestBadRequestError_marshalling(c *C) {
	data, err := json.Marshal(BadRequestErrorf("the_err_here"))
	c.Assert(err, IsNil)
	c.Assert(`{"err":"the_err_here"}`, Equals, string(data))
}
