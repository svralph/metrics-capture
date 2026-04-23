package main

import (
	"bov-2/metricscapture"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

func main() {
	module.ModularMain(resource.APIModel{API: generic.API, Model: metricscapture.Model})
}
