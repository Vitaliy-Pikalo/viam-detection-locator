// Command module is the viam-server entrypoint for the detection-locator module.
package main

import (
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	detectionlocator "github.com/Vitaliy-Pikalo/viam-detection-locator"
)

func main() {
	module.ModularMain(resource.APIModel{API: sensor.API, Model: detectionlocator.Locator})
}
