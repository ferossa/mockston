package cfg

import (
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v2"
)

// Suite parameters for mock suite
// describes protocols we are mocking
type Suite struct {
	Connections []Connection `yaml:"connections"`
	Listen      []Listen     `yaml:"listen"`
}

// Connection channel we are getting or sending requests through
type Connection struct {
	Name       string            `yaml:"name"`       // unique name
	Protocol   string            `yaml:"protocol"`   // connection protocol
	Host       string            `yaml:"host"`       // host
	Port       int16             `yaml:"port"`       // port
	Login      string            `yaml:"login"`      // login
	Password   string            `yaml:"password"`   // password
	Properties map[string]string `yaml:"properties"` // additional properties
}

// Listen description of test cases when and what we should respond
type Listen struct {
	Source    string     `yaml:"source"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Endpoint emulated endpoint
// e.g. http url or amqp queue
type Endpoint struct {
	Name       string `yaml:"name"`
	Url        string `yaml:"url"`        // http connection parameter
	Method     string `yaml:"method"`     // http connection parameter
	Exchange   string `yaml:"exchange"`   // amqp connection parameter
	RoutingKey string `yaml:"routingKey"` // amqp connection parameter
	Queue      string `yaml:"queue"`      // amqp connection parameter
	Tests      []Test `yaml:"tests"`
}

// Test one of possible actions
type Test struct {
	When interface{} `yaml:"when"` // condition when this action should be executed
	Then TestAction  `yaml:"then"` // action to execute
}

// TestAction action descriptor
type TestAction struct {
	Response interface{} `yaml:"response"` // data to send in response
}

// LoadSuite parse suite configs
// configs can be split in several files so we are read all yaml in directory
// and after that merge them in one suite
func LoadSuite(workDir string) (*Suite, error) {
	p := path.Join(workDir, "main.yml")
	y, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	s := &Suite{}
	err = yaml.Unmarshal(y, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}
