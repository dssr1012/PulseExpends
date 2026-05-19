module pulse-expends-mcp-test

go 1.21

require (
    github.com/dssr1012/pulse-expends v0.0.0
    github.com/stretchr/testify v1.8.4
    github.com/go-chi/chi/v5 v5.0.10
)

replace github.com/dssr1012/pulse-expends => ../..

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    github.com/stretchr/objx v0.5.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)