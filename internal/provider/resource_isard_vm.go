package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tknika/terraform-provider-isard/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &vmResource{}
	_ resource.ResourceWithConfigure = &vmResource{}
)

// NewVMResource is a helper function to simplify the provider implementation.
func NewVMResource() resource.Resource {
	return &vmResource{}
}

// vmResource is the resource implementation.
type vmResource struct {
	client *client.Client
}

// vmResourceModel maps the resource schema data.
type vmResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	TemplateID  types.String  `tfsdk:"template_id"`
	VCPUs       types.Int64   `tfsdk:"vcpus"`
	Memory      types.Float64 `tfsdk:"memory"`
}

// Metadata returns the resource type name.
func (r *vmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

// Schema defines the schema for the resource.
func (r *vmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gestiona un persistent desktop en Isard VDI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identificador único del desktop",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Nombre del desktop (mínimo 4 caracteres, máximo 50)",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Descripción del desktop (máximo 255 caracteres)",
			},
			"template_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID de la plantilla a utilizar para crear el desktop",
			},
			"vcpus": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "N\u00famero de CPUs virtuales (por defecto usa el del template)",
			},
			"memory": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Memoria RAM en GB (por defecto usa la del template)",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *vmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

	// Create creates a new resource.
func (r *vmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preparar hardware personalizado si se especifica
	var vcpus *int64
	var memory *float64
	
	if !plan.VCPUs.IsNull() && !plan.VCPUs.IsUnknown() {
		v := plan.VCPUs.ValueInt64()
		vcpus = &v
	}
	
	if !plan.Memory.IsNull() && !plan.Memory.IsUnknown() {
		m := plan.Memory.ValueFloat64()
		memory = &m
	}

	// Crear el persistent desktop usando la API
	desktopID, err := r.client.CreatePersistentDesktop(
		plan.Name.ValueString(),
		plan.Description.ValueString(),
		plan.TemplateID.ValueString(),
		vcpus,
		memory,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creando el persistent desktop",
			fmt.Sprintf("No se pudo crear el desktop: %s", err.Error()),
		)
		return
	}

	// Actualizar el plan con el ID devuelto por la API
	plan.ID = types.StringValue(desktopID)

	// Obtener los valores finales del desktop creado (incluyendo valores por defecto)
	desktop, err := r.client.GetDesktop(desktopID)
	if err == nil {
		if desktop.Description != "" {
			plan.Description = types.StringValue(desktop.Description)
		}
		if desktop.VCPUs > 0 {
			plan.VCPUs = types.Int64Value(desktop.VCPUs)
		}
		if desktop.Memory > 0 {
			plan.Memory = types.Float64Value(desktop.Memory)
		}
	}

	// Escribir el estado
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Obtener el desktop de la API
	desktop, err := r.client.GetDesktop(state.ID.ValueString())
	if err != nil {
		// Si el desktop no existe (404), eliminarlo del estado
		if err.Error() == "desktop not found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error leyendo el desktop",
			fmt.Sprintf("No se pudo leer el desktop (ID: %s): %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Actualizar el estado con los valores de la API
	state.Name = types.StringValue(desktop.Name)
	state.Description = types.StringValue(desktop.Description)
	state.TemplateID = types.StringValue(desktop.TemplateID)
	
	// Actualizar hardware si está disponible
	if desktop.VCPUs > 0 {
		state.VCPUs = types.Int64Value(desktop.VCPUs)
	}
	if desktop.Memory > 0 {
		state.Memory = types.Float64Value(desktop.Memory)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 1. Usar r.client para actualizar el recurso
	// err := r.client.UpdateSomething(plan.ID.ValueString(), plan.Name.ValueString())
	// ...

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Eliminar el desktop usando la API
	err := r.client.DeleteDesktop(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error eliminando la máquina virtual",
			fmt.Sprintf("No se pudo eliminar la VM (ID: %s): %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// El estado se elimina automáticamente si la función termina sin errores
}
