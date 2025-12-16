package main

import (
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/gripper"
	"go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/vision"
	"go.viam.com/utils/trace"

	"github.com/erh/vmodutils/touch"
)

func main() {
	trace.SetTracerWithExporters(otelresource.Empty())
	module.ModularMain(
		resource.APIModel{camera.API, touch.CropCameraModel},
		resource.APIModel{camera.API, touch.DetectCropCameraModel},
		resource.APIModel{camera.API, touch.MergeModel},
		resource.APIModel{camera.API, touch.MultipleArmPosesModel},
		resource.APIModel{toggleswitch.API, touch.ArmPositionSaverModel},
		resource.APIModel{gripper.API, touch.ObstacleModel},
		resource.APIModel{gripper.API, touch.ObstacleOpenBoxModel},
		resource.APIModel{vision.API, touch.ClusterModel},
		resource.APIModel{camera.API, touch.LookAtCameraModel},
	)

}
