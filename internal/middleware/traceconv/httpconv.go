package traceconv

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type HttpConv struct {
	NetConv *netConv
}

type netConv struct {
}

func NewHttpConv() *HttpConv {
	hc := &HttpConv{
		NetConv: &netConv{},
	}

	return hc
}

func (c *HttpConv) HTTPMethod(method string) attribute.KeyValue {
	if method == "" {
		return attribute.Key("http.method").String(http.MethodGet)
	}

	return attribute.Key("http.method").String(method)
}

func (c *HttpConv) HTTPScheme(https bool) attribute.KeyValue {
	if https {
		return attribute.Key("http.scheme").String("https")
	}

	return attribute.Key("http.scheme").String("http")
}

func (c *HttpConv) SplitHostPort(hostport string) (string, int) {
	hosts := strings.Split(hostport, ":")
	port, _ := strconv.Atoi(hosts[1])
	fmt.Println(port)
	return hosts[0], port
}

func (n *netConv) NetHostName(host string) attribute.KeyValue {
	return attribute.Key("net.host.name").String(host)
}

func (n *netConv) NetProtocol(protocol string) []string {
	protos := strings.Split(protocol, "/")
	return protos
}
