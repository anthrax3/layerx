package lx_core_helpers

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/layerx/layerx-core/lxstate"
	"github.com/emc-advanced-dev/layerx/layerx-core/rpi_messenger"
	"github.com/emc-advanced-dev/layerx/layerx-core/tpi_messenger"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/mesos/mesos-go/mesosproto"
)

func KillTask(state *lxstate.State, tpiUrl, taskProviderId, taskId string) error {
	if _, err := state.GetTaskFromAnywhere(taskId); err != nil {
		logrus.WithFields(logrus.Fields{"task_id": taskId, "task_provider": taskProviderId}).Warnf("requested to kill a task that Layer-X has no knowledge of, replying with TASK_LOST")
		err = sendTaskKilledStatus(state, mesosproto.TaskState_TASK_LOST, tpiUrl, taskProviderId, taskId)
		if err != nil {
			return lxerrors.New("sending TASK_KILLED status to task provider"+taskProviderId, err)
		}
		return nil
	}
	taskPool, err := state.GetTaskPoolContainingTask(taskId)
	if err != nil {
		return lxerrors.New("could not find task pool containing task "+taskId, err)
	}
	//if task is staging or pending, just delete it and say we did
	if taskPool == state.PendingTaskPool || taskPool == state.StagingTaskPool {
		logrus.WithFields(logrus.Fields{"task_id": taskId, "task_provider": taskProviderId}).Warnf("requested to kill a task before task staging was complete, deleting from pool")
		err = taskPool.DeleteTask(taskId)
		if err != nil {
			return lxerrors.New("deleting task from staging or pending pool after kill was requested", err)
		}
		err = sendTaskKilledStatus(state, mesosproto.TaskState_TASK_KILLED, tpiUrl, taskProviderId, taskId)
		if err != nil {
			return lxerrors.New("sending TASK_KILLED status to task provider"+taskProviderId, err)
		}
		return nil
	}

	taskToKill, err := taskPool.GetTask(taskId)
	if err != nil {
		return lxerrors.New("could not find task pool containing task "+taskId, err)
	}

	node, err := state.NodePool.GetNode(taskToKill.NodeId)
	if err != nil {
		return lxerrors.New("getting node for task "+taskId, err)
	}

	rpi, err := state.RpiPool.GetRpi(node.GetRpiName())
	if err != nil {
		return lxerrors.New("getting rpi "+node.GetRpiName(), err)
	}
	err = rpi_messenger.SendKillTaskRequest(rpi.Url, taskId)
	if err != nil {
		return lxerrors.New("rpi did not respond to kill task request", err)
	}

	taskToKill.KillRequested = true
	err = taskPool.ModifyTask(taskId, taskToKill)
	if err != nil {
		return lxerrors.New("could not task with KillRequested set back into task pool", err)
	}
	return nil
}

func sendTaskKilledStatus(state *lxstate.State, taskState mesosproto.TaskState, tpiUrl, taskProviderId, taskId string) error {
	taskProvider, err := state.TaskProviderPool.GetTaskProvider(taskProviderId)
	if err != nil {
		return lxerrors.New("finding task provider for kill request", err)
	}
	taskKilledStatus := generateTaskStatus(taskId, taskState, "Kill Task was requested before task staging was complete")
	err = tpi_messenger.SendStatusUpdate(tpiUrl, taskProvider, taskKilledStatus)
	if err != nil {
		return lxerrors.New("udpating tpi with TASK_KILLED status for task before task staging was complete", err)
	}
	return nil
}
