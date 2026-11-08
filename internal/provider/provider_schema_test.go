package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var providerConfigType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"host":    tftypes.String,
	"api_key": tftypes.String,
	"org_id":  tftypes.String,
}}

// providerConfigValue builds a provider config; nil means the attribute is not set.
func providerConfigValue(host, apiKey, orgId *string) tftypes.Value {
	str := func(v *string) tftypes.Value {
		if v == nil {
			return tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(tftypes.String, *v)
	}
	return tftypes.NewValue(providerConfigType, map[string]tftypes.Value{
		"host":    str(host),
		"api_key": str(apiKey),
		"org_id":  str(orgId),
	})
}

// TestProviderSchemas validates every resource and data source schema without calling the Apollo API.
func TestProviderSchemas(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	for _, newResource := range p.Resources(ctx) {
		r := newResource()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "apollostudio"}, &meta)
		t.Run(meta.TypeName, func(t *testing.T) {
			var resp resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
			}
			if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
				t.Fatalf("invalid schema: %v", diags)
			}
		})
	}

	for _, newDataSource := range p.DataSources(ctx) {
		d := newDataSource()
		var meta datasource.MetadataResponse
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "apollostudio"}, &meta)
		t.Run("data."+meta.TypeName, func(t *testing.T) {
			var resp datasource.SchemaResponse
			d.Schema(ctx, datasource.SchemaRequest{}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
			}
			if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
				t.Fatalf("invalid schema: %v", diags)
			}
			for name, attr := range resp.Schema.Attributes {
				checkComputedNesting(t, name, attr)
			}
		})
	}

	// The provider server also validates schemas as a whole when Terraform asks for them.
	server, err := providerserver.NewProtocol6WithError(p)()
	if err != nil {
		t.Fatal(err)
	}
	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range schemaResp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("GetProviderSchema error: %s: %s", d.Summary, d.Detail)
		}
	}
}

// TestProviderConfigureRequiresCredentials checks that a missing API key or org ID is reported clearly.
func TestProviderConfigureRequiresCredentials(t *testing.T) {
	// t.Setenv restores the original values after the test; empty values count as unset.
	t.Setenv("APOLLO_KEY", "")
	t.Setenv("APOLLO_ORG_ID", "")

	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{}); err != nil {
		t.Fatal(err)
	}

	config, err := tfprotov6.NewDynamicValue(providerConfigType, providerConfigValue(nil, nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: &config})
	if err != nil {
		t.Fatal(err)
	}

	summaries := map[string]bool{}
	for _, d := range resp.Diagnostics {
		summaries[d.Summary] = true
	}
	for _, want := range []string{"Missing Apollo API key", "Missing Apollo Organization ID"} {
		if !summaries[want] {
			t.Errorf("expected diagnostic %q, got %v", want, summaries)
		}
	}
}

// Duck-typed views of the framework's (internal) schema attribute interfaces.
type schemaAttribute interface {
	IsComputed() bool
	IsOptional() bool
	IsRequired() bool
}

// checkComputedNesting fails when a computed-only nested attribute contains a Required or Optional child:
// Terraform can never receive configuration for such a child. The framework does not catch this itself.
func checkComputedNesting(t *testing.T, path string, attr interface{}) {
	t.Helper()
	parent, ok := attr.(schemaAttribute)
	if !ok {
		return
	}
	for name, child := range nestedChildren(attr) {
		childPath := path + "." + name
		c, ok := child.(schemaAttribute)
		if !ok {
			continue
		}
		if parent.IsComputed() && !parent.IsOptional() && !parent.IsRequired() && (c.IsRequired() || c.IsOptional()) {
			t.Errorf("attribute %s is configurable but its parent %s is computed-only", childPath, path)
		}
		checkComputedNesting(t, childPath, child)
	}
}

// nestedChildren returns the attributes nested under attr.
func nestedChildren(attr interface{}) map[string]interface{} {
	children := map[string]interface{}{}
	switch a := attr.(type) {
	case dsschema.ListNestedAttribute:
		for k, v := range a.NestedObject.Attributes {
			children[k] = v
		}
	case dsschema.SetNestedAttribute:
		for k, v := range a.NestedObject.Attributes {
			children[k] = v
		}
	case dsschema.MapNestedAttribute:
		for k, v := range a.NestedObject.Attributes {
			children[k] = v
		}
	case dsschema.SingleNestedAttribute:
		for k, v := range a.Attributes {
			children[k] = v
		}
	case rsschema.ListNestedAttribute:
		for k, v := range a.NestedObject.Attributes {
			children[k] = v
		}
	case rsschema.SetNestedAttribute:
		for k, v := range a.NestedObject.Attributes {
			children[k] = v
		}
	case rsschema.MapNestedAttribute:
		for k, v := range a.NestedObject.Attributes {
			children[k] = v
		}
	case rsschema.SingleNestedAttribute:
		for k, v := range a.Attributes {
			children[k] = v
		}
	}
	return children
}
