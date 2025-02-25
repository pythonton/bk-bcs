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

package storage

import (
	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	schedtypes "bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"fmt"
	"time"
)

//AppHandler handle for Application
type AppHandler struct {
	oper         DataOperator
	dataType     string
	ClusterID    string
	DoCheckDirty bool
}

func (handler *AppHandler) GetType() string {
	return handler.dataType
}

func (handler *AppHandler) CheckDirty() error {
	if handler.DoCheckDirty {
		blog.Info("check dirty data for type: %s", handler.dataType)
	} else {
		return nil
	}

	conditionData := &commtypes.BcsStorageDynamicBatchDeleteIf{
		UpdateTimeBegin: 0,
		UpdateTimeEnd:   time.Now().Unix() - 600,
	}

	dataNode := fmt.Sprintf("/bcsstorage/v1/mesos/dynamic/all_resources/clusters/%s/%s",
		handler.ClusterID, handler.dataType)
	err := handler.oper.DeleteDCNodes(dataNode, conditionData, "DELETE")
	if err != nil {
		blog.Error("delete timeover node(%s) failed: %+v", dataNode, err)
		return err
	}

	return nil
}

//Add add event
func (handler *AppHandler) Add(data interface{}) error {
	dataType := data.(*schedtypes.Application)
	blog.Info("App add event, AppID: %s.%s", dataType.RunAs, dataType.ID)
	reportType, _ := handler.FormatConv(dataType)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.RunAs + "/" + handler.dataType + "/" + dataType.ID
	err := handler.oper.CreateDCNode(dataNode, reportType, "PUT")
	if err != nil {
		blog.Error("App add node(%s) failed: %+v", dataNode, err)
		return err
	}
	return nil
}

//Delete delete info
func (handler *AppHandler) Delete(data interface{}) error {
	dataType := data.(*schedtypes.Application)
	blog.Info("App delete event, AppID: %s.%s", dataType.RunAs, dataType.ID)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.RunAs + "/" + handler.dataType + "/" + dataType.ID
	err := handler.oper.DeleteDCNode(dataNode, "DELETE")
	if err != nil {
		blog.Error("App delete node(%s) failed: %+v", dataNode, err)
		return err
	}

	return err
}

//Update update in zookeeper
func (handler *AppHandler) Update(data interface{}) error {
	dataType := data.(*schedtypes.Application)
	blog.V(3).Infof("App update event, AppID: %s.%s", dataType.RunAs, dataType.ID)
	reportType, _ := handler.FormatConv(dataType)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.RunAs + "/" + handler.dataType + "/" + dataType.ID
	err := handler.oper.CreateDCNode(dataNode, reportType, "PUT")
	if err != nil {
		blog.Error("App update node(%s) failed: %+v", dataNode, err)
		return err
	}

	return nil
}

func (handler *AppHandler) FormatConv(app *schedtypes.Application) (*commtypes.BcsReplicaControllerStatus, error) {
	status := new(commtypes.BcsReplicaControllerStatus)
	status.ObjectMeta = app.ObjectMeta

	status.ObjectMeta.Name = app.ID

	status.Instance = int(app.DefineInstances)
	status.BuildedInstance = int(app.Instances)
	status.RunningInstance = int(app.RunningInstances)
	status.CreateTime = time.Unix(app.Created, 0)
	// should be changed
	status.LastUpdateTime = time.Unix(app.UpdateTime, 0)
	status.ReportTime = time.Now()
	status.Status = commtypes.ReplicaControllerStatus(app.Status)
	status.LastStatus = commtypes.ReplicaControllerStatus(app.LastStatus)

	status.Message = app.Message
	status.Pods = app.Pods
	status.Kind = app.Kind

	blog.V(3).Infof("before post to CC, application format convert to: %+v", status)

	return status, nil
}
