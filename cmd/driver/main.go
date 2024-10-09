/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package main runs the main function
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/provider"
	"github.com/huawei/cosi-driver/pkg/utils/log"
	"github.com/huawei/cosi-driver/pkg/utils/version"
)

var (
	driverAddress = flag.String("driver-address", "var/lib/cosi/cosi.sock",
		"driver address for socket")
	driverName = flag.String("driver-name", "cosi.huawei.com",
		"name for cosi driver, e.g. cosi.huawei.com")
	logFile = flag.String("log-file", "huawei-cosi-driver",
		"The log file name of the xuanwu volume provider")
	kubeConfigPath = flag.String("kube-config-path", "", "absolute path to the kubeConfig file")
)

const (
	containerName = "cosi-driver"

	endpointDirPermission = 0755
)

func init() {
	flag.Parse()
}

func main() {
	err := log.InitLogging(*logFile)
	if err != nil {
		logrus.Errorf("init log failed, error is [%v]", err)
		return
	}

	ctx, err := log.SetRequestInfo(context.Background())
	if err != nil {
		log.Errorf("set request info failed, error is [%v]", err)
		return
	}

	err = version.RegisterVersion(containerName, version.COSIDriverVersion, *kubeConfigPath)
	if err != nil {
		log.AddContext(ctx).Errorf("init version file failed, error is [%v]", err)
		return
	}

	err = createEndpointDir(*driverAddress)
	if err != nil {
		log.AddContext(ctx).Errorf("create unix endpoint dir [%s] failed, error is [%v]", *driverAddress, err)
		return
	}

	// get unix endpoint listener
	listener, err := net.Listen("unix", *driverAddress)
	if err != nil {
		log.AddContext(ctx).Errorf("unix listen on [%s] failed, error is [%v]", *driverAddress, err)
		return
	}
	log.AddContext(ctx).Infof("unix listen on [%s] successfully!", *driverAddress)

	svc, err := initDriverSvc(ctx, *driverName, *kubeConfigPath)
	if err != nil {
		log.AddContext(ctx).Errorf("init cosi driver service failed, error is [%v]", err)
		return
	}

	// run blocking call in a separate goroutine, report errors via channel
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGILL, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		if err = svc.Serve(listener); err != nil {
			log.AddContext(ctx).Errorf("cosi driver service start failed, error is [%v]", err)
			signalChan <- syscall.SIGINT
		}
	}()
	// terminate grpc server gracefully before leaving main function
	defer func() {
		svc.GracefulStop()
	}()
	log.AddContext(ctx).Infoln("start cosi driver service successfully!")

	stopSignal := <-signalChan
	log.AddContext(ctx).Warningf("stop cosi driver service successfully, stopSignal is [%v]", stopSignal)
}

func initDriverSvc(ctx context.Context, driverName, kubeConfigPath string) (*grpc.Server, error) {
	identityServer, provisionerServer, err := provider.NewDriver(ctx, driverName, kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("new driver failed, error is [%v]", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(log.EnsureGRPCContext),
	}
	grpcServer := grpc.NewServer(opts...)
	cosispec.RegisterIdentityServer(grpcServer, identityServer)
	cosispec.RegisterProvisionerServer(grpcServer, provisionerServer)

	return grpcServer, nil
}

func createEndpointDir(endpoint string) error {
	endpointDir := filepath.Dir(endpoint)
	_, err := os.Stat(endpointDir)
	if err != nil {
		// endpoint dir not exist, create dir and return,
		// no need to check endpoint.
		if os.IsNotExist(err) {
			err = os.MkdirAll(endpointDir, endpointDirPermission)
			if err != nil {
				return fmt.Errorf("create endpoint directory [%s] failed, error is [%v]", endpointDir, err)
			}

			return nil
		}

		return fmt.Errorf("os stat endpoint dir [%s] failed, error is [%v]", endpointDir, err)
	}

	// if endpoint dir exist, then delete endpoint anyway
	err = os.Remove(endpoint)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("remove old endpoint [%s] failed, error is [%v]", endpoint, err)
	}

	return nil
}
