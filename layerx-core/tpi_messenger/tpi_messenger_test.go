package tpi_messenger_test

import (
	. "github.com/emc-advanced-dev/layerx/layerx-core/tpi_messenger"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/layerx/layerx-core/fakes"
	"github.com/emc-advanced-dev/layerx/layerx-core/layerx_rpi_client"
	"github.com/emc-advanced-dev/layerx/layerx-core/lxserver"
	"github.com/emc-advanced-dev/layerx/layerx-core/lxstate"
	"github.com/emc-advanced-dev/layerx/layerx-core/lxtypes"
	"github.com/layer-x/layerx-commons/lxdatabase"
	"github.com/layer-x/layerx-commons/lxmartini"
	"github.com/mesos/mesos-go/mesosproto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func PurgeState() {
	lxdatabase.Rmdir("/state", true)
}

var _ = Describe("TpiMessenger", func() {
	var serverErr error
	var state *lxstate.State

	Describe("setup", func() {

		It("sets up for the tests", func() {
			state = lxstate.NewState()
			err := state.InitializeState("http://127.0.0.1:4001")
			Expect(err).To(BeNil())
			driverErrc := make(chan error)
			coreServerWrapper := lxserver.NewLayerXCoreServerWrapper(state, lxmartini.QuietMartini(), driverErrc)

			err = state.SetTpi("127.0.0.1:8866")
			Expect(err).To(BeNil())
			err = state.RpiPool.AddRpi(&layerx_rpi_client.RpiInfo{
				Name: "fake-rpi",
				Url:  "127.0.0.1:8855",
			})

			m := coreServerWrapper.WrapServer()
			go m.RunOnAddr(fmt.Sprintf(":7766"))
			go fakes.RunFakeTpiServer("127.0.0.1:7766", 8866, driverErrc)
			go fakes.RunFakeRpiServer("127.0.0.1:7766", 8855, driverErrc)
			logrus.SetLevel(logrus.DebugLevel)

			go func() {
				for {
					serverErr = <-driverErrc
				}
			}()
		})
	})
	Describe("SendTaskCollectionMessage(tpiUrl string []*lxtypes.TaskProvider)", func() {
		It("sends a task collection request to the TPI", func() {
			PurgeState()
			err2 := state.InitializeState("http://127.0.0.1:4001")
			Expect(err2).To(BeNil())
			fakeTaskProvider1 := fakes.FakeTaskProvider("fake_framework_1", "ff@fakeip1:fakeport")
			err := state.TaskProviderPool.AddTaskProvider(fakeTaskProvider1)
			Expect(err).To(BeNil())
			fakeTaskProvider2 := fakes.FakeTaskProvider("fake_framework_2", "ff@fakeip2:fakeport")
			err = state.TaskProviderPool.AddTaskProvider(fakeTaskProvider2)
			Expect(err).To(BeNil())
			fakeTaskProvider3 := fakes.FakeTaskProvider("fake_framework_3", "ff@fakeip3:fakeport")
			err = state.TaskProviderPool.AddTaskProvider(fakeTaskProvider3)
			Expect(err).To(BeNil())
			fakeTaskProviders := []*lxtypes.TaskProvider{fakeTaskProvider1, fakeTaskProvider2, fakeTaskProvider3}
			err = SendTaskCollectionMessage("127.0.0.1:8866", fakeTaskProviders)
			Expect(err).To(BeNil())
		})
	})
	Describe("SendStatusUpdate(tpiUrl *lxtypes.TaskProvider *mesosproto.TaskStatus)", func() {
		It("sends a status update to the TPI for a specific task, specific task provider", func() {
			fakeStatus1 := fakes.FakeTaskStatus("fake_task_id_1", mesosproto.TaskState_TASK_RUNNING)
			fakeTaskProvider1 := fakes.FakeTaskProvider("fake_framework_1", "ff@fakeip1:fakeport")
			err := SendStatusUpdate("127.0.0.1:8866", fakeTaskProvider1, fakeStatus1)
			Expect(err).To(BeNil())
		})
	})
	Describe("HealthCheck(tpiUrl *lxtypes.TaskProvider)", func() {
		Context("the task provider is no longer connected", func() {
			It("returns false", func() {
				fakeTaskProvider1 := fakes.FakeTaskProvider("fake_framework_1_"+fakes.FAIL_ON_PURPOSE, "ff@fakeip1:fakeport")
				healthy, err := HealthCheck("127.0.0.1:8866", fakeTaskProvider1)
				Expect(err).To(BeNil())
				Expect(healthy).To(BeFalse())
			})
		})
		Context("the task provider is still connected", func() {
			It("returns true", func() {
				fakeTaskProvider1 := fakes.FakeTaskProvider("fake_framework_1_", "ff@fakeip1:fakeport")
				healthy, err := HealthCheck("127.0.0.1:8866", fakeTaskProvider1)
				Expect(err).To(BeNil())
				Expect(healthy).To(BeTrue())
			})
		})
	})

})
