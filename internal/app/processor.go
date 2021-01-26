package app

import (
	"errors"
	"fmt"
	"reflect"

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
	marshaller IMarshaller
	threads    int64
	rqChan     chan *ProcessRequest
	endpoints  map[string][]cfg.Test
}

// NewProcessor create new Processor
func NewProcessor(m IMarshaller, threads int64) *Processor {
	return &Processor{
		marshaller: m,
		threads:    threads,
		rqChan:     make(chan *ProcessRequest),
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

	// preprocess request
	rqMap, err := p.unmarshalRequest(rq.Endpoint, rq.Content)
	if err != nil {
		resp.Error = err
		return resp
	}

	// choose tests
	tests, ok := p.endpoints[rq.Endpoint]
	if !ok {
		resp.Error = errors.New("endpoint not found")
		return resp
	}

	var foundTest *cfg.Test
	for _, test := range tests {
		if ok, err := p.checkCondition(&test, rqMap, rq.Context); ok {
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

// unmarshalRequest unpack message to map
func (p *Processor) unmarshalRequest(endpoint string, body []byte) (map[string]interface{}, error) {
	rqMap := map[string]interface{}{}
	err := p.marshaller.Unmarshal(endpoint, body, &rqMap)
	return rqMap, err
}

// checkCondition check test condition
func (p *Processor) checkCondition(test *cfg.Test, rq map[string]interface{}, ctx map[string]interface{}) (bool, error) {
	if test.When == nil {
		return true, nil
	}

	// TODO: need to rethink how to compare test with real values
	whenVal := reflect.ValueOf(test.When)
	if whenVal.Kind() != reflect.Map {
		return false, nil
	}

	for _, key := range whenVal.MapKeys() {
		testKey := key.Interface().(string)
		testValue := whenVal.MapIndex(key)
		if rqValue, ok := rq[testKey]; ok {
			if fmt.Sprintf("%v", testValue.Interface()) != fmt.Sprintf("%v", rqValue) {
				return false, nil
			}
		} else {
			return false, nil
		}
	}

	return true, nil
}
