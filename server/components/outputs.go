package components

import (
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs/mqtt"
)

// createSceneService wires the output drivers of the configured integrations into a scene service.
// New integrations are added here. The caller still has to add the service to the lifecycle manager.
func createSceneService(
	lfManager *lifecycle.Manager,
	configs application.SceneConfigStore,
	states application.SceneStateStore,
	rooms application.RoomConfigStore,
	builtin ...application.OutputDriver,
) (*application.SceneService, *application.DeviceMonitor, error) {
	monitor := application.NewDeviceMonitor(rooms, outputs.GetConfig().ApplyTimeout)

	drivers := make([]application.OutputDriver, 0, len(builtin)+1)
	for _, d := range builtin {
		drivers = append(drivers, monitor.Track(d))
	}

	var listeners []application.SceneListener
	if cfg := mqtt.GetConfig(); cfg.Enabled() {
		driver, err := mqtt.NewDriver(&cfg)
		if err != nil {
			return nil, nil, err
		}
		lfManager.Add(driver)
		drivers = append(drivers, monitor.Track(driver))
		listeners = append(listeners, driver)
	}

	service := application.NewSceneService(configs, states, rooms, drivers...)
	for _, l := range listeners {
		service.AddListener(l)
	}

	return service, monitor, nil
}
