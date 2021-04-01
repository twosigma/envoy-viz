package configreader

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
)

func TestDumper(t *testing.T) {
	tests := []struct {
		name           string
		inputFile      string
		expectedOutput *v3.Bootstrap
	}{
		{name: "basic", inputFile: "../testdata/static.yaml", expectedOutput: staticBootstrap},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancelFunc := context.WithCancel(context.Background())
			defer cancelFunc()
			cmd := exec.CommandContext(ctx, "envoy", "-c", test.inputFile)
			go func() {
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(output))
			}()
			time.Sleep(time.Second)
			configDump, err := ReadEnvoyConfig("http://localhost:9901")
			if err != nil {
				t.Error(err)
				return
			}
			envoyConfig, err := ParseEnvoyConfig(configDump)
			if err != nil {
				t.Error(err)
				return
			}
			envoyConfig.Boostrap.Bootstrap.Node = nil

			compareBootstrap(t, test.expectedOutput, envoyConfig.Boostrap.Bootstrap)
		})
	}
}
