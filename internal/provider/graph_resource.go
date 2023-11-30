package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

var (
	_ resource.Resource                = &GraphResource{}
	_ resource.ResourceWithConfigure   = &GraphResource{}
	_ resource.ResourceWithImportState = &GraphResource{}
)

type GraphResource struct {
	client *client.ApolloClient
}

type GraphResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
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
		Description: "Manage a graph",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the graph. This is an immutable value and cannot be changed and must be unique across all graphs",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 40),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]+$`),
						"must starts with a letter and contains only letters, numbers, and dashes",
					),
				},
			},
			// TODO: add ways to add name_prefix attribute (it would conflict with name attribute)
			"name": schema.StringAttribute{
				Description: "Name of the graph",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 40),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the graph",
				Required:    true,
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
			fmt.Sprintf("Expected *client.ApolloClient got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *GraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Return values from plan
	var plan GraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the graph
	graph, err := r.client.CreateGraph(ctx, plan.Id.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create graph",
			fmt.Sprintf("Failed to create graph: %s", err.Error()),
		)
		return
	}

	// Map response body to schema and populate response
	plan.Id = types.StringValue(graph.Id)
	plan.Name = types.StringValue(graph.Name)
	plan.Description = types.StringValue(graph.Description)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Return values from state
	var state GraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the graph
	graph, err := r.client.GetGraph(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get graph",
			fmt.Sprintf("Failed to get graph: %s", err.Error()),
		)
		return
	}

	// Map response body to schema and populate response
	state.Id = types.StringValue(graph.Id)
	state.Name = types.StringValue(graph.Name)
	state.Description = types.StringValue(graph.Description)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Return values from plan
	var plan GraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Return values from state
	var state GraphResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update description
	if plan.Description.ValueString() != state.Description.ValueString() {
		err := r.client.UpdateGraphDescription(ctx, state.Id.ValueString(), plan.Description.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update graph description",
				fmt.Sprintf("Failed to update graph description: %s", err.Error()),
			)
			return
		}
	}

	// Update name
	if plan.Name.ValueString() != state.Name.ValueString() {
		err := r.client.UpdateGraphName(ctx, state.Id.ValueString(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update graph name",
				fmt.Sprintf("Failed to update graph name: %s", err.Error()),
			)
			return
		}
	}

	// Saved updated values to state
	plan.Id = state.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Return values from state
	var state GraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the graph
	err := r.client.RemoveGraph(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete graph",
			fmt.Sprintf("Failed to delete graph: %s", err.Error()),
		)
		return
	}
}

func (r *GraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
