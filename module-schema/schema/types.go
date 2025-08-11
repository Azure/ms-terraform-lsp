package schema

type TerraformObjects map[string]*TerraformObject

type TerraformObject struct {
	Name       string                      `json:"name"`
	Fields     map[string]*SchemaAttribute `json:"fields"`
	ExampleHCL string                      `json:"example_hcl"`
	Details    string                      `json:"details"` // Start from first h2 header (after description)
}
