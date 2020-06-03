package app

import (
	"flag"
	"log"
	"os"

	"github.com/ferossa/mockston/internal/cfg"
	"github.com/ferossa/mockston/internal/conn"
)

// Mockston main class
type Mockston struct {
	connections map[string]conn.IConnector
	listeners   []IListener
}

// NewMockston create Mockston object
func NewMockston() *Mockston {
	return new(Mockston)
}

// Run start processing
func (m *Mockston) Run() {
	var err error

	// read command line parameters
	var workDir string
	flag.StringVar(&workDir, "d", "", "suite directory")

	flag.Parse()

	if workDir == "" {
		workDir, err = os.Getwd()
		if err != nil {
			log.Fatalln("Unable to get current directory")
		}
	}

	// read suite
	suite, err := cfg.LoadSuite(workDir)
	if err != nil {
		log.Fatalln(err)
	}

	// create connectors
	m.connections = map[string]conn.IConnector{}
	for _, connConfig := range suite.Connections {
		m.connections[connConfig.Name] = conn.NewConnector(connConfig)
	}

	c := make(chan []byte)

	// start listeners
	for _, listenConfig := range suite.Listen {
		// setup processor for specific connection
		p := NewProcessor(1)
		p.SetEndpoints(listenConfig.Endpoints)
		p.Run()

		l := NewListener(m.connections[listenConfig.Source], nil, p)
		m.listeners = append(m.listeners, l)
		go func(lc cfg.Listen) {
			log.Println("listening", lc.Source)
			if err := l.Listen(lc); err != nil {
				log.Println("error connecting to", lc.Source)
			}
		}(listenConfig)
	}

	// waiting for termination
	<-c
}
