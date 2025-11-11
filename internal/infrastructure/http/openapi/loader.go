package openapi

import (
	_ "embed"
	"log"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pb33f/libopenapi"
)

//go:embed openapi.yaml
var embeddedYAML []byte

var (
	initOnce sync.Once
	yamlBuf  []byte
	jsonBuf  []byte
)

// Init parses and validates the embedded OpenAPI YAML.
func Init() {
	initOnce.Do(func() {
		yamlBuf = embeddedYAML

		// Validate with libopenapi (3.1 support)
		_, err := libopenapi.NewDocument(embeddedYAML)
		if err != nil {
			log.Printf("openapi: warning: spec failed libopenapi parse: %v", err)
		}

		// Produce JSON using kin-openapi for serving
		loader := &openapi3.Loader{IsExternalRefsAllowed: true}
		doc, err := loader.LoadFromData(embeddedYAML)
		if err != nil {
			log.Printf("openapi: warning: unable to load spec into kin-openapi: %v", err)
			jsonBuf = nil
			return
		}
		b, err := doc.MarshalJSON()
		if err != nil {
			log.Printf("openapi: warning: unable to marshal spec to JSON: %v", err)
			jsonBuf = nil
			return
		}
		jsonBuf = b
	})
}

func GetYAML() []byte {
	Init()
	return yamlBuf
}

func GetJSON() []byte {
	Init()
	return jsonBuf
}


