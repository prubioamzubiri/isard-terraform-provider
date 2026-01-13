package provider

import (
	"context"

	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tknika/terraform-provider-isard/internal/client"
)

// Ensure IsardProvider satisfies various provider interfaces.
var _ provider.Provider = &IsardProvider{}

// IsardProvider defines the provider implementation.
type IsardProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// IsardProviderModel describes the provider data model.
type IsardProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	AuthMethod  types.String `tfsdk:"auth_method"`
	CathegoryID types.String `tfsdk:"cathegory_id"`
	Token       types.String `tfsdk:"token"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IsardProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *IsardProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "isard"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *IsardProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Isard VDI.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "EndPoint of the Isard VDI Server",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Authentication token for API access",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method to use (token or credentials)",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("token", "form"),
				},
			},
			"cathegory_id": schema.StringAttribute{
				MarkdownDescription: "Cathegory ID to scope the operations",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure prepares a IsardProvider for data source and resource operations.
func (p *IsardProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var data IsardProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.CathegoryID.IsNull() {
		data.CathegoryID = types.StringValue("default")
	}

	if data.AuthMethod.ValueString() == "form" {
		if data.Username.IsNull() || data.Password.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When using 'form' authentication method, both 'username' and 'password' must be provided.",
			)
			return
		}
	}

	if data.AuthMethod.ValueString() == "token" {
		if data.Token.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When using 'token' authentication method, 'token' must be provided.",
			)
			return
		}
	}

	// Configuration values are now available.

	// Create the client
	c := client.NewClient(data.Endpoint.ValueString(), data.Token.ValueString())

	// Authenticate
	// SignIn manejará "salm" y "form". Si es "token", no hará nada (ya tenemos el token).
	err := c.SignIn(
		data.AuthMethod.ValueString(),
		data.CathegoryID.ValueString(),
		data.Username.ValueString(),
		data.Password.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Authenticate",
			fmt.Sprintf("Failed to sign in: %s", err),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources defines the resources implemented in the provider.
func (p *IsardProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *IsardProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
	}
}
