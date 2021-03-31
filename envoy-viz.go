package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/spf13/pflag"
	"github.com/twosigma/envoy-viz/configreader"
	"github.com/twosigma/envoy-viz/graph"
)

var (
	file     = pflag.String("file", "", "A file to read envoy config from")
	adminUrl = pflag.String("admin", "", "The url the admin api is running on")
	render   = pflag.String("render", "", "How to render the result")
)

func main() {
	pflag.Parse()
	if *file != "" && *adminUrl != "" {
		log.Fatal("Only one of --file or --admin can be set at the same time")
	}
	var bs *v3.Bootstrap
	if *file != "" {
		bootstrap, err := configreader.FromFile(*file)
		if err != nil {
			panic(err)
		}
		bs = bootstrap
	} else if *adminUrl != "" {
		r, err := configreader.ReadEnvoyConfig(*adminUrl)
		if err != nil {
			panic(err)
		}
		ec, err := configreader.ParseEnvoyConfig(r)
		if err != nil {
			panic(err)
		}
		bs = ec.Boostrap.Bootstrap
	} else {
		panic("Must either set --file or --admin option")
	}
	if *render == "" {
		json.NewEncoder(os.Stdout).Encode(&bs)
	} else {
		g, err := graph.BuildGraph(bs)
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
