package configreader

import (
	"io"
	"net/http"

	envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/encoding/protojson"
)

type EnvoyConfig struct {
	Boostrap     envoy_admin_v3.BootstrapConfigDump
	Cluster      envoy_admin_v3.ClustersConfigDump
	Endpoints    envoy_admin_v3.EndpointsConfigDump
	Listeners    envoy_admin_v3.ListenersConfigDump
	ScopedRoutes envoy_admin_v3.ScopedRoutesConfigDump
	Routes       envoy_admin_v3.RoutesConfigDump
}

func ReadEnvoyConfig(adminUrl string) (*envoy_admin_v3.ConfigDump, error) {
	result, err := http.Get(adminUrl + "/config_dump")
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	all, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	var configDump envoy_admin_v3.ConfigDump
	err = protojson.Unmarshal(all, &configDump)
	return &configDump, err
}

func ParseEnvoyConfig(dump *envoy_admin_v3.ConfigDump) (*EnvoyConfig, error) {
	var e EnvoyConfig
	if err := ptypes.UnmarshalAny(dump.Configs[0], &e.Boostrap); err != nil {
		return nil, err
	}
	if err := ptypes.UnmarshalAny(dump.Configs[1], &e.Cluster); err != nil {
		return nil, err
	}
	if err := ptypes.UnmarshalAny(dump.Configs[2], &e.Listeners); err != nil {
		return nil, err
	}
	if err := ptypes.UnmarshalAny(dump.Configs[3], &e.ScopedRoutes); err != nil {
		return nil, err
	}
	if err := ptypes.UnmarshalAny(dump.Configs[4], &e.Routes); err != nil {
		return nil, err
	}
	return &e, nil
}
