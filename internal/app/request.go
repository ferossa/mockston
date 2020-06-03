package app

// ProcessRequest data to be processed
type ProcessRequest struct {
	ResponseChannel chan *ProcessResponse  // channel for response
	Endpoint        string                 // endpoint name
	Content         []byte                 // request data (e.g. http request body)
	Context         map[string]interface{} // additional data (e.g. http headers)
}

// ProcessResponse data to be sent in response
type ProcessResponse struct {
	Error   error                  // processing error
	Content []byte                 // content to send
	Context map[string]interface{} // additional response data (e.g. http headers)
}
