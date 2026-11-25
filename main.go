package main

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/vthiery/steampipe-plugin-serval/serval"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: serval.Plugin})
}
