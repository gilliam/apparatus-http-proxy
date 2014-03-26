package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type ServiceRegistryClient struct {
	httpClient *http.Client
	url        *url.URL
}

type Instance struct {
	Service  string            `json: service`
	Instance string            `json: instance`
	Host     string            `json: host`
	Ports    map[string]string `json: ports`
}

func MakeServiceRegistryClient(envvar string) (*ServiceRegistryClient, error) {
	if !strings.HasPrefix(envvar, "http://") {
		envvar = "http://" + envvar
	}
	serviceRegistryURL, err := url.ParseRequestURI(envvar)
	if err != nil {
		return nil, err
	}
	client := &ServiceRegistryClient{&http.Client{}, serviceRegistryURL}
	return client, nil
}

type InstanceMap map[string]Instance

func (client *ServiceRegistryClient) queryFormation(formation string) (InstanceMap, error) {
	var instances InstanceMap
	resp, err := client.httpClient.Get(client.url.String() + "/" + formation)
	if err != nil {
		log.Print("cannot read instances from service registry")
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (client *ServiceRegistryClient) makeNetloc(inst Instance, port int) string {
	instance_port, _ := strconv.Atoi(inst.Ports[strconv.Itoa(port)])
	return fmt.Sprintf("%s:%d", inst.Host, instance_port)
}

func (client *ServiceRegistryClient) Query(formation, service string, port int) (string, error) {
	instances, err := client.queryFormation(formation)
	if err != nil {
		log.Print("cannot read instances from service registry")
		return "", err
	}
	for _, inst := range instances {
		if inst.Service == service {
			return client.makeNetloc(inst, port), nil
		}
	}
	return "", nil
}

func (client *ServiceRegistryClient) QuerySpecificInstance(formation, service, instance string, port int) (string, error) {
	instances, err := client.queryFormation(formation)
	if err != nil {
		log.Print("cannot read instances from service registry")
		return "", err
	}
	for _, inst := range instances {
		if inst.Service == service && inst.Instance == instance {
			return client.makeNetloc(inst, port), nil
		}
	}
	return "", nil
}

func splitNetloc(url *url.URL) (string, int, error) {
	var port int
	var host string
	if strings.Contains(url.Host, ":") {
		comps := strings.Split(url.Host, ":")
		host = comps[0]
		var err error
		port, err = strconv.Atoi(comps[1])
		if err != nil {
			return "", 0, err
		}
	} else {
		host = url.Host
		if url.Scheme == "http" {
			port = 80
		} else {
			port = 443
		}
	}
	return host, port, nil
}

type Route struct {
	formation string
	role      string
	instance  string
	port      int
}

func ParseRouteFromRequest(request *http.Request) (*Route, error) {
	host, port, err := splitNetloc(request.URL)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(host, ".")
	if strings.HasSuffix(host, ".service") {
		l := len(parts)
		if l == 4 {
			// <instance>.<role>.<formation>.service
			return &Route{parts[2], parts[1], parts[0], port}, nil
		} else if l == 3 {
			// <role>.<formation>.service
			return &Route{parts[1], parts[0], "", port}, nil
		}
	}
	return nil, errors.New("cannot parse")
}

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	client, err := MakeServiceRegistryClient(os.Getenv("GILLIAM_SERVICE_REGISTRY"))
	if err != nil {
	}

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			route, err := ParseRouteFromRequest(r)
			if err != nil {
				// just ignore and pass on the request
				return r, nil
			}
			var netloc string
			if route.instance != "" {
				netloc, err = client.QuerySpecificInstance(route.formation,
					route.role, route.instance, route.port)
			} else {
				netloc, err = client.Query(route.formation, route.role,
					route.port)
			}
			if err != nil {
				return r, nil
			}
			r.URL.Host = netloc
			return r, nil
		})

	log.Fatal(http.ListenAndServe(":4100", proxy))
}
