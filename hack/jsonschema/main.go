package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	group           = "argoproj.io"
	version         = "v1alpha1"
	eventBusKind    = "EventBus"
	eventSourceKind = "EventSource"
	sensorKind      = "Sensor"
)

type obj = map[string]interface{}

func main() {
	swagger := obj{}
	{
		f, err := os.Open("api/openapi-spec/swagger.json")
		if err != nil {
			panic(err)
		}
		err = json.NewDecoder(f).Decode(&swagger)
		if err != nil {
			panic(err)
		}
	}
	{
		crdKinds := []string{
			eventBusKind,
			eventSourceKind,
			sensorKind,
		}
		definitions := swagger["definitions"]
		oneOf := make([]obj, 0, len(crdKinds))
		for _, kind := range crdKinds {
			lowerCaseKind := strings.ToLower(kind)
			definitionKey := fmt.Sprintf("io.argoproj.%s.%s.%s", lowerCaseKind, version, kind)
			v := definitions.(obj)[definitionKey].(obj)
			v["x-kubernetes-group-version-kind"] = []obj{
				{
					"group":   group,
					"kind":    kind,
					"version": version,
				},
			}
			props := v["properties"].(obj)
			props["apiVersion"].(obj)["const"] = fmt.Sprintf("%s/%s", group, version)
			props["kind"].(obj)["const"] = kind
			oneOf = append(oneOf, obj{"$ref": "#/definitions/" + definitionKey})
		}

		schema := obj{
			"$id":         "http://events.argoproj.io/events.json",
			"$schema":     "http://json-schema.org/schema#",
			"type":        "object",
			"oneOf":       oneOf,
			"definitions": definitions,
		}
		f, err := os.Create("api/jsonschema/schema.json")
		if err != nil {
			panic(err)
		}

		e := json.NewEncoder(f)
		e.SetIndent("", "  ")
		err = e.Encode(schema)
		if err != nil {
			panic(err)
		}

		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
}
