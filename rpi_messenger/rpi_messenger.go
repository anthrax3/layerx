package rpi_messenger
import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-core_v2/lxtypes"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"fmt"
	"github.com/layer-x/layerx-core_v2/layerx_rpi_client"
)

const (
	COLLECT_RESOURCES = "/collect_resources"
	LAUNCH_TASKS = "/launch_tasks"
	KILL_TASK = "/kill_task"
)

func SendResourceCollectionRequest(rpiUrl string) error {
	resp, _, err := lxhttpclient.Post(rpiUrl, COLLECT_RESOURCES, nil, nil)
	if err != nil {
		return lxerrors.New("POSTing COLLECT_RESOURCES to RPI server", err)
	}
	if resp.StatusCode != 202 {
		msg := fmt.Sprintf("POSTing COLLECT_RESOURCES to RPI server; status code was %v, expected 202", resp.StatusCode)
		return lxerrors.New(msg, err)
	}
	return nil
}

func SendLaunchTasksMessage(rpiUrl string, tasksToLaunch []*lxtypes.Task, resourcesToUse []*lxtypes.Resource) error {
	launchTasksMessage := &layerx_rpi_client.LaunchTasksMessage{
		TasksToLaunch: tasksToLaunch,
		ResourcesToUse: resourcesToUse,
	}
	resp, _, err := lxhttpclient.Post(rpiUrl, LAUNCH_TASKS, nil, launchTasksMessage)
	if err != nil {
		return lxerrors.New("POSTing tasksToLaunch to RPI server", err)
	}
	if resp.StatusCode != 202 {
		msg := fmt.Sprintf("POSTing tasksToLaunch to RPI server; status code was %v, expected 202", resp.StatusCode)
		return lxerrors.New(msg, err)
	}
	return nil
}

func SendKillTaskRequest(rpiUrl string, taskId string) error {
	resp, _, err := lxhttpclient.Post(rpiUrl, KILL_TASK+"/"+taskId, nil, nil)
	if err != nil {
		return lxerrors.New("POSTing KillTask request for task "+taskId+" to RPI server", err)
	}
	if resp.StatusCode != 202 {
		msg := fmt.Sprintf("POSTing KillTask request for task "+taskId+" to RPI server; status code was %v, expected 202", resp.StatusCode)
		return lxerrors.New(msg, err)
	}
	return nil
}