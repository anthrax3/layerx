package layerx_tpi_test

import (
	. "github.com/layer-x/layerx-core_v2/layerx_tpi"

	"github.com/layer-x/layerx-core_v2/fakes"
	"github.com/layer-x/layerx-core_v2/lxtypes"
	"github.com/mesos/mesos-go/mesosproto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LayerxTpi", func() {

	fakeStatus1 := fakes.FakeTaskStatus("fake_task_id_1", mesosproto.TaskState_TASK_RUNNING)
	fakeStatus2 := fakes.FakeTaskStatus("fake_task_id_2", mesosproto.TaskState_TASK_KILLED)
	fakeStatus3 := fakes.FakeTaskStatus("fake_task_id_3", mesosproto.TaskState_TASK_FINISHED)

	fakeStatuses := []*mesosproto.TaskStatus{fakeStatus1, fakeStatus2, fakeStatus3}

	go fakes.RunFakeLayerXServer(fakeStatuses, 12345)
	lxTpi := LayerXTpi{
		CoreURL: "127.0.0.1:12345",
	}
	Describe("RegisterTaskProvider", func() {
		It("submits a new task provider to the LX Server", func() {
			taskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provder_id",
				Source: "taskprovider@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider)
			Expect(err).To(BeNil())
		})
	})
	Describe("DeregisterTaskProvider", func() {
		It("Requests the server to delete the task provider", func() {
			err := lxTpi.DeregisterTaskProvider("fake_task_provder_id")
			Expect(err).To(BeNil())
			err = lxTpi.DeregisterTaskProvider("fake_task_provder_id")
			Expect(err).ToNot(BeNil())
		})
	})
	Describe("GetTaskProvider(id)", func() {
		It("returns the task provider for the id, or error if it does not exist", func() {
			fakeTaskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provder_id_1",
				Source: "taskprovider1@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(fakeTaskProvider)
			Expect(err).To(BeNil())

			taskProvider, err := lxTpi.GetTaskProvider("fake_task_provder_id_1")
			Expect(err).To(BeNil())
			Expect(taskProvider).To(Equal(fakeTaskProvider))
			taskProvider2, err := lxTpi.GetTaskProvider("fake_task_provder_id_2")
			Expect(err).NotTo(BeNil())
			Expect(taskProvider2).To(BeNil())
		})
	})
	Describe("GetTaskProviders", func() {
		It("returns a list of registered task providers", func() {
			taskProvider1 := &lxtypes.TaskProvider{
				Id:     "fake_task_provder_id_1",
				Source: "taskprovider1@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider1)
			Expect(err).To(BeNil())

			taskProvider2 := &lxtypes.TaskProvider{
				Id:     "fake_task_provder_id_2",
				Source: "taskprovider2@tphost:port",
			}
			err = lxTpi.RegisterTaskProvider(taskProvider2)
			Expect(err).To(BeNil())

			taskProvider3 := &lxtypes.TaskProvider{
				Id:     "fake_task_provder_id_3",
				Source: "taskprovider2@tphost:port",
			}
			err = lxTpi.RegisterTaskProvider(taskProvider3)
			Expect(err).To(BeNil())

			taskProviders, err := lxTpi.GetTaskProviders()
			Expect(err).To(BeNil())
			Expect(taskProviders).To(ContainElement(taskProvider1))
			Expect(taskProviders).To(ContainElement(taskProvider2))
			Expect(taskProviders).To(ContainElement(taskProvider3))
		})
	})
	Describe("GetStatusUpdates", func() {
		It("returns a list of status updates", func() {
			statuses, err := lxTpi.GetStatusUpdates()
			Expect(err).To(BeNil())
			Expect(statuses).To(ContainElement(fakeStatus1))
			Expect(statuses).To(ContainElement(fakeStatus2))
			Expect(statuses).To(ContainElement(fakeStatus3))
		})
	})
	Describe("SubmitTask", func() {
		It("submits a task to the server", func() {
			fakeLxTask := fakes.FakeLXTask("fake_task_id", "fake_task_name", "fake_slave_id", "echo FAKE_COMMAND")
			err := lxTpi.SubmitTask(fakeLxTask)
			Expect(err).To(BeNil())
		})
	})
	Describe("KillTask", func() {
		It("requests server to flag task with KillRequested", func() {
			err := lxTpi.KillTask("fake_task_id")
			Expect(err).To(BeNil())
		})
	})
	Describe("PurgeTask", func() {
		It("requests server to flag remove the task from its database", func() {
			fakeLxTask := fakes.FakeLXTask("fake_task_id", "fake_task_name", "fake_slave_id", "echo FAKE_COMMAND")
			err := lxTpi.SubmitTask(fakeLxTask)
			Expect(err).To(BeNil())
			err = lxTpi.PurgeTask("fake_task_id")
			Expect(err).To(BeNil())
			err = lxTpi.PurgeTask("fake_task_id")
			Expect(err).ToNot(BeNil())
		})
	})

})
