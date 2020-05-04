package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/plugin"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
    "os"
	"gitlab.com/mayara/private/anthos/debug"
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