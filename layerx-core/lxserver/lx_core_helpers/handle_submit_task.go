package lx_core_helpers

import (
	"github.com/emc-advanced-dev/layerx/layerx-core/lxstate"
	"github.com/emc-advanced-dev/layerx/layerx-core/lxtypes"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func SubmitTask(state *lxstate.State, tpId string, task *lxtypes.Task) error {
	taskProvider, err := state.TaskProviderPool.GetTaskProvider(tpId)
	if err != nil {
		return lxerrors.New("getting task provider from pool "+tpId, err)
	}
	task.TaskProvider = taskProvider
	err = state.PendingTaskPool.AddTask(task)
	if err != nil {
		return lxerrors.New("adding task to pending task pool", err)
	}
	return nil
}
