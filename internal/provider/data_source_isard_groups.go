package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tknika/terraform-provider-isard/internal/client"
)

var _ datasource.DataSource = &groupsDataSource{}

func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

type groupsDataSource struct {
	client *client.Client
}

type groupsDataSourceModel struct {
	ID          types.String   `tfsdk:"id"`
	NameFilter  types.String   `tfsdk:"name_filter"`
	CategoryID  types.String   `tfsdk:"category_id"`
	Groups      []groupModel   `tfsdk:"groups"`
}

type groupModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ParentCategory types.String `tfsdk:"parent_category"`
}

func (d *groupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of available groups from Isard VDI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier for the data source.",
				Computed:    true,
			},
			"name_filter": schema.StringAttribute{
				Description: "Optional filter to match group names (case-insensitive substring match).",
				Optional:    true,
			},
			"category_id": schema.StringAttribute{
				Description: "Optional filter to match groups by category ID.",
				Optional:    true,
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of groups available.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Group ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Group name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Group description.",
							Computed:    true,
						},
						"parent_category": schema.StringAttribute{
							Description: "Parent category ID.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *groupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.GetGroups()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read groups, got error: %s", err))
		return
	}

	// Apply filters if provided
	nameFilter := data.NameFilter.ValueString()
	categoryFilter := data.CategoryID.ValueString()
	
	var filteredGroups []client.Group
	for _, group := range groups {
		// Apply name filter
		if nameFilter != "" && !containsIgnoreCase(group.Name, nameFilter) {
			continue
		}
		
		// Apply category filter
		if categoryFilter != "" && group.ParentCategory != categoryFilter {
			continue
		}
		
		filteredGroups = append(filteredGroups, group)
	}

	// Map filtered groups to model
	data.Groups = make([]groupModel, len(filteredGroups))
	for i, group := range filteredGroups {
		data.Groups[i] = groupModel{
			ID:             types.StringValue(group.ID),
			Name:           types.StringValue(group.Name),
			Description:    types.StringValue(group.Description),
			ParentCategory: types.StringValue(group.ParentCategory),
		}
	}

	data.ID = types.StringValue("groups")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
