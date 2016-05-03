package main_loop

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-core_v2/health_checker"
	"github.com/layer-x/layerx-core_v2/lxstate"
	"github.com/layer-x/layerx-core_v2/lxtypes"
	"github.com/layer-x/layerx-core_v2/rpi_messenger"
	"github.com/layer-x/layerx-core_v2/task_launcher"
	"github.com/layer-x/layerx-core_v2/tpi_messenger"
)

var mainLoopLock = &sync.Mutex{}

//run as goroutine
func MainLoop(taskLauncher *task_launcher.TaskLauncher, healthChecker *health_checker.HealthChecker, state *lxstate.State, driverErrc chan error) {
	for {
		errc := make(chan error)
		go func() {
			result := singleExeuction(state, taskLauncher, healthChecker)
			errc <- result
		}()
		err := <-errc
		if err != nil {
			driverErrc <- lxerrors.New("main loop failed while running", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func singleExeuction(state *lxstate.State, taskLauncher *task_launcher.TaskLauncher, healthChecker *health_checker.HealthChecker) error {
	mainLoopLock.Lock()
	defer mainLoopLock.Unlock()
	taskProviderMap, err := state.TaskProviderPool.GetTaskProviders()
	if err != nil {
		return lxerrors.New("retrieving list of task providers from state", err)
	}
	taskProviders := []*lxtypes.TaskProvider{}
	for _, taskProvider := range taskProviderMap {
		taskProviders = append(taskProviders, taskProvider)
	}
	tpiErr := tpi_messenger.SendTaskCollectionMessage(state.GetTpiUrl(), taskProviders)
	if tpiErr != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Warnf("failed sending task collection message to tpi. Is Tpi connected?")
	}
	var rpiErr error
	for _, rpiUrl := range state.GetRpiUrls() {
		rpiErr = rpi_messenger.SendResourceCollectionRequest(rpiUrl)
		if rpiErr != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Warnf("failed sending resource collection request to rpi. Is Rpi connected?")
		}
	}

	if tpiErr != nil || rpiErr != nil {
		return nil
	}

	err = healthChecker.FailDisconnectedTaskProviders()
	if err != nil {
		return lxerrors.New("failing disconnected task providers", err)
	}

	err = healthChecker.ExpireTimedOutTaskProviders()
	if err != nil {
		return lxerrors.New("expiring timed out providers", err)
	}

	err = taskLauncher.LaunchStagedTasks()
	if err != nil {
		return lxerrors.New("launching staged tasks", err)
	}

	return nil
}
