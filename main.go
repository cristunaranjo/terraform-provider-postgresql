package main

import (
	"flag"

	"github.com/cristunaranjo/terraform-provider-postgresql/postgresql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: postgresql.Provider,
		ProviderAddr: "registry.terraform.io/cristunaranjo/postgresql",
		Debug:        debug,
	})
}
