package crd

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/controlplane-com/libs-go/pkg/types"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/yaml"
	"strings"
)

// ConvertStructToCRD creates a CustomResourceDefinition (apiextensions.k8s.io/v1)
// from a Go struct using reflection to build the OpenAPIV3 schema.
func ConvertStructToCRD(
	exampleObj interface{},
	group, version, kind, plural string,
) (*apiextensionsv1.CustomResourceDefinition, error) {

	objType := reflect.TypeOf(exampleObj)
	if objType == nil {
		return nil, fmt.Errorf("cannot reflect on a nil type")
	}

	// Follow pointer(s) to a concrete type if necessary
	concreteType, _ := types.FollowPointersToConcreteType(objType)

	properties := deriveSchema(concreteType).Properties
	extendStatus(properties)
	properties["org"] = apiextensionsv1.JSONSchemaProps{
		Description: "The organization that owns the resource",
		Type:        "string",
	}

	kind = strings.ToLower(kind)
	required := []string{"org"}
	if isGvcScoped(kind) {
		required = append(required, "gvc")
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", strings.ToLower(plural), group),
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   kind,
				Plural: plural,
			},
			Scope: "Namespaced",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: version,
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Properties: properties,
							Type:       "object",
							Required:   required,
						},
					},
				},
			},
		},
	}

	return crd, nil
}

// deriveSchema handles structs, slices/arrays, and maps.
// If we dig into these “containers” and eventually reach a scalar field,
// we call deriveScalarSchema on the scalar type.
func deriveSchema(t reflect.Type) *apiextensionsv1.JSONSchemaProps {
	switch t.Kind() {

	case reflect.Struct:
		// Build an object schema with Properties for each field
		s := &apiextensionsv1.JSONSchemaProps{
			Type:       "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{},
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			fieldName := strcase.ToLowerCamel(field.Name)
			jsonTag, ok := field.Tag.Lookup("json")
			if ok {
				p := strings.Split(jsonTag, ",")
				if len(p) > 0 {
					fieldName = p[0]
				}
			}

			// Recursively derive schema for each field
			fieldType, _ := types.FollowPointersToConcreteType(field.Type)
			fieldSchema := deriveSchema(fieldType)
			s.Properties[fieldName] = *fieldSchema
		}
		return s

	case reflect.Slice, reflect.Array:
		// Build an array schema. We examine the element type.
		elemType, _ := types.FollowPointersToConcreteType(t.Elem())
		itemSchema := deriveSchema(elemType)
		return &apiextensionsv1.JSONSchemaProps{
			Type: "array",
			Items: &apiextensionsv1.
				JSONSchemaPropsOrArray{Schema: itemSchema},
		}

	case reflect.Map:
		// Build an object schema with AdditionalProperties describing the value type
		valType, _ := types.FollowPointersToConcreteType(t.Elem())
		valSchema := deriveSchema(valType)
		return &apiextensionsv1.JSONSchemaProps{
			Type: "object",
			AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
				Schema: valSchema,
			},
		}
	case reflect.Interface:
		if t.NumMethod() > 0 {
			return deriveScalarSchema(t.Elem())
		}
		return deriveScalarSchema(t)
	default:
		//Everything else is a scalar
		return deriveScalarSchema(t)
	}
}

// deriveScalarSchema handles primitive types: strings, bools, ints, floats.
// Everything else (struct, slice, map, array) is delegated to deriveSchema.
func deriveScalarSchema(t reflect.Type) *apiextensionsv1.JSONSchemaProps {
	// We assume t is already “pointer-unwrapped” by the caller
	switch t.Kind() {
	case reflect.String:
		return &apiextensionsv1.JSONSchemaProps{Type: "string"}
	case reflect.Bool:
		return &apiextensionsv1.JSONSchemaProps{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &apiextensionsv1.JSONSchemaProps{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &apiextensionsv1.JSONSchemaProps{Type: "number"}
	default:
		// If it's not one of the recognized scalar kinds, fallback or treat as string
		return &apiextensionsv1.JSONSchemaProps{Type: "string"}
	}
}

func isGvcScoped(kind string) bool {
	switch kind {
	case "workload":
		return true
	case "volumeset":
		return true
	case "identity":
		return true
	default:
		return false
	}
}

// CRDToYAML marshals a CRD object into YAML.
func CRDToYAML(crd *apiextensionsv1.CustomResourceDefinition) (string, error) {
	// 1) Convert typed CRD -> JSON bytes
	jsonBytes, err := json.Marshal(crd)
	if err != nil {
		return "", fmt.Errorf("marshal to JSON: %w", err)
	}

	// 2) Unmarshal into map[string]interface{}
	var unstructuredCRD map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &unstructuredCRD); err != nil {
		return "", fmt.Errorf("unmarshal to map: %w", err)
	}

	// 3) Remove unwanted fields from the map

	// Remove status
	delete(unstructuredCRD, "status")

	// Remove creationTimestamp
	if meta, ok := unstructuredCRD["metadata"].(map[string]interface{}); ok {
		delete(meta, "creationTimestamp")
	}

	// If you want to remove other fields like managedFields, finalizers, etc.:
	// delete(meta, "managedFields")
	// delete(meta, "finalizers")
	// ... etc.

	// 4) Marshal the cleaned map to YAML
	y, err := yaml.Marshal(unstructuredCRD)
	if err != nil {
		return "", fmt.Errorf("yaml marshal: %w", err)
	}

	return string(y), nil
}
