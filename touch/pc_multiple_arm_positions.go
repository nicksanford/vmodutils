package touch

import (
	"context"
	"fmt"
	"time"

	"go.viam.com/rdk/components/camera"
	toggleswitch "go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot/framesystem"
	"go.viam.com/rdk/spatialmath"

	"github.com/erh/vmodutils"
)

var MultipleArmPosesModel = vmodutils.NamespaceFamily.WithModel("pc-multiple-arm-poses")

func init() {
	resource.RegisterComponent(
		camera.API,
		MultipleArmPosesModel,
		resource.Registration[camera.Camera, *MultipleArmPosesConfig]{
			Constructor: newMultipleArmPoses,
		})
}

type MultipleArmPosesConfig struct {
	Src                   string
	SleepSeconds          float64 `json:"sleep_seconds"`
	Positions             []string
	MergeUsingCameraFrame bool `json:"merge_using_camera_frame"`
}

func (c *MultipleArmPosesConfig) sleepTime() time.Duration {
	if c.SleepSeconds <= 0 {
		return time.Second
	}
	return time.Duration(c.SleepSeconds * float64(time.Second))
}

func (c *MultipleArmPosesConfig) Validate(path string) ([]string, []string, error) {
	if c.Src == "" {
		return nil, nil, fmt.Errorf("need a src camera")
	}

	if len(c.Positions) == 0 {
		return nil, nil, fmt.Errorf("no positions")
	}

	return append(c.Positions, c.Src), nil, nil
}

func newMultipleArmPoses(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (camera.Camera, error) {
	newConf, err := resource.NativeConfig[*MultipleArmPosesConfig](config)
	if err != nil {
		return nil, err
	}

	cc := &MultipleArmPosesCamera{
		name:      config.ResourceName(),
		cfg:       newConf,
		positions: []toggleswitch.Switch{},
	}

	cc.src, err = camera.FromProvider(deps, newConf.Src)
	if err != nil {
		return nil, err
	}

	for _, p := range newConf.Positions {
		s, err := toggleswitch.FromProvider(deps, p)
		if err != nil {
			return nil, err
		}
		cc.positions = append(cc.positions, s)
	}
	cc.fsSvc, err = framesystem.FromDependencies(deps)
	if err != nil {
		return nil, err
	}
	fsConfig, err := cc.fsSvc.FrameSystemConfig(ctx)
	if err != nil {
		return nil, err
	}

	cc.fs, err = referenceframe.NewFrameSystem("", fsConfig.Parts, nil)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

type MultipleArmPosesCamera struct {
	resource.AlwaysRebuild
	resource.TriviallyCloseable

	name resource.Name
	cfg  *MultipleArmPosesConfig

	fsSvc framesystem.Service
	fs    *referenceframe.FrameSystem

	src       camera.Camera
	positions []toggleswitch.Switch
}

func (mapc *MultipleArmPosesCamera) Name() resource.Name {
	return mapc.name
}

func (mapc *MultipleArmPosesCamera) Image(ctx context.Context, mimeType string, extra map[string]interface{}) ([]byte, camera.ImageMetadata, error) {
	return nil, camera.ImageMetadata{}, fmt.Errorf("image not supported")
}

func (mapc *MultipleArmPosesCamera) Images(ctx context.Context, _ []string, _ map[string]interface{}) ([]camera.NamedImage, resource.ResponseMetadata, error) {
	return nil, resource.ResponseMetadata{}, fmt.Errorf("image not supported")
}

func (mapc *MultipleArmPosesCamera) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (mapc *MultipleArmPosesCamera) NextPointCloud(ctx context.Context, extra map[string]interface{}) (pointcloud.PointCloud, error) {
	inputs := []pointcloud.PointCloud{}
	totalSize := 0

	for _, p := range mapc.positions {

		err := p.SetPosition(ctx, 2, nil)
		if err != nil {
			return nil, err
		}

		time.Sleep(mapc.cfg.sleepTime())

		pc, err := mapc.src.NextPointCloud(ctx, extra)
		if err != nil {
			return nil, err
		}

		totalSize += pc.Size()

		inputs = append(inputs, pc)
	}

	big := pointcloud.NewBasicPointCloud(totalSize)
	pif, err := mapc.fsSvc.GetPose(ctx, mapc.src.Name().Name, "", nil, nil)
	if err != nil {
		return nil, err
	}

	for _, pc := range inputs {
		var err error
		if mapc.cfg.MergeUsingCameraFrame {
			err = pointcloud.ApplyOffset(pc, pif.Pose(), big)
		} else {
			err = pointcloud.ApplyOffset(pc, nil, big)
		}
		if err != nil {
			return nil, err
		}
	}

	return big, nil
}

func (mapc *MultipleArmPosesCamera) Properties(ctx context.Context) (camera.Properties, error) {
	return camera.Properties{
		SupportsPCD: true,
	}, nil
}

func (mapc *MultipleArmPosesCamera) Geometries(ctx context.Context, _ map[string]interface{}) ([]spatialmath.Geometry, error) {
	return nil, nil
}
