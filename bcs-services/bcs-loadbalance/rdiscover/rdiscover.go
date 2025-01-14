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

package rdiscover

import (
	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"

	"golang.org/x/net/context"
)

var (
	_, classA, _  = net.ParseCIDR("10.0.0.0/8")
	_, classA2, _ = net.ParseCIDR("9.0.0.0/8")
	_, classAa, _ = net.ParseCIDR("100.64.0.0/10")
	_, classB, _  = net.ParseCIDR("172.16.0.0/12")
	_, classC, _  = net.ParseCIDR("192.168.0.0/16")
)

//RDiscover register and discover
type RDiscover struct {
	isMaster   bool
	zkSubPath  string
	rd         *RegisterDiscover.RegDiscover
	clusterid  string
	proxy      string
	metricPort uint
	rootCxt    context.Context
	cancel     context.CancelFunc

	bcsLBServs []*types.LoadBalanceInfo
	bcsLBLock  sync.RWMutex
}

//NewRDiscover new a register discover object to register zookeeper
func NewRDiscover(zkserv, subPath, clusterid, proxy string, metricPort uint) *RDiscover {
	return &RDiscover{
		zkSubPath:  subPath,
		rd:         RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		clusterid:  clusterid,
		proxy:      proxy,
		metricPort: metricPort,
	}
}

//Start the register and discover
func (r *RDiscover) Start() error {
	//create root context
	r.rootCxt, r.cancel = context.WithCancel(context.Background())

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Errorf("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	//register loadbalance to bcs zk
	if err := r.registerLoadBalance(); err != nil {
		blog.Errorf("fail to register err:%s", err.Error())
		return err
	}

	var bcsLoadbalanceEvent <-chan *RegisterDiscover.DiscoverEvent

	bcsLBPath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_LOADBALANCE + "/" + r.zkSubPath
	bcsLoadbalanceEvent, err := r.rd.DiscoverService(bcsLBPath)
	if err != nil {
		blog.Errorf("fail to register discover for bcs loadbalance. err:%s", err.Error())
		return err
	}
	blog.Infof("register bcs lb path %s discover success", bcsLBPath)

	//here: discover other bcs services
	go r.CheckMasterStatus()
	for {
		select {
		case bcsLBEnv := <-bcsLoadbalanceEvent:
			err = r.discoverBCSLBServ(bcsLBEnv.Server)
			if err != nil {
				blog.Warnf("discover bcs lb %v, err %s", bcsLBEnv.Server, err.Error())
			}
		case <-r.rootCxt.Done():
			blog.Warn("register and discover serv done")
			return nil
		}
	}
}

//Stop the register and discover
func (r *RDiscover) Stop() error {
	r.cancel()
	if err := r.rd.Stop(); err != nil {
		return fmt.Errorf("register discover stop failed, err %s", err.Error())
	}
	return nil
}

//GetAvailableIP get local host ip address
func GetAvailableIP() string {

	ifName := os.Getenv("LB_NETWORKCARD")
	if len(ifName) == 0 {
		ifName = "eth1"
	}
	netIf, err := net.InterfaceByName(ifName)
	if err != nil {
		return ""
	}
	addrs, err := netIf.Addrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			if classA.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classA2.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classAa.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classB.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classC.Contains(ip.IP) {
				return ip.IP.String()
			}
		}
	}

	return ""
}

//CheckMasterStatus timer to check master status
func (r *RDiscover) CheckMasterStatus() {
	for {
		r.bcsLBLock.Lock()
		if len(r.bcsLBServs) > 0 {
			if r.bcsLBServs[0].IP == GetAvailableIP() {
				if !r.isMaster {
					blog.Infof("#######Status chanaged, Self node become Master#############")
				}
				r.isMaster = true
			} else {
				r.isMaster = false
			}
		}
		r.bcsLBLock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

//IsMaster return a bool to indicate whether i am a master
func (r *RDiscover) IsMaster() bool {
	return r.isMaster
}

//GetBCSLBServList get bcs lb as awselb/awsclb target
func (r *RDiscover) GetBCSLBServList() []types.LoadBalanceInfo {
	r.bcsLBLock.Lock()
	defer r.bcsLBLock.Unlock()
	var rst []types.LoadBalanceInfo
	for _, value := range r.bcsLBServs {
		rst = append(rst, *value)
	}
	return rst
}

func (r *RDiscover) discoverBCSLBServ(servInfos []string) error {
	blog.Debug(fmt.Sprintf("discover loadbalance(%v)", servInfos))

	lbs := []*types.LoadBalanceInfo{}
	for _, serv := range servInfos {
		lb := new(types.LoadBalanceInfo)
		if err := json.Unmarshal([]byte(serv), lb); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		lbs = append(lbs, lb)
	}

	r.bcsLBLock.Lock()
	defer r.bcsLBLock.Unlock()
	r.bcsLBServs = lbs
	return nil
}

func (r *RDiscover) registerLoadBalance() error {
	lbInfo := new(types.LoadBalanceInfo)
	//need to judge getInnerIP() succeed or not, because must go ahead
	lbInfo.IP = GetAvailableIP()
	//metric port, let healthz check
	lbInfo.Port = r.metricPort
	lbInfo.MetricPort = r.metricPort
	//include http must, may be include tcp ,https also
	lbInfo.Scheme = "http"
	lbInfo.Version = version.GetVersion()
	lbInfo.Pid = os.Getpid()
	lbInfo.Cluster = r.clusterid

	data, err := json.Marshal(lbInfo)
	if err != nil {
		blog.Errorf("fail to marshal loadbalance info to json. err:%s", err.Error())
		return err
	}
	path := r.getLBZookeeperParentPath() + "/" + lbInfo.IP

	return r.rd.RegisterAndWatchService(path, data)
}

func (r *RDiscover) getLBZookeeperParentPath() string {
	return types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_LOADBALANCE + "/" + r.zkSubPath
}
