package lxserver


import (
	"github.com/go-martini/martini"
	"net/http"
	"github.com/layer-x/layerx-commons/lxactionqueue"
	"github.com/layer-x/layerx-core_v2/lxstate"
	"github.com/Sirupsen/logrus"
"github.com/layer-x/layerx-commons/lxlog"
	"encoding/json"
	"github.com/layer-x/layerx-core_v2/layerx_tpi_client"
"io/ioutil"
	"github.com/layer-x/layerx-core_v2/lxserver/lx_core_helpers"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-core_v2/layerx_rpi_client"
"github.com/layer-x/layerx-core_v2/lxtypes"
	"github.com/mesos/mesos-go/mesosproto"
	"github.com/golang/protobuf/proto"
)

const (
//tpi
	RegisterTpi             = "/RegisterTpi"
	RegisterTaskProvider   = "/RegisterTaskProvider"
	DeregisterTaskProvider = "/DeregisterTaskProvider"
	GetTaskProviders       = "/GetTaskProviders"
	GetStatusUpdates       = "/GetStatusUpdates"
	SubmitTask             = "/SubmitTask"
	KillTask               = "/KillTask"
	PurgeTask              = "/PurgeTask"
//rpi
	RegisterRpi             = "/RegisterRpi"
	SubmitResource             = "/SubmitResource"
	SubmitStatusUpdate         = "/SubmitStatusUpdate"
	GetNodes         = "/GetNodes"
)

var empty = []byte{}

type layerxCoreServerWrapper struct {
	actionQueue      lxactionqueue.ActionQueue
	state			*lxstate.State
}

func NewLayerXCoreServerWrapper(state *lxstate.State, actionQueue lxactionqueue.ActionQueue) *layerxCoreServerWrapper {
	return &layerxCoreServerWrapper{
		actionQueue: actionQueue,
		state: state,
	}
}

func (wrapper *layerxCoreServerWrapper) WrapServer(m *martini.ClassicMartini, tpiUrl, rpiUrl string, driverErrc chan error) *martini.ClassicMartini {
	registerTpiHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			data, err := ioutil.ReadAll(req.Body)
			if req.Body != nil {
				defer req.Body.Close()
			}
			if err != nil {
				return empty, 400, lxerrors.New("parsing register tpi request", err)
			}
			var tpiRegistrationMessage layerx_tpi_client.TpiRegistrationMessage
			err = json.Unmarshal(data, &tpiRegistrationMessage)
			if err != nil {
				return empty, 500, lxerrors.New("could not parse json to register tpi message", err)
			}
			err = lx_core_helpers.RegisterTpi(wrapper.state, tpiRegistrationMessage.TpiUrl)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle register tpi request")
				return empty, 500, lxerrors.New("could not handle register tpi request", err)
			}
			lxlog.Infof(logrus.Fields{"tpi_url": tpiRegistrationMessage.TpiUrl}, "Registered TPI to LayerX")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing register tpi message")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	registerRpiHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			data, err := ioutil.ReadAll(req.Body)
			if req.Body != nil {
				defer req.Body.Close()
			}
			if err != nil {
				return empty, 400, lxerrors.New("parsing register rpi request", err)
			}
			var rpiRegistrationMessage layerx_rpi_client.RpiRegistrationMessage
			err = json.Unmarshal(data, &rpiRegistrationMessage)
			if err != nil {
				return empty, 500, lxerrors.New("could not parse json to register rpi message", err)
			}
			err = lx_core_helpers.RegisterRpi(wrapper.state, rpiRegistrationMessage.RpiUrl)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle register rpi request")
				return empty, 500, lxerrors.New("could not handle register rpi request", err)
			}
			lxlog.Infof(logrus.Fields{"rpi_url": rpiRegistrationMessage.RpiUrl}, "Registered TPI to LayerX")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing register rpi message")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	registerTaskProviderHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			data, err := ioutil.ReadAll(req.Body)
			if req.Body != nil {
				defer req.Body.Close()
			}
			if err != nil {
				return empty, 400, lxerrors.New("parsing register TaskProvider request", err)
			}
			var taskProvider lxtypes.TaskProvider
			err = json.Unmarshal(data, &taskProvider)
			if err != nil {
				return empty, 500, lxerrors.New("could not parse json to register TaskProvider message", err)
			}
			err = lx_core_helpers.RegisterTaskProvider(wrapper.state, &taskProvider)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle register TaskProvider request")
				return empty, 500, lxerrors.New("could not handle register TaskProvider request", err)
			}
			lxlog.Infof(logrus.Fields{"task_provider": taskProvider}, "Added new TaskProvider to LayerX")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing register TaskProvider message")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	deregisterTaskProviderHandler := func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		fn := func() ([]byte, int, error) {
			tpId := params["task_provider_id"]
			err := lx_core_helpers.DeregisterTaskProvider(wrapper.state, tpId)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle deregister TaskProvider request")
				return empty, 500, lxerrors.New("could not handle deregister TaskProvider request", err)
			}
			lxlog.Infof(logrus.Fields{"task_provider_id": tpId}, "removed task provider from LayerX")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing deregister TaskProvider message")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	getTaskProvidersHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			taskProviders, err := lx_core_helpers.GetTaskProviders(wrapper.state)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle Get TaskProviders request")
				return empty, 500, lxerrors.New("could not handle Get TaskProviders request", err)
			}
			data, err := json.Marshal(taskProviders)
			if err != nil {
				lxlog.Errorf(logrus.Fields{}, "could not marshal task providers to json")
				return empty, 500, lxerrors.New("marshalling taskProviders to json", err)
			}
			lxlog.Debugf(logrus.Fields{"task_providers": taskProviders}, "Added new TaskProvider to LayerX")
			return data, 200, nil
		}
		data, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing Get TaskProviders message")
			driverErrc <- err
			return
		}
		res.Write(data)
	}

	getStatusUpdatesHandler := func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		fn := func() ([]byte, int, error) {
			tpId := params["task_provider_id"]
			statuses, err := lx_core_helpers.GetStatusUpdates(wrapper.state, tpId)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle Get Status Updates request")
				return empty, 500, lxerrors.New("could not handle Get Status Updates request", err)
			}
			data, err := json.Marshal(statuses)
			if err != nil {
				lxlog.Errorf(logrus.Fields{}, "could not marshal Status Updates to json")
				return empty, 500, lxerrors.New("marshalling Statuses to json", err)
			}
			return data, 200, nil
		}
		data, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing Get Status Updates message")
			driverErrc <- err
			return
		}
		res.Write(data)
	}

	submitTaskHandler := func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		fn := func() ([]byte, int, error) {
			tpId := params["task_provider_id"]
			data, err := ioutil.ReadAll(req.Body)
			if req.Body != nil {
				defer req.Body.Close()
			}
			if err != nil {
				return empty, 400, lxerrors.New("parsing SubmitTask request", err)
			}
			var task lxtypes.Task
			err = json.Unmarshal(data, &task)
			if err != nil {
				return empty, 500, lxerrors.New("could not parse json to task", err)
			}
			err = lx_core_helpers.SubmitTask(wrapper.state, tpId, &task)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle Get SubmitTask request")
				return empty, 500, lxerrors.New("could not handle SubmitTask request", err)
			}
			lxlog.Infof(logrus.Fields{"task_provider_id": tpId, "task": task}, "accepted task from task provider")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing Get Status Updates message")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	killTaskHandler := func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		fn := func() ([]byte, int, error) {
			taskId := params["task_id"]
			err := lx_core_helpers.KillTask(wrapper.state, taskId)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle kill task request")
				return empty, 500, lxerrors.New("could not handle kill task request", err)
			}
			lxlog.Infof(logrus.Fields{"task_id": taskId}, "killed task")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "killing task")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	purgeTaskHandler := func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		fn := func() ([]byte, int, error) {
			taskId := params["task_id"]
			err := lx_core_helpers.PurgeTask(wrapper.state, taskId)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle purge task request")
				return empty, 500, lxerrors.New("could not handle purge task request", err)
			}
			lxlog.Infof(logrus.Fields{"task_id": taskId}, "removed task from LayerX")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "purging task")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	submitResourceHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			data, err := ioutil.ReadAll(req.Body)
			if req.Body != nil {
				defer req.Body.Close()
			}
			if err != nil {
				return empty, 400, lxerrors.New("parsing submit resource request", err)
			}
			var resource lxtypes.Resource
			err = json.Unmarshal(data, &resource)
			if err != nil {
				return empty, 500, lxerrors.New("could not parse json to resource", err)
			}
			err = lx_core_helpers.SubmitResource(wrapper.state, &resource)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle Submit Resource request")
				return empty, 500, lxerrors.New("could not handle SubmitResource request", err)
			}
			lxlog.Infof(logrus.Fields{"resource": resource}, "accepted resource from rpi")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing Submit Resource request")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	submitStatusUpdateHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			data, err := ioutil.ReadAll(req.Body)
			if req.Body != nil {
				defer req.Body.Close()
			}
			if err != nil {
				return empty, 400, lxerrors.New("parsing submit status request", err)
			}
			var status mesosproto.TaskStatus
			err = proto.Unmarshal(data, &status)
			if err != nil {
				return empty, 500, lxerrors.New("could not parse protobuf to status", err)
			}
			err = lx_core_helpers.SubmitStatusUpdate(wrapper.state, tpiUrl, &status)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle Submit status request")
				return empty, 500, lxerrors.New("could not handle submit status request", err)
			}
			lxlog.Infof(logrus.Fields{"status": status}, "accepted status update from rpi")
			return empty, 202, nil
		}
		_, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing submit status request")
			driverErrc <- err
			return
		}
		res.WriteHeader(statusCode)
	}

	getNodesHandler := func(res http.ResponseWriter, req *http.Request) {
		fn := func() ([]byte, int, error) {
			nodes, err := lx_core_helpers.GetNodes(wrapper.state)
			if err != nil {
				lxlog.Errorf(logrus.Fields{
					"error": err,
				}, "could not handle Get Nodes request")
				return empty, 500, lxerrors.New("could not handle Get Nodes request", err)
			}
			data, err := json.Marshal(nodes)
			if err != nil {
				lxlog.Errorf(logrus.Fields{}, "could not marshal nodes to json")
				return empty, 500, lxerrors.New("marshalling nodes to json", err)
			}
			lxlog.Debugf(logrus.Fields{"nodes": nodes}, "Added new TaskProvider to LayerX")
			return data, 200, nil
		}
		data, statusCode, err := wrapper.queueOperation(fn)
		if err != nil {
			res.WriteHeader(statusCode)
			lxlog.Errorf(logrus.Fields{
				"error": err.Error(),
			}, "processing Get Nodes message")
			driverErrc <- err
			return
		}
		res.Write(data)
	}

	m.Post(RegisterTpi, registerTpiHandler)
	m.Post(RegisterRpi, registerRpiHandler)
	m.Post(RegisterTaskProvider, registerTaskProviderHandler)
	m.Post(DeregisterTaskProvider+"/:task_provider_id", deregisterTaskProviderHandler)
	m.Get(GetTaskProviders, getTaskProvidersHandler)
	m.Get(GetStatusUpdates+"/:task_provider_id", getStatusUpdatesHandler)
	m.Post(SubmitTask+"/:task_provider_id", submitTaskHandler)
	m.Post(KillTask+"/:task_id", killTaskHandler)
	m.Post(PurgeTask+"/:task_id", purgeTaskHandler)
	m.Post(SubmitResource, submitResourceHandler)
	m.Post(SubmitStatusUpdate, submitStatusUpdateHandler)
	m.Get(GetNodes, getNodesHandler)

	return m
}



func (wrapper *layerxCoreServerWrapper) queueOperation(f func() ([]byte, int, error)) ([]byte, int, error) {
	datac := make(chan []byte)
	statusCodec := make(chan int)
	errc := make(chan error)
	wrapper.actionQueue.Push(
		func() {
			data, statusCode, err := f()
			datac <- data
			statusCodec <- statusCode
			errc <- err
		})
	return <-datac, <-statusCodec, <-errc
}