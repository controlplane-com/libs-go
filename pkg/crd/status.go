package crd

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

func extendStatus(base map[string]apiextensionsv1.JSONSchemaProps) {
	status, ok := base["status"]
	if !ok {
		base["status"] = apiextensionsv1.JSONSchemaProps{}
		status = base["status"]
	}
	if status.Properties == nil {
		status.Properties = map[string]apiextensionsv1.JSONSchemaProps{}
	}
	status.Type = "object"
	if status.AdditionalProperties != nil {
		status.AdditionalProperties = nil
	}
	status.Properties["operator"] = apiextensionsv1.JSONSchemaProps{
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"lastSyncedGeneration": {
				Type: "number",
			},
			"lastProcessedGeneration": {
				Type: "number",
			},
			"validationError": {
				Type: "string",
			},
			"healthStatusMessage": {
				Type: "string",
			},
			"lastSyncTime": {
				Type:   "string",
				Format: "datetime",
			},
			"syncRetries": {
				Type: "number",
			},
			"downstreamOnly": {
				Type: "boolean",
			},
		},
		Type: "object",
	}
	status.Properties["phase"] = apiextensionsv1.JSONSchemaProps{
		Type: "string",
	}
	status.Properties["conditions"] = apiextensionsv1.JSONSchemaProps{
		Items: &apiextensionsv1.JSONSchemaPropsOrArray{
			Schema: &apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"type": {
						Type: "string",
					},
					"status": {
						Type: "string",
					},
				},
				Type: "object",
			},
		},
		Type: "array",
	}
	base["status"] = status
}
