package conn

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ferossa/mockston/internal/cfg"
)

// regexp to detect dynamic routes
var dynamicRouteRegexp *regexp.Regexp

func init() {
	dynamicRouteRegexp = regexp.MustCompile("\\{(\\w+)\\}")
}

// HttpConnector
type HttpConnector struct {
	config           cfg.Connection
	connectionString string
	routes           map[string]map[string]*route
	handler          MessageHandler
}

type route struct {
	name   string
	tester *regexp.Regexp
	params []string
}

// addEndpoint add one endpoint from config
func (c *HttpConnector) addEndpoint(endpoint *cfg.Endpoint) error {
	var ok bool
	if _, ok = c.routes[endpoint.Method]; !ok {
		c.routes[endpoint.Method] = map[string]*route{}
	}

	if _, ok = c.routes[endpoint.Method][endpoint.Url]; ok {
		return fmt.Errorf("duplicate path %s for method %s", endpoint.Url, endpoint.Method)
	}

	route := &route{name: endpoint.Name}

	route.params = getParamsFromPath(endpoint.Url)
	if len(route.params) > 0 {
		route.tester = getDynamicRouteTester(endpoint.Url, route.params)
	}

	c.routes[endpoint.Method][endpoint.Url] = route
	return nil
}

// SetEndpoints set endpoints to process
func (c *HttpConnector) SetEndpoints(endpoints []cfg.Endpoint, h MessageHandler) error {
	c.handler = h
	c.routes = map[string]map[string]*route{}

	var err error
	for _, endpoint := range endpoints {
		if err = c.addEndpoint(&endpoint); err != nil {
			return err
		}
	}

	return nil
}

// Connect start listening for data
func (c *HttpConnector) Connect() error {
	c.connectionString = c.config.Host + ":" + strconv.FormatInt(int64(c.config.Port), 10)
	return http.ListenAndServe(
		c.connectionString,
		c,
	)
}

// ServeHTTP handle incoming requests
func (c *HttpConnector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Requested %s from %s", r.RequestURI, r.RemoteAddr)

	// enable CORS
	if origin := r.Header.Get("origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// read request
	var body []byte
	var err error
	if r.ContentLength > 0 {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// check static routes
	var foundRoute *route
	if route, ok := c.routes[r.Method][r.RequestURI]; ok && route.tester == nil {
		foundRoute = route
	}

	// check dynamic routes if static route not found
	ctx := map[string]interface{}{}
	if foundRoute == nil {
		for _, route := range c.routes[r.Method] {
			if route.tester == nil {
				// this is static route, skip
				continue
			}

			if route.tester.MatchString(r.RequestURI) {
				sm := route.tester.FindAllStringSubmatch(r.RequestURI, -1)
				for i, m := range sm[0][1:] {
					log.Println("add " + route.params[i] + " as " + m)
					ctx["path."+route.params[i]] = m
				}

				foundRoute = route
				break
			}
		}
	}

	if foundRoute == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response, err := c.handler(foundRoute.name, body, ctx)
	if err != nil {
		log.Println("processing error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

// getParamsFromPath parse dynamic url and extract all path parameters
func getParamsFromPath(path string) []string {
	params := dynamicRouteRegexp.FindAllStringSubmatch(path, -1)
	res := make([]string, len(params))
	for i, param := range params {
		res[i] = param[1]
	}

	return res
}

// getDynamicRouteTester build regexp to test url
func getDynamicRouteTester(path string, params []string) *regexp.Regexp {
	for _, param := range params {
		path = strings.Replace(path, "{"+param+"}", "(\\w+)", 1)
	}

	return regexp.MustCompile(path)
}
