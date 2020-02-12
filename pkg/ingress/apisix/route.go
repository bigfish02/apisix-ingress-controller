package apisix

import (
	ingress "github.com/gxthrj/apisix-ingress-types/pkg/apis/config/v1"
	apisix "github.com/gxthrj/apisix-types/pkg/apis/apisix/v1"
	"strconv"
	"github.com/iresty/ingress-controller/pkg/ingress/endpoint"
)

const DefaultLBType = "roundrobin"

type ApisixRoute ingress.ApisixRoute

// Convert convert to  apisix.Route from ingress.ApisixRoute CRD
func (ar *ApisixRoute) Convert() ([]*apisix.Route, []*apisix.Service, []*apisix.Upstream, error) {
	ns := ar.Namespace
	// Host
	rules := ar.Spec.Rules
	routes := make([]*apisix.Route, 0)
	services := make([]*apisix.Service, 0)
	upstreams := make([]*apisix.Upstream, 0)
	rv := ar.ObjectMeta.ResourceVersion
	for _, r := range rules {
		host := r.Host
		paths := r.Http.Paths
		for _, p := range paths {
			uri := p.Path
			svcName := p.Backend.ServiceName
			svcPort := strconv.FormatInt(p.Backend.ServicePort, 10)
			// apisix route name = host + path
			apisixRouteName := host + uri
			// apisix service name = namespace_svcName_svcPort
			apisixSvcName := ns + "_" + svcName + "_" + svcPort
			// apisix route name = namespace_svcName_svcPort = apisix service name
			apisixUpstreamName := ns + "_" + svcName + "_" + svcPort
			// todo plugins

			// routes
			route := &apisix.Route{
				ResourceVersion: &rv,
				Name: &apisixRouteName,
				Host: &host,
				Path: &uri,
				ServiceName: &apisixSvcName,
				UpstreamName: &apisixUpstreamName,
			}
			routes = append(routes, route)
			// services
			service := &apisix.Service{
				Name: &apisixSvcName,
				UpstreamName:  &apisixUpstreamName,
				ResourceVersion: &rv,
			}
			services = append(services, service)
			// upstreams
			LBType := DefaultLBType
			port, _:= strconv.Atoi(svcPort)
			nodes := endpoint.BuildEps(ns, svcName, port)
			upstream := &apisix.Upstream{
				ResourceVersion: &rv,
				Name: &apisixUpstreamName,
				Type: &LBType,
				Nodes: nodes,
			}
			upstreams = append(upstreams, upstream)
		}
	}
	return routes, services, upstreams,nil
}