package murphy

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/helloeave/json"
	"github.com/pascallouisperez/goutil/errors"
	"github.com/pascallouisperez/goutil/httpstub"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MurphySuite struct{}

var _ = Suite(&MurphySuite{})

type sampleRequest struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}
type sampleResponse struct{}

type collectionWithNils struct {
	NilSlice     []int                   `json:"nil_slice"`
	PNilSlice    *[]int                  `json:"ptr_nil_slice"`
	NilMap       map[string]interface{}  `json:"nil_map"`
	PNilMap      *map[string]interface{} `json:"ptr_nil_map"`
	AnotherValue string                  `json:"another_value"`
}

func correct(ctx HttpContext, request *sampleRequest, response *sampleResponse) error {
	return nil
}

func correctEmptyRequestStruct(ctx HttpContext, _ *struct{}, response *sampleResponse) error {
	return nil
}

func nilCollectionHandler(ctx HttpContext, request *collectionWithNils, response *collectionWithNils) error {
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

func (_ *MurphySuite) TestJsonHandler_numTimesInvoked(c *C) {
	cases := map[string]int{
		``:   1,
		`{}`: 1,
		`{"foo":"str str str"}`:          1,
		`{"bar": 123}`:                   1,
		`{"foo":"hellostr", "bar": 456}`: 1,
		`{"foo": this-is-bad-json}`:      0,
		`Z1nt3BFxAp0CjMlq`:               0,
	}

	for reqBody, expectedCount := range cases {
		w, r := httpstub.New(c)
		r.Body = ioutil.NopCloser(strings.NewReader(reqBody))

		invoked := 0
		handler := func(_ HttpContext, _ *sampleRequest, _ *sampleResponse) error {
			invoked++
			return nil
		}
		JsonHandler(handler)(w, r)

		c.Assert(invoked, Equals, expectedCount)
	}
}

func (_ *MurphySuite) TestJsonHandler_badRequest(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader(`{"foo": this-is-bad-json}`))

	JsonHandler(correct)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusBadRequest)
	c.Assert(w.RecordedBody, Equals, "{\"err\":\"unable to parse request\"}\n")
}

func (_ *MurphySuite) TestJsonHandler_goodParsedDefaults(c *C) {
	cases := []string{
		``,
		`{}`,
	}

	for _, reqBody := range cases {
		w, r := httpstub.New(c)
		r.Body = ioutil.NopCloser(strings.NewReader(reqBody))

		var exportedReq *sampleRequest
		handler := func(_ HttpContext, req *sampleRequest, _ *sampleResponse) error {
			exportedReq = req
			return nil
		}
		JsonHandler(handler)(w, r)

		c.Assert(exportedReq.Foo, Equals, "")
		c.Assert(exportedReq.Bar, Equals, 0)
	}
}

func (_ *MurphySuite) TestJsonHandler_goodParsedValues(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader(`{"foo": "hihihi", "bar": 2014}`))

	var exportedReq *sampleRequest
	handler := func(_ HttpContext, req *sampleRequest, _ *sampleResponse) error {
		exportedReq = req
		return nil
	}
	JsonHandler(handler)(w, r)

	c.Assert(exportedReq.Foo, Equals, "hihihi")
	c.Assert(exportedReq.Bar, Equals, 2014)
}

func (_ *MurphySuite) TestJsonHandler_goodEmptyRequest(c *C) {
	cases := []string{
		"",
		"{}",
	}

	for _, reqBody := range cases {
		w, r := httpstub.New(c)
		r.Body = ioutil.NopCloser(strings.NewReader(reqBody))

		JsonHandler(correct)(w, r)

		c.Assert(w.RecordedCode, Equals, http.StatusOK)
		c.Assert(w.RecordedBody, Equals, "{}\n")
	}
}

func (_ *MurphySuite) TestJsonHandler_goodEmptyRequestStruct(c *C) {
	cases := []string{
		``,
		`{}`,
		`{"moo": "cow"}`,
		`{"wat": 888}`,
	}

	for _, reqBody := range cases {
		w, r := httpstub.New(c)
		r.Body = ioutil.NopCloser(strings.NewReader(reqBody))

		JsonHandler(correctEmptyRequestStruct)(w, r)

		c.Assert(w.RecordedCode, Equals, http.StatusOK)
		c.Assert(w.RecordedBody, Equals, "{}\n")
	}
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

func (_ *MurphySuite) TestJsonSlices_marshalling(c *C) {
	w, r := httpstub.New(c)
	r.Body = ioutil.NopCloser(strings.NewReader(`{"nil_slice":[],"ptr_nil_slice":null,"nil_map":{},"ptr_nil_map":null,"another_value":""}`))
	JsonHandler(nilCollectionHandler)(w, r)

	c.Assert(w.RecordedCode, Equals, http.StatusOK)
	c.Assert(w.RecordedBody, Equals, "{\"nil_slice\":[],\"ptr_nil_slice\":null,\"nil_map\":{},\"ptr_nil_map\":null,\"another_value\":\"\"}\n")
}
