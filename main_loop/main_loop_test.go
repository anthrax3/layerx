package main_loop_test

import (
	. "github.com/layer-x/layerx-core_v2/main_loop"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
"github.com/layer-x/layerx-commons/lxlog"
"github.com/layer-x/layerx-core_v2/fakes"
	"fmt"
	"github.com/layer-x/layerx-commons/lxmartini"
"github.com/layer-x/layerx-core_v2/lxserver"
	"github.com/layer-x/layerx-core_v2/lxstate"
	"github.com/layer-x/layerx-core_v2/layerx_tpi_client"
	"github.com/layer-x/layerx-core_v2/layerx_rpi_client"
	"time"
	"github.com/layer-x/layerx-commons/lxdatabase"
	"github.com/layer-x/layerx-core_v2/task_launcher"
	"github.com/layer-x/layerx-core_v2/health_checker"
)

func PurgeState() {
	lxdatabase.Rmdir("/state", true)
}

var _ = Describe("MainLoop", func() {
	var lxRpiClient *layerx_rpi_client.LayerXRpi
	var lxTpiClient *layerx_tpi_client.LayerXTpi
	var state *lxstate.State
	var serverErr error

	driverErrc := make(chan error)
	var taskLauncher *task_launcher.TaskLauncher

	Describe("setup", func() {
		It("sets up for the tests", func() {
			lxRpiClient = &layerx_rpi_client.LayerXRpi{
				CoreURL: "127.0.0.1:2277",
			}
			lxTpiClient = &layerx_tpi_client.LayerXTpi{
				CoreURL: "127.0.0.1:2277",
			}
			state = lxstate.NewState()
			err := state.InitializeState("http://127.0.0.1:4001")
			Expect(err).To(BeNil())
			coreServerWrapper := lxserver.NewLayerXCoreServerWrapper(state, lxmartini.QuietMartini(), driverErrc)

			err = state.SetTpi( "127.0.0.1:2288")
			Expect(err).To(BeNil())
			err = state.RpiPool.AddRpi(&layerx_rpi_client.RpiInfo{
				Name: "fake-rpi",
				Url: "127.0.0.1:2299",
			})

			taskLauncher = task_launcher.NewTaskLauncher(state)

			go func() {
				for {
					serverErr = <-driverErrc
				}
			}()

			m := coreServerWrapper.WrapServer()
			go m.RunOnAddr(fmt.Sprintf(":2277"))
			go fakes.RunFakeTpiServer("127.0.0.1:2277", 2288, make(chan error))
			go fakes.RunFakeRpiServer("127.0.0.1:2277", 2299, make(chan error))
			lxlog.ActiveDebugMode()
		})
	})

	Describe("MainLoop", func(){
		It("collects tasks from tpi, collects resources from rpi, and launches staging tasks", func(){
			PurgeState()
			err2 := state.InitializeState("http://127.0.0.1:4001")
			Expect(err2).To(BeNil())
			err := state.SetTpi( "127.0.0.1:2288")
			Expect(err).To(BeNil())
			err = state.RpiPool.AddRpi(&layerx_rpi_client.RpiInfo{
				Name: "fake-rpi",
				Url: "127.0.0.1:2299",
			})
			healthChecker := health_checker.NewHealthChecker(state)
			go MainLoop(taskLauncher, healthChecker, state, driverErrc)
			time.Sleep(1000 * time.Millisecond)
			Expect(serverErr).To(BeNil())
		})
	})
})
