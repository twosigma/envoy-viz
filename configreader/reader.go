package configreader

import (
	"os"
	"strings"

	adminapi "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/twosigma/envoy-viz/configdump"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func FromFile(filepath string) (*v3.Bootstrap, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	// Convert yaml to json
	if strings.HasSuffix(filepath, ".yaml") || strings.HasSuffix(filepath, ".yml") {
		contents, err = yaml.YAMLToJSON(contents)
		if err != nil {
			return nil, err
		}
	}
	var bs v3.Bootstrap
	err = protojson.Unmarshal(contents, &bs)
	return &bs, err
}

func ConfigDumpFromFile(filepath string) (*adminapi.ConfigDump, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var configDump configdump.Wrapper
	err = configDump.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}
	return configDump.ConfigDump, nil
}
