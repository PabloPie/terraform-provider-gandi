package main

import (
	"github.com/PabloPie/terraform-provider-gandi/gandi"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: gandi.Provider,
	})
}
