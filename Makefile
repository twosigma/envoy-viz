.PHONY: all build clean test

all:: build assets/static.svg assets/original-dst.svg assets/advanced.svg assets/http-filters.svg assets/envoy-edge.svg

assets:
	mkdir -p assets

assets/static.svg: assets envoy-viz
	./envoy-viz --file ./testdata/static.yaml --render svg > assets/static.svg 

assets/original-dst.svg: assets envoy-viz
	./envoy-viz --file ./testdata/original-dst.yaml --render svg > assets/original-dst.svg 

assets/advanced.svg: assets envoy-viz
	./envoy-viz --file ./testdata/advanced.yaml --render svg > assets/advanced.svg 

assets/http-filters.svg: assets envoy-viz
	./envoy-viz --file ./testdata/http-filters.yaml --render svg > assets/http-filters.svg 

assets/envoy-edge.svg: assets envoy-viz
	./envoy-viz --file ./testdata/envoy-edge.yaml --render svg > assets/envoy-edge.svg

build: envoy-viz

envoy-viz: envoy-viz.go
	go build ./envoy-viz.go

test: main.go
	go test ./...

clean::
	rm -rf ./assets
	rm -f envoy-viz