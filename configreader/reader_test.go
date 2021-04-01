package configreader

import (
	"encoding/json"
	"testing"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	v36 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	v31 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	v35 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	http_manger_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	staticBootstrap *v3.Bootstrap
)

func init() {
	httpConnectionManager, err := ptypes.MarshalAny(&http_manger_v3.HttpConnectionManager{
		StatPrefix: "ingress_http",
		CodecType:  http_manger_v3.HttpConnectionManager_AUTO,
		RouteSpecifier: &http_manger_v3.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoy_config_route_v3.RouteConfiguration{
				Name: "local_route",
				VirtualHosts: []*envoy_config_route_v3.VirtualHost{
					{
						Name:    "local_service",
						Domains: []string{"*"},
						Routes: []*envoy_config_route_v3.Route{
							{

								Match: &envoy_config_route_v3.RouteMatch{
									PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
										Prefix: "/",
									},
								},
								Action: &envoy_config_route_v3.Route_Route{
									Route: &envoy_config_route_v3.RouteAction{
										ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
											Cluster: "some_service",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		HttpFilters: []*http_manger_v3.HttpFilter{
			{Name: "envoy.filters.http.router"},
		},
	})
	if err != nil {
		panic(err)
	}
	staticBootstrap = &v3.Bootstrap{
		Admin: &v3.Admin{
			AccessLogPath: "/tmp/admin_access.log",
			Address: &envoy_config_core_v3.Address{
				Address: &envoy_config_core_v3.Address_SocketAddress{
					SocketAddress: &envoy_config_core_v3.SocketAddress{
						Address: "127.0.0.1",
						PortSpecifier: &envoy_config_core_v3.SocketAddress_PortValue{
							PortValue: 9901,
						},
					},
				},
			},
		},
		StaticResources: &v3.Bootstrap_StaticResources{
			Clusters: []*v36.Cluster{
				{
					Name: "some_service",
					ClusterDiscoveryType: &v36.Cluster_Type{
						Type: v36.Cluster_STATIC,
					},
					ConnectTimeout: &duration.Duration{Nanos: 250000000},
					LoadAssignment: &v31.ClusterLoadAssignment{
						ClusterName: "some_service",
						Endpoints: []*v31.LocalityLbEndpoints{
							{
								LbEndpoints: []*v31.LbEndpoint{
									{
										HostIdentifier: &envoy_config_endpoint_v3.LbEndpoint_Endpoint{
											Endpoint: &envoy_config_endpoint_v3.Endpoint{
												Address: &envoy_config_core_v3.Address{
													Address: &envoy_config_core_v3.Address_SocketAddress{
														SocketAddress: &envoy_config_core_v3.SocketAddress{
															Address: "127.0.0.1",
															PortSpecifier: &envoy_config_core_v3.SocketAddress_PortValue{
																PortValue: 1234,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Listeners: []*v35.Listener{
				{
					Name: "listener_0",
					Address: &envoy_config_core_v3.Address{
						Address: &envoy_config_core_v3.Address_SocketAddress{
							SocketAddress: &envoy_config_core_v3.SocketAddress{
								Address: "127.0.0.1",
								PortSpecifier: &envoy_config_core_v3.SocketAddress_PortValue{
									PortValue: 10000,
								},
							},
						},
					},
					FilterChains: []*envoy_config_listener_v3.FilterChain{
						{
							Filters: []*envoy_config_listener_v3.Filter{
								{
									Name: "envoy.filters.network.http_connection_manager",
									ConfigType: &v35.Filter_TypedConfig{
										TypedConfig: httpConnectionManager,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestReader(t *testing.T) {
	tests := []struct {
		name           string
		inputFile      string
		expectedOutput *v3.Bootstrap
	}{
		{name: "Basic", inputFile: "../testdata/static.yaml", expectedOutput: staticBootstrap},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bootstrap, err := FromFile(test.inputFile)
			if err != nil {
				t.Error(err)
				return
			}
			compareBootstrap(t, test.expectedOutput, bootstrap)
		})
	}
}

func compareBootstrap(t *testing.T, expected, actual *v3.Bootstrap) {
	t.Helper()
	expectedBytes, err := protojson.Marshal(expected)
	if err != nil {
		t.Error(err)
		return
	}
	var expectedMap map[string]interface{}
	if err := json.Unmarshal(expectedBytes, &expectedMap); err != nil {
		t.Error(err)
		return
	}
	actualBytes, err := protojson.Marshal(actual)
	if err != nil {
		t.Error(err)
		return
	}
	var actualMap map[string]interface{}
	if err := json.Unmarshal(actualBytes, &actualMap); err != nil {
		t.Error(err)
		return
	}
	if diff := cmp.Diff(expectedMap, actualMap); diff != "" {
		t.Errorf("Diff: %s", diff)
	}

}
