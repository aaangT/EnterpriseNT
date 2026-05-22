// .\helpers\mappings.go

package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type LookupConfig struct {
	From         string `json:"from"`
	LocalField   string `json:"localField"`
	ForeignField string `json:"foreignField"`
	As           string `json:"as"`
}

type MappingConfig struct {
	FilePath       string            `json:"filePath"`
	Database       string            `json:"database"`
	Collection     string            `json:"collection"`
	Entity         string            `json:"entity"`
	Description    string            `json:"description"`
	Lookups        []LookupConfig    `json:"lookups"`
	Fields         map[string]string `json:"fields"`
	ComputedFields map[string]string `json:"computedFields"`
}

func LoadMapping(filePath string) (*MappingConfig, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo mapping %s: %w", filePath, err)
	}

	var mapping MappingConfig

	err = json.Unmarshal(fileContent, &mapping)
	if err != nil {
		return nil, fmt.Errorf("error interpretando JSON del mapping %s: %w", filePath, err)
	}

	if mapping.Database == "" {
		return nil, fmt.Errorf("mapping inválido: database está vacío en %s", filePath)
	}

	if mapping.Collection == "" {
		return nil, fmt.Errorf("mapping inválido: collection está vacío en %s", filePath)
	}

	if len(mapping.Fields) == 0 {
		return nil, fmt.Errorf("mapping inválido: fields está vacío en %s", filePath)
	}

	return &mapping, nil
}

func GetNestedValue(doc map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")

	var current interface{} = doc

	for _, part := range parts {
		switch value := current.(type) {

		case map[string]interface{}:
			nextValue, exists := value[part]
			if !exists {
				return nil, false
			}
			current = nextValue

		case []interface{}:
			if len(value) == 0 {
				return nil, false
			}

			firstItem, ok := value[0].(map[string]interface{})
			if !ok {
				return nil, false
			}

			nextValue, exists := firstItem[part]
			if !exists {
				return nil, false
			}
			current = nextValue

		default:
			return nil, false
		}
	}

	return current, true
}

func MapDocument(doc map[string]interface{}, fields map[string]string) map[string]interface{} {
	mappedDoc := make(map[string]interface{})

	for technicalField, friendlyField := range fields {
		value, exists := GetNestedValue(doc, technicalField)
		if exists {
			mappedDoc[friendlyField] = value
		}
	}

	return mappedDoc
}
