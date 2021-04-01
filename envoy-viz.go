package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/twosigma/envoy-viz/configreader"
	"github.com/twosigma/envoy-viz/graph"
)

var (
	file       = pflag.String("file", "", "A file to read envoy config from")
	adminUrl   = pflag.String("admin", "", "The url the admin api is running on")
	render     = pflag.String("render", "", "How to render the result")
	fullConfig = pflag.BoolP("full-config", "f", false, "Whether to use the full config or just the bootstrap of the envoy config dump")
)

func main() {
	pflag.Parse()
	if *file != "" && *adminUrl != "" {
		log.Fatal("Only one of --file or --admin can be set at the same time")
	}
	var graphable graph.Graphable
	if *file != "" {
		if !*fullConfig {
			bootstrap, err := configreader.FromFile(*file)
			if err != nil {
				panic(err)
			}
			graphable = &graph.GraphableBoostrap{
				Bootstrap: bootstrap,
			}
		} else {
			configDump, err := configreader.ConfigDumpFromFile(*file)
			if err != nil {
				panic(err)
			}
			ec, err := configreader.ParseEnvoyConfig(configDump)
			if err != nil {
				panic(err)
			}
			graphable = &graph.GraphableConfigDump{
				EnvoyConfig: ec,
			}
		}
	} else if *adminUrl != "" {
		r, err := configreader.ReadEnvoyConfig(*adminUrl)
		if err != nil {
			panic(err)
		}
		ec, err := configreader.ParseEnvoyConfig(r)
		if err != nil {
			panic(err)
		}
		if !*fullConfig {
			graphable = &graph.GraphableBoostrap{
				Bootstrap: ec.Boostrap.Bootstrap,
			}
		} else {
			graphable = &graph.GraphableConfigDump{
				EnvoyConfig: ec,
			}
		}
	} else {
		panic("Must either set --file or --admin option")
	}
	if *render == "" {
		json.NewEncoder(os.Stdout).Encode(graphable)
	} else {
		g, err := graph.BuildGraph(graphable)
		if err != nil {
			panic(err)
		}
		b, err := graph.Render(g, *render)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}
