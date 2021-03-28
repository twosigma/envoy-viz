package configreader

import (
	"os"
	"strings"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
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
