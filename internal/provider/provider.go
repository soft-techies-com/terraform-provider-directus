package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/soft-techies-com/terraform-provider-directus/internal/client"
	resourcepkg "github.com/soft-techies-com/terraform-provider-directus/internal/resource"
)

// New returns a new instance of the provider.
func New() provider.Provider {
	return &directusProvider{}
}

// directusProvider implements the Terraform provider.
type directusProvider struct{}

// provider configuration model
type directusProviderModel struct {
	URL          types.String `tfsdk:"url"`
	Token        types.String `tfsdk:"token"`
	TimeoutSec   types.Int64  `tfsdk:"timeout_seconds"`
	InsecureHTTP types.Bool   `tfsdk:"insecure_http"`
}

func (p *directusProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "directus-terraform"
}

func (p *directusProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "Base URL of your Directus instance, e.g. https://directus.example.com",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Static admin token or personal access token with admin privileges.",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional:    true,
				Description: "HTTP client timeout in seconds (default 30).",
			},
			"insecure_http": schema.BoolAttribute{
				Optional:    true,
				Description: "Allow HTTP (no TLS) for local dev.",
			},
		},
	}
}

func (p *directusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg directusProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	base := strings.TrimRight(cfg.URL.ValueString(), "/")
	if base == "" {
		resp.Diagnostics.AddError("invalid url", "Provider requires `url`.")
		return
	}
	if !cfg.InsecureHTTP.ValueBool() && strings.HasPrefix(base, "http://") {
		resp.Diagnostics.AddError("insecure url", "Refusing plain HTTP URL unless `insecure_http = true`.")
		return
	}
	if _, err := url.Parse(base); err != nil {
		resp.Diagnostics.AddError("invalid url", fmt.Sprintf("%v", err))
		return
	}

	timeout := 30 * time.Second
	if !cfg.TimeoutSec.IsNull() && cfg.TimeoutSec.ValueInt64() > 0 {
		timeout = time.Duration(cfg.TimeoutSec.ValueInt64()) * time.Second
	}

	dclient := client.NewDirectus(base, cfg.Token.ValueString(), timeout)

	resp.DataSourceData = dclient
	resp.ResourceData = dclient
}

func (p *directusProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resourcepkg.NewRoleResource,
		resourcepkg.NewPermissionResource,
		resourcepkg.NewSettingResource,
		resourcepkg.NewFileResource,
		resourcepkg.NewPolicyResource,
	}
}

func (p *directusProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
