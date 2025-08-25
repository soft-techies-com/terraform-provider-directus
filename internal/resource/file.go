package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/soft-techies-com/terraform-provider-directus/internal/client"
)

// FileResource implements the directus_file resource
type FileResource struct{ client *client.Directus }

// FileModel represents the file resource model
type FileModel struct {
	ID               types.String `tfsdk:"id"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	Type             types.String `tfsdk:"type"`
	FilenameDisk     types.String `tfsdk:"filename_disk"`
	FilenameDownload types.String `tfsdk:"filename_download"`
	Storage          types.String `tfsdk:"storage"`
	Folder           types.String `tfsdk:"folder"`
	UploadedBy       types.String `tfsdk:"uploaded_by"`
	UploadedOn       types.String `tfsdk:"uploaded_on"`
	ModifiedBy       types.String `tfsdk:"modified_by"`
	ModifiedOn       types.String `tfsdk:"modified_on"`
	Metadata         types.String `tfsdk:"metadata"`
	Checksum         types.String `tfsdk:"checksum"`
	Width            types.Int64  `tfsdk:"width"`
	Height           types.Int64  `tfsdk:"height"`
	Filesize         types.Int64  `tfsdk:"filesize"`
	Duration         types.Int64  `tfsdk:"duration"`
}

// NewFileResource returns a new file resource
func NewFileResource() resource.Resource { return &FileResource{} }

// Metadata returns the resource type name
func (r *FileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

// Schema defines the schema for the resource
func (r *FileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":                rschema.StringAttribute{Computed: true},
			"title":             rschema.StringAttribute{Optional: true},
			"description":       rschema.StringAttribute{Optional: true},
			"type":              rschema.StringAttribute{Optional: true},
			"filename_disk":     rschema.StringAttribute{Optional: true},
			"filename_download": rschema.StringAttribute{Optional: true},
			"storage":           rschema.StringAttribute{Optional: true},
			"folder":            rschema.StringAttribute{Optional: true},
			"uploaded_by":       rschema.StringAttribute{Computed: true},
			"uploaded_on":       rschema.StringAttribute{Computed: true},
			"modified_by":       rschema.StringAttribute{Computed: true},
			"modified_on":       rschema.StringAttribute{Computed: true},
			"metadata":          rschema.StringAttribute{Optional: true},
			"checksum":          rschema.StringAttribute{Optional: true},
			"width":             rschema.Int64Attribute{Optional: true},
			"height":            rschema.Int64Attribute{Optional: true},
			"filesize":          rschema.Int64Attribute{Optional: true},
			"duration":          rschema.Int64Attribute{Optional: true},
		},
	}
}

// Configure configures the resource
func (r *FileResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Directus)
}

// Create a new file
func (r *FileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{}
	if v := plan.Title.ValueString(); v != "" {
		payload["title"] = v
	}
	if v := plan.Description.ValueString(); v != "" {
		payload["description"] = v
	}
	if v := plan.FilenameDownload.ValueString(); v != "" {
		payload["filename_download"] = v
	}
	if v := plan.Storage.ValueString(); v != "" {
		payload["storage"] = v
	}
	if v := plan.Folder.ValueString(); v != "" {
		payload["folder"] = v
	}
	if v := plan.Metadata.ValueString(); v != "" {
		payload["metadata"] = v
	}
	if v := plan.Checksum.ValueString(); v != "" {
		payload["checksum"] = v
	}
	if v := plan.Width.ValueInt64(); v != 0 {
		payload["width"] = v
	}
	if v := plan.Height.ValueInt64(); v != 0 {
		payload["height"] = v
	}
	if v := plan.Filesize.ValueInt64(); v != 0 {
		payload["filesize"] = v
	}
	if v := plan.Duration.ValueInt64(); v != 0 {
		payload["duration"] = v
	}
	if v := plan.Type.ValueString(); v != "" {
		payload["type"] = v
	}
	if v := plan.FilenameDisk.ValueString(); v != "" {
		payload["filename_disk"] = v
	}

	httpResp, err := r.client.Request(ctx, http.MethodPost, "/files", payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	r.readIntoState(ctx, &plan, resp, apiResp.Data)
}

// Read a file
func (r *FileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "/files/" + state.ID.ValueString()
	httpResp, err := r.client.Request(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	r.readIntoState(ctx, &state, resp, apiResp.Data)
}

// Update a file
func (r *FileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{}
	if v := plan.Title.ValueString(); v != "" {
		payload["title"] = v
	} else {
		payload["title"] = nil
	}
	if v := plan.Description.ValueString(); v != "" {
		payload["description"] = v
	} else {
		payload["description"] = nil
	}
	if v := plan.FilenameDownload.ValueString(); v != "" {
		payload["filename_download"] = v
	} else {
		payload["filename_download"] = nil
	}
	if v := plan.Storage.ValueString(); v != "" {
		payload["storage"] = v
	} else {
		payload["storage"] = nil
	}
	if v := plan.Folder.ValueString(); v != "" {
		payload["folder"] = v
	} else {
		payload["folder"] = nil
	}
	if v := plan.Metadata.ValueString(); v != "" {
		payload["metadata"] = v
	} else {
		payload["metadata"] = nil
	}

	var state FileModel
	req.State.Get(ctx, &state)
	url := "/files/" + state.ID.ValueString()
	httpResp, err := r.client.Request(ctx, http.MethodPatch, url, payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	r.readIntoState(ctx, &plan, resp, apiResp.Data)
}

// Delete a file
func (r *FileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Request(ctx, http.MethodDelete, "/files/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// helper: map API data into state
func (r *FileResource) readIntoState(ctx context.Context, plan *FileModel, resp interface{}, data map[string]any) {
	type diagLike interface{ AddError(string, string) }

	plan.ID = types.StringValue(str(data["id"]))
	plan.Title = strPtrToType(data["title"])
	plan.Description = strPtrToType(data["description"])
	plan.Type = strPtrToType(data["type"])
	plan.FilenameDisk = strPtrToType(data["filename_disk"])
	plan.FilenameDownload = strPtrToType(data["filename_download"])
	plan.Storage = strPtrToType(data["storage"])
	plan.Folder = strPtrToType(data["folder"])
	plan.UploadedBy = strPtrToType(data["uploaded_by"])
	plan.UploadedOn = strPtrToType(data["uploaded_on"])
	plan.ModifiedBy = strPtrToType(data["modified_by"])
	plan.ModifiedOn = strPtrToType(data["modified_on"])
	plan.Metadata = strPtrToType(data["metadata"])
	plan.Checksum = strPtrToType(data["checksum"])

	if v, ok := data["width"].(float64); ok {
		plan.Width = types.Int64Value(int64(v))
	}
	if v, ok := data["height"].(float64); ok {
		plan.Height = types.Int64Value(int64(v))
	}
	if v, ok := data["filesize"].(float64); ok {
		plan.Filesize = types.Int64Value(int64(v))
	}
	if v, ok := data["duration"].(float64); ok {
		plan.Duration = types.Int64Value(int64(v))
	}

	switch v := resp.(type) {
	case *resource.CreateResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	case *resource.ReadResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	case *resource.UpdateResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	}
}
