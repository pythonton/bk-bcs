/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/blog/glog"
	"bk-bcs/bcs-common/common/license"
	"bk-bcs/bcs-common/common/metric"
	commtype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-api/config"
	"bk-bcs/bcs-services/bcs-api/options"
	"bk-bcs/bcs-services/bcs-api/processor"
	//"bk-bcs/bcs-services/bcs-api/regdiscv"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	// skip tls handshake error log for tencent tgw tcp check
	if strings.HasPrefix(string(data), "http: TLS handshake error from") {
		return len(data), nil
	}
	glog.Info(string(data))
	return len(data), nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := options.NewServerOption()
	if err := options.Parse(op); err != nil {
		fmt.Printf("parse options failed: %v\n", err)
		os.Exit(1)
	}

	blog.InitLogs(op.LogConfig)
	// to adapt to tencent tgw，give a temporary method to skip tls handshake error log. see https://github.com/Tencent/bk-bcs/issues/32
	log.SetOutput(GlogWriter{})
	defer blog.CloseLogs()

	blog.Info("init config success")
	license.CheckLicense(op.LicenseServerConfig)

	//run apiserver
	run(op)

	ch := make(chan int)
	<-ch
}

//Run the apiserver
func run(op *options.ServerOption) {

	conf := parseConfig(op)

	//run register and discover
	//regdiscv.RunRDiscover(conf.RegDiscvSrv, conf)

	proc := processor.NewProcessor(conf)
	//start processor, and http & websokect service
	err := proc.Start()
	if err != nil {
		blog.Errorf("start processor error %s, and exit", err.Error())
		os.Exit(1)
	}

	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	runMetric(conf, nil)

	return
}

func parseConfig(op *options.ServerOption) *config.ApiServConfig {
	apiServConfig := config.NewApiServConfig()

	apiServConfig.Address = op.Address
	apiServConfig.Port = op.Port
	apiServConfig.InsecureAddress = op.InsecureAddress
	apiServConfig.InsecurePort = op.InsecurePort
	apiServConfig.RegDiscvSrv = op.BCSZk
	apiServConfig.LocalIp = op.LocalIP
	apiServConfig.MetricPort = op.MetricPort
	apiServConfig.BKIamAuth = op.BKIamAuth
	apiServConfig.BKE = op.BKE
	apiServConfig.TKE = op.TKE
	apiServConfig.Edition = op.Edition
	apiServConfig.MesosWebconsoleProxyPort = op.MesosWebconsoleProxyPort
	config.Edition = apiServConfig.Edition
	config.BKIamAuth = apiServConfig.BKIamAuth
	config.TurnOnRBAC = apiServConfig.BKE.TurnOnRBAC
	config.ClusterCredentialsFixtures = apiServConfig.BKE.ClusterCredentialsFixtures
	config.MesosWebconsoleProxyPort = apiServConfig.MesosWebconsoleProxyPort
	config.TkeConf = op.TKE

	//server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		apiServConfig.ServCert.CertFile = op.CertConfig.ServerCertFile
		apiServConfig.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		apiServConfig.ServCert.CAFile = op.CertConfig.CAFile
		apiServConfig.ServCert.IsSSL = true
	}

	//client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.ClientKeyFile != "" {
		apiServConfig.ClientCert.CertFile = op.CertConfig.ClientCertFile
		apiServConfig.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		apiServConfig.ClientCert.CAFile = op.CertConfig.CAFile
		apiServConfig.ClientCert.IsSSL = true
	}

	// custom cert config
	if op.CustomCertConfig.ServerCertFile != "" && op.CustomCertConfig.ServerKeyFile != "" {
		apiServConfig.ServCert.CertFile = op.CustomCertConfig.ServerCertFile
		apiServConfig.ServCert.KeyFile = op.CustomCertConfig.ServerKeyFile
		apiServConfig.ServCert.CAFile = op.CustomCertConfig.CAFile
		apiServConfig.ServCert.CertPasswd = op.CustomCertConfig.ServerKeyPwd
		apiServConfig.ServCert.IsSSL = true
	}
	if op.CustomCertConfig.ClientCertFile != "" && op.CustomCertConfig.ClientKeyFile != "" {
		apiServConfig.ClientCert.CertFile = op.CustomCertConfig.ClientCertFile
		apiServConfig.ClientCert.KeyFile = op.CustomCertConfig.ClientKeyFile
		apiServConfig.ClientCert.CAFile = op.CustomCertConfig.CAFile
		apiServConfig.ClientCert.CertPasswd = op.CustomCertConfig.ClientKeyPwd
		apiServConfig.ClientCert.IsSSL = true
	}

	apiServConfig.VerifyClientTLS = op.VerifyClientTLS

	return apiServConfig
}

func runMetric(conf *config.ApiServConfig, err error) {

	blog.Infof("run metric: port(%d)", conf.MetricPort)

	metricConf := metric.Config{
		RunMode:     metric.Master_Master_Mode,
		ModuleName:  commtype.BCS_MODULE_APISERVER,
		MetricPort:  conf.MetricPort,
		IP:          conf.LocalIp,
		SvrCaFile:   conf.ServCert.CAFile,
		SvrCertFile: conf.ServCert.CertFile,
		SvrKeyFile:  conf.ServCert.KeyFile,
		SvrKeyPwd:   conf.ServCert.CertPasswd,
	}

	healthFunc := func() metric.HealthMeta {
		var isHealthy bool
		var msg string
		if err == nil {
			isHealthy = true
		} else {
			msg = err.Error()
		}

		return metric.HealthMeta{
			CurrentRole: metric.MasterRole,
			IsHealthy:   isHealthy,
			Message:     msg,
		}
	}

	if err := metric.NewMetricController(
		metricConf,
		healthFunc); nil != err {
		blog.Errorf("run metric fail: %s", err.Error())
	}

	blog.Infof("run metric ok")
}
