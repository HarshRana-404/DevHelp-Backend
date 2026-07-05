package dto

// JSONGeneratorRequest is the input payload for the JSON Generator tool.
type JSONGeneratorRequest struct {
	// JSON is the raw JSON string to generate type definitions from.
	JSON string `json:"json" binding:"required" example:"{\"name\":\"Alice\",\"age\":30,\"address\":{\"city\":\"NYC\"}}"`
	// RootName is the optional name for the root type. Defaults to "Root".
	RootName string `json:"root_name,omitempty" example:"User"`
}

// JSONGeneratorResponse holds the generated code for each target language.
type JSONGeneratorResponse struct {
	// GoStruct is the generated Go struct definition.
	GoStruct string `json:"go_struct"`
	// TypeScriptInterface is the generated TypeScript interface definition.
	TypeScriptInterface string `json:"typescript_interface"`
}
