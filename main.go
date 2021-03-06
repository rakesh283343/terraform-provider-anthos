package main

import (
	"os"

	"github.com/MayaraCloud/terraform-provider-anthos/debug"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func main() {
	if os.Getenv("GO_DEBUG") != "" {
		debug.DebugMode = true
	}
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return Provider()
		},
	})
}
