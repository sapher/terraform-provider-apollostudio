package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var (
	_ resource.Resource              = &SubGraphResource{}
	_ resource.ResourceWithConfigure = &SubGraphResource{}
)

type SubGraphResource struct {
	client *client.ApolloClient
}

type SubGraphResourceModel struct {
	GraphId      types.String `tfsdk:"graph_id"`
	VariantName  types.String `tfsdk:"variant_name"`
	SubgraphName types.String `tfsdk:"subgraph_name"`
	Sdl          types.String `tfsdk:"sdl"`
	Revision     types.String `tfsdk:"revision"`
	Url          types.String `tfsdk:"url"`
}

func NewSubGraphResource() resource.Resource {
	return &SubGraphResource{}
}

func (r *SubGraphResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraph"
}

func (r *SubGraphResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a resource to manage a Subgraph.",
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "Graph ID",
				Required:    true,
			},
			"variant_name": schema.StringAttribute{
				Description: "Variant name",
				Required:    true,
			},
			"subgraph_name": schema.StringAttribute{
				Description: "Subgraph name",
				Required:    true,
			},
			"sdl": schema.StringAttribute{
				Description: "SDL",
				Required:    true,
			},
			"revision": schema.StringAttribute{
				Description: "Revision",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL",
				Required:    true,
			},
		},
	}
}

func (r *SubGraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ApolloClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ApolloClientn got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SubGraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Warn(ctx, "Read")

	// Get current state
	var state SubGraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed graph api key from Apollo Studio
	_, err := r.client.GetSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.SubgraphName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subgraph",
			"Could not read subgraph "+state.SubgraphName.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Warn(ctx, "Create")

	// Return values from plan
	var plan SubGraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Publish the subgraph
	err := r.client.PublishSubGraph(
		ctx,
		plan.GraphId.ValueString(),
		plan.VariantName.ValueString(),
		plan.SubgraphName.ValueString(),
		plan.Revision.ValueString(),
		plan.Sdl.ValueString(),
		plan.Url.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error publishing subgraph",
			"Could not publish subgraph "+plan.SubgraphName.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Warn(ctx, "Update")

	// // Return values from plan
	// var plan SubGraphResourceModel
	// diags := req.Plan.Get(ctx, &plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // Get current state
	// var state SubGraphResourceModel
	// diags = req.State.Get(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // Update the API key
	// err := r.client.RenameGraphApiKey(ctx, state.GraphId.ValueString(), state.Id.ValueString(), plan.KeyName.ValueString())
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error updating graph api key",
	// 		"Could not update graph api key "+plan.Id.ValueString()+", unexpected error: "+err.Error(),
	// 	)
	// 	return
	// }

	// // Map response body to schema and populate response
	// plan.Id = state.Id
	// plan.Role = state.Role
	// plan.Token = state.Token
	// plan.CreatedAt = state.CreatedAt

	// diags = resp.State.Set(ctx, plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
}

func (r *SubGraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state SubGraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the API key
	err := r.client.RemoveSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.SubgraphName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subgraph",
			"Could not delete subgraph "+state.SubgraphName.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}
}
