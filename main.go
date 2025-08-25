package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/soft-techies-com/terraform-provider-directus/internal/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/soft-techies-com/directus-terraform",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}

// -------------------------------------------------
// Examples (copy into ./examples/basic/main.tf)
// -------------------------------------------------
//
// terraform {
//   required_providers {
//     directus = {
//       source = "your-org/directus"
//       version = ">= 0.1.0"
//     }
//   }
// }
//
// provider "directus" {
//   url   = var.directus_url
//   token = var.directus_token
// }
//
// resource "directus_role" "editor" {
//   name         = "Editor"
//   description  = "Editorial staff"
//   app_access   = true
//   admin_access = false
// }
//
// resource "directus_permission" "article_read" {
//   role_id    = directus_role.editor.id
//   collection = "articles"
//   action     = "read"
//   fields     = "*"
//   permissions = jsonencode({
//     _and = [ { status = { _eq = "published" } } ]
//   })
// }
//
// resource "directus_setting" "instance" {
//   project_name   = "My Directus Project"
//   project_url    = "https://example.com"
//   default_locale = "en-US"
// }
//
// output "editor_role_id" { value = directus_role.editor.id }
//
// -------------------------------------------------
// Build & Test
// -------------------------------------------------
// $ go mod tidy
// $ go build -o terraform-provider-directus
//
// For local testing, place the binary under: ~/.terraform.d/plugins/your-org/directus/0.1.0/linux_amd64/
// and name it "terraform-provider-directus_v0.1.0". Then run:
// $ cd examples/basic && terraform init && terraform apply -auto-approve
//
// Notes:
// - This is a minimal provider focused on project initialization. Extend with more resources
//   (flows, presets, webhooks, folders, files) following the same pattern.
// - Directus API versions can differ (e.g., default_language vs default_locale). This code
//   handles the common fields and avoids double-encoding JSON for permission rules.
// - Use admin-level token for creation of roles/permissions.
