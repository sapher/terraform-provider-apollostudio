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
	_ resource.Resource              = &GraphResource{}
	_ resource.ResourceWithConfigure = &GraphResource{}
)

type GraphResource struct {
	client *client.ApolloClient
}

type GraphResourceModel struct {
	GraphId     types.String `tfsdk:"graph_id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

func NewGraphResource() resource.Resource {
	return &GraphResource{}
}

func (r *GraphResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph"
}

func (r *GraphResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a resource to manage a Graph.",
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "Graph ID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Graph name",
				Required:    true,
			},
			"title": schema.StringAttribute{
				Description: "Graph title",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Graph description",
				Required:    true, // TODO: change this to optional
			},
		},
	}
}

func (r *GraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state GraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the graph
	graph, err := r.client.GetGraph(ctx, state.GraphId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading graph",
			"Could not read graph "+state.GraphId.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate response
	state.GraphId = types.StringValue(graph.Id)
	state.Name = types.StringValue(graph.Name)
	state.Title = types.StringValue(graph.Title)
	state.Description = types.StringValue(graph.Description)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Warn(ctx, "Create")

	// Return values from plan
	var plan GraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the graph
	graph, err := r.client.CreateGraph(ctx, plan.Name.ValueString(), plan.Title.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating graph",
			"Could not create graph "+plan.GraphId.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate response
	plan.GraphId = types.StringValue(graph.Id)
	plan.Name = types.StringValue(graph.Name)
	plan.Title = types.StringValue(graph.Title)
	plan.Description = types.StringValue(graph.Description)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Warn(ctx, "Update")

	// // Return values from plan
	// var plan GraphResourceModel
	// diags := req.Plan.Get(ctx, &plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // Get current state
	// var state GraphResourceModel
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

func (r *GraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state GraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the Graph
	err := r.client.RemoveGraph(ctx, state.GraphId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting graph",
			"Could not delete graph "+state.GraphId.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}
}
