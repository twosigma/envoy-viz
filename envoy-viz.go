package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/blevz/envoy-viz/configreader"
	"github.com/blevz/envoy-viz/graph"
	"github.com/spf13/pflag"
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
	if *file != "" {
		bs, err := configreader.FromFile(*file)
		if err != nil {
			panic(err)
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
		return
	}
	r, err := configreader.ReadEnvoyConfig(*adminUrl)
	if err != nil {
		panic(err)
	}
	ec, err := configreader.ParseEnvoyConfig(r)
	if err != nil {
		panic(err)
	}
	if ec != nil {
		json.NewEncoder(os.Stdout).Encode(&ec)
	}

}
