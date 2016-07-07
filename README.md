# Murphy

_An opiniated view on JSON APIs, and a short implementation in Golang._

__tl;dr__

    type Request struct { Name string }
    
    type Response struct { Greeting string }
    
    func Hello(ctx *murphy.HttpContext, req *Request, resp *Response) error {
    	if len(req.Name) == 0 {
    		return murphy.BadRequestErrorf("name missing")
    	}
    	resp.Greeting = "Hello " + req.Name
    	return nil
    }

and installed as such

    http.Handle("/hello", murphy.JsonHandler(Hello))
    log.Fatal(http.ListenAndServe(":8080", nil))

## Remote APIs

At the simplest level, remote APIs such as "JSON APIs" can be looked at through the lens of normal API design. The domain the API deals with has its concepts, lexicon, short and long lived entities, and operations to create, mutate, discover / list, or otherwise destroy these.

Where things depart from simple API design is the remotness. This introduces two complexities: transport, and error handling.

From the client's perspective (i.e. the caller), invoking a remote method results in one of two states: _unknown_ where you do not know whether things succeeded or failed; and _known_ where you either know things succeeded, or know they failed. The former typically occurs with timeout, network separation, or any other I/O level problem; the latter is a property of the API being invoked. Dealing with _unknown_ state is fun, and such a large topic by itself that we're goind to elude it here. Dealing with _known_ state is _almost_ similar as API design... you got it, error handling.

