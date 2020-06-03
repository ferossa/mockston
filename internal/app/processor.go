package app

import (
	"errors"
	"github.com/ferossa/mockston/internal/cfg"
)

// IProcessor interface for message processor
type IProcessor interface {
	SetEndpoints(endpoints []cfg.Endpoint)
	Run()
	Process(request *ProcessRequest) (*ProcessResponse, error)
}

// Processor processor struct
type Processor struct {
	threads   int64
	rqChan    chan *ProcessRequest
	endpoints map[string][]cfg.Test
}

// NewProcessor create new Processor
func NewProcessor(threads int64) *Processor {
	return &Processor{
		threads: threads,
		rqChan:  make(chan *ProcessRequest),
	}
}

// SetEndpoints set endpoints which this processor understands
func (p *Processor) SetEndpoints(endpoints []cfg.Endpoint) {
	p.endpoints = map[string][]cfg.Test{}
	for _, endpoint := range endpoints {
		p.endpoints[endpoint.Name] = endpoint.Tests
	}
}

// Run start processing
func (p *Processor) Run() {
	for i := int64(0); i < p.threads; i++ {
		go func() {
			p.processThread()
		}()
	}
}

// Process send request to internal processing thread
func (p *Processor) Process(rq *ProcessRequest) (resp *ProcessResponse, err error) {
	respChan := make(chan *ProcessResponse)
	rq.ResponseChannel = respChan

	p.rqChan <- rq

	// todo: timeout
	resp = <-respChan

	return resp, nil
}

// processThread waiting for requests
func (p *Processor) processThread() {
	for {
		select {
		case rq := <-p.rqChan:
			resp := p.doProcess(rq)
			rq.ResponseChannel <- resp
		}
	}
}

// doProcess do process request
func (p *Processor) doProcess(rq *ProcessRequest) *ProcessResponse {
	resp := &ProcessResponse{}

	// choose tests
	tests, ok := p.endpoints[rq.Endpoint]
	if !ok {
		resp.Error = errors.New("endpoint not found")
		return resp
	}

	var foundTest *cfg.Test
	for _, test := range tests {
		if ok, err := p.checkCondition(&test, rq.Content, rq.Context); ok {
			foundTest = &test
			break
		} else if err != nil {
			resp.Error = err
			return resp
		}
	}

	if foundTest == nil {
		resp.Error = errors.New("suitable test not found")
		return resp
	}

	resp.Content = []byte(foundTest.Then.Response.(string))
	return resp
}

// checkCondition check test condition
func (p *Processor) checkCondition(test *cfg.Test, body []byte, ctx map[string]interface{}) (ok bool, err error) {
	if test.When == nil {
		return true, nil
	}

	return false, nil
}
