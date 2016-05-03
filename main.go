package main

import (
	"flag"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"github.com/layer-x/layerx-commons/lxdatabase"
	"github.com/layer-x/layerx-commons/lxmartini"
	"github.com/layer-x/layerx-core_v2/health_checker"
	"github.com/layer-x/layerx-core_v2/lxserver"
	"github.com/layer-x/layerx-core_v2/lxstate"
	"github.com/layer-x/layerx-core_v2/main_loop"
	"github.com/layer-x/layerx-core_v2/task_launcher"
)

func purgeState() error {
	return lxdatabase.Rmdir("/state", true)
}

func main() {
	portPtr := flag.Int("port", 6666, "port to run core on")
	etcdUrlPtr := flag.String("etcd", "127.0.0.1:4001", "url of etcd cluster")
	purgePtr := flag.Bool("purge", false, "purge ETCD state")
	debugPtr := flag.Bool("debug", false, "Run Layer-X in debug mode")
	flag.Parse()

	if *debugPtr {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithFields(logrus.Fields{
		"port": *portPtr,
		"etcd": *etcdUrlPtr,
	}).Infof("Booting Layer-X Core...")

	state := lxstate.NewState()
	err := state.InitializeState("http://" + *etcdUrlPtr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"etcd": *etcdUrlPtr,
		}).Fatalf("Failed to initialize Layer-X State")
	}
	if *purgePtr {
		err = purgeState()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"etcd": *etcdUrlPtr,
			}).Fatalf("Failed to purge Layer-X State")
		}
		err = state.InitializeState("http://" + *etcdUrlPtr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"etcd": *etcdUrlPtr,
			}).Fatalf("Failed to initialize Layer-X State")
		}
	}

	logrus.WithFields(logrus.Fields{
		"port": *portPtr,
		"etcd": *etcdUrlPtr,
	}).Infof("Layer-X Core Initialized. Waiting for registration of TPI and RPI...")

	driverErrc := make(chan error)
	mainServer := lxmartini.QuietMartini()
	coreServerWrapper := lxserver.NewLayerXCoreServerWrapper(state, mainServer, driverErrc)

	mainServer = coreServerWrapper.WrapServer()

	mainServer.Use(martini.Static("web"))

	go mainServer.RunOnAddr(fmt.Sprintf(":%v", *portPtr))

	clearRpisAndResources(state)

	taskLauncher := task_launcher.NewTaskLauncher(state)
	healthChecker := health_checker.NewHealthChecker(state)
	go main_loop.MainLoop(taskLauncher, healthChecker, state, driverErrc)
	logrus.WithFields(logrus.Fields{}).Infof("Layer-X Server initialized successfully.")

	for {
		err = <-driverErrc
		if err != nil {
			logrus.WithError(err).Errorf("Layer-X Core had an error!")
		}
	}
}

func clearRpisAndResources(state *lxstate.State) {
	//clear previous rpis
	oldRpis, _ := state.RpiPool.GetRpis()
	for _, rpi := range oldRpis {
		state.RpiPool.DeleteRpi(rpi.Name)
	}
	//clear previous resources
	oldNodes, _ := state.NodePool.GetNodes()
	for _, node := range oldNodes {
		nodeResourcePool, err := state.NodePool.GetNodeResourcePool(node.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err, "node": node}).Warnf("retreiving resource pool for node")
			continue
		}
		resources, _ := nodeResourcePool.GetResources()
		for _, resource := range resources {
			nodeResourcePool.DeleteResource(resource.Id)
		}
	}
}
