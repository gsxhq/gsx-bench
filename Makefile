# gsx-bench — runtime render benchmarks: gsx vs templ vs html/template.
#
# Lives outside the gsx module on purpose: pulling templ (and the templ/gsx CLIs)
# in here keeps gsx's own go.mod dependency-free. gsx is wired via a replace
# directive to ../gsx (the local working tree); templ is the published module.

GSX_DIR ?= ../gsx

.PHONY: bench test generate generate-gsx generate-templ tools clean

bench: ## run the render benchmarks
	go test -bench . -benchmem -run '^$$' .

test: ## verify gsx and templ render identical output across scenarios
	go test -v -run 'Agree' .

generate: generate-gsx generate-templ ## regenerate both .go outputs

# gsx is generated from its own module so gsx's tooling deps (x/tools) never
# leak into this module's go.mod — only the gsx runtime library is required here.
# gsxr uses the default class merger; tw/ has its own gsx.toml configuring the
# mock merger, so it is generated with -C tw/ for that config to be discovered.
generate-gsx:
	cd $(GSX_DIR) && go run ./cmd/gsx -C $(CURDIR) generate ./gsxr
	cd $(GSX_DIR) && go run ./cmd/gsx -C $(CURDIR)/tw generate .

# Needs the standalone templ CLI (see `make tools`); running it via `go run`
# from this module would drag the CLI's deps into go.mod.
generate-templ:
	templ generate

tools: ## install the templ CLI (gsx is run from $(GSX_DIR))
	go install github.com/a-h/templ/cmd/templ@v0.3.1020

clean:
	rm -f gsxr/*.x.go tw/*.x.go templr/*_templ.go
