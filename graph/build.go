package graph

import (
	"bytes"
	"fmt"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extAuthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/golang/protobuf/ptypes"
)

var (
	g = graphviz.New()
)

func BuildGraph(bs *v3.Bootstrap) (*cgraph.Graph, error) {
	graph, err := g.Graph()
	if err != nil {
		return nil, err
	}

	// Downstream represents any requests flowing into envoy
	downstream, err := graph.CreateNode("Downstream")
	downstream.SetRoot(true)
	if err != nil {
		return nil, err
	}

	clusterSubgraph := graph.SubGraph("cluster_clusters", 1)
	clusterSubgraph.SetStyle(cgraph.FilledGraphStyle)
	clusterSubgraph.SetLabel("Clusters")
	clusterSubgraph.SetLabelJust("l")
	// Create cluster nodes first so we can link to them from the filters they are used in
	clusters := map[string]*cgraph.Node{}
	for _, c := range bs.StaticResources.Clusters {
		// For each cluster, create a node
		clusterNode, err := clusterSubgraph.CreateNode("Cluster: " + c.Name)
		if err != nil {
			return nil, err
		}
		clusterNode.SetLabel(fmt.Sprintf("%s\n %v", c.Name, c.GetType()))
		clusters[c.Name] = clusterNode
		if loadAssignment := c.LoadAssignment; loadAssignment != nil {
			for _, la := range loadAssignment.Endpoints {
				for _, lb := range la.LbEndpoints {
					endpoint := lb.GetEndpoint()
					// hostname := endpoint.Hostname
					address := endpoint.Address
					egress := toString(address)
					egressNode, err := graph.CreateNode(egress)
					if err != nil {
						return nil, err
					}
					graph.CreateEdge("", clusterNode, egressNode)
				}
			}
		}
	}
	listenerGraph := graph.SubGraph("cluster_listeners", 1)
	listenerGraph.SetStyle(cgraph.FilledGraphStyle)
	listenerGraph.SetLabel("Listeners")
	listenerGraph.SetLabelJust("l")
	for _, l := range bs.StaticResources.Listeners {
		listener, err := listenerGraph.CreateNode("Listener: " + l.Name)
		if err != nil {
			return nil, err
		}
		edge, err := graph.CreateEdge("", downstream, listener)
		if err != nil {
			return nil, err
		}
		edge.SetLabel(toString(l.Address))
		for _, fc := range l.FilterChains {
			for _, f := range fc.Filters {
				filterName := f.Name
				connectionManagerSubgraph := graph.SubGraph("cluster_"+l.Name+" http cnx mgr", 1)
				connectionManagerSubgraph.SetStyle(cgraph.FilledGraphStyle)
				connectionManagerSubgraph.SetLabel("http_connection_manager:" + l.Name)
				connectionManagerSubgraph.SetLabelJust("l")
				if filterName == "envoy.filters.network.http_connection_manager" {
					filterNode, err := connectionManagerSubgraph.CreateNode(l.Name + " http cnx mngr")
					if err != nil {
						return nil, err
					}
					graph.CreateEdge("", listener, filterNode)
					var m httpv3.HttpConnectionManager
					if err := ptypes.UnmarshalAny(f.GetTypedConfig(), &m); err != nil {
						return nil, err
					}
					lastNode := filterNode
					for _, httpFilter := range m.HttpFilters {
						nextNode, err := connectionManagerSubgraph.CreateNode(l.Name + httpFilter.Name)
						if err != nil {
							return nil, err
						}
						nextNode.SetLabel(httpFilter.Name)
						connectionManagerSubgraph.CreateEdge("", lastNode, nextNode)
						if httpFilter.Name == "envoy.filters.http.ext_authz" {
							var extAuth extAuthv3.ExtAuthz
							if err := ptypes.UnmarshalAny(httpFilter.GetTypedConfig(), &extAuth); err != nil {
								return nil, err
							}
							serverUri := extAuth.GetHttpService().GetServerUri()
							cluster := serverUri.GetCluster()
							edge, err := graph.CreateEdge("", nextNode, clusters[cluster])
							if err != nil {
								return nil, err
							}
							edge.SetLabel(extAuth.GetHttpService().GetPathPrefix())
						}
						lastNode = nextNode
					}
					rc := m.GetRouteConfig()
					routeNode, err := graph.CreateNode("Route: " + rc.Name)
					if err != nil {
						return nil, err
					}
					graph.CreateEdge("", lastNode, routeNode)
					for _, vh := range rc.VirtualHosts {
						//virtualHostName := vh.Name
						// TODO: virtualHostDomains := vh.Domains
						for _, r := range vh.Routes {
							match := r.GetMatch()
							cluster := r.GetRoute().GetCluster()
							if cluster != "" {
								clusterNode := clusters[cluster]
								edge, err := graph.CreateEdge("", routeNode, clusterNode)
								if err != nil {
									return nil, err
								}
								edge.SetLabel(fmt.Sprintf("%v", match))
							}
							if directResponse := r.GetDirectResponse(); directResponse != nil {
								directResponseNode, err := graph.CreateNode("DirectResponse: " + directResponse.String())
								if err != nil {
									return nil, err
								}
								directResponseNode.SetShape("record")
								directResponseNode.SetLabel(fmt.Sprintf("<f0> Direct Response\n|{%d | %s}", directResponse.GetStatus(), directResponse.Body.GetInlineString()))
								edge, err := graph.CreateEdge("", routeNode, directResponseNode)
								if err != nil {
									return nil, err
								}
								edge.SetLabel(fmt.Sprintf("%v", match))
							}
						}
					}
				}
			}
		}
	}
	return graph, nil
}

func Render(graph *cgraph.Graph, render string) ([]byte, error) {
	var buf bytes.Buffer
	if err := g.Render(graph, graphviz.Format(render), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toString(socketAddress *core_v3.Address) string {
	switch v := socketAddress.Address.(type) {
	case *core_v3.Address_SocketAddress:
		return fmt.Sprintf("%s:%d", v.SocketAddress.Address, v.SocketAddress.GetPortValue())
	case *core_v3.Address_Pipe:
		return v.Pipe.Path
	default:
		return ""
	}

}
