package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"

	"sdd-cli/internal/domain"
)

type SchemaValidator struct{}

func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

func (validator *SchemaValidator) ValidateYAML(baseDir, schemaFile string, data []byte) error {
	var value interface{}
	if err := yaml.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("%w: parse YAML: %v", domain.ErrSchemaValidation, err)
	}
	return validator.ValidateValue(baseDir, schemaFile, value)
}

func (validator *SchemaValidator) ValidateValue(baseDir, schemaFile string, value interface{}) error {
	schemaPath, err := containedPath(filepath.Join(baseDir, ".sdd"), "schemas", schemaFile)
	if err != nil {
		return err
	}
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("%w: read %s: %v", domain.ErrSchemaValidation, schemaFile, err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat = true
	if err := compiler.AddResource(schemaFile, bytes.NewReader(schemaData)); err != nil {
		return fmt.Errorf("%w: load %s: %v", domain.ErrSchemaValidation, schemaFile, err)
	}
	schema, err := compiler.Compile(schemaFile)
	if err != nil {
		return fmt.Errorf("%w: compile %s: %v", domain.ErrSchemaValidation, schemaFile, err)
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: encode value: %v", domain.ErrSchemaValidation, err)
	}
	var jsonValue interface{}
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	if err := decoder.Decode(&jsonValue); err != nil {
		return fmt.Errorf("%w: decode value: %v", domain.ErrSchemaValidation, err)
	}
	if err := schema.Validate(jsonValue); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrSchemaValidation, err)
	}

	return nil
}
