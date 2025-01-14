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
	lbtypes "bk-bcs/bcs-common/pkg/loadbalance/v2"
)

//ExpServiceHandler handle for taskgroup
type ExpServiceHandler struct {
	oper      DataOperator
	dataType  string
	ClusterID string
}

func (handler *ExpServiceHandler) GetType() string {
	return handler.dataType
}

func (handler *ExpServiceHandler) CheckDirty() error {

	// CANNOT DELETE NOW, BECAUSE WE DONNOT SYNC every 180s!!!!!

	return nil
}

//Add handler to add
func (handler *ExpServiceHandler) Add(data interface{}) error {
	dataType := data.(*lbtypes.ExportService)
	blog.V(3).Infof("ExportService %s-%s.%s handle add Event", handler.ClusterID, dataType.Namespace, dataType.ServiceName)

	dataNode := "/bcsstorage/v1/mesos/watch/clusters/" + handler.ClusterID + "/namespaces/" + dataType.Namespace + "/" + handler.dataType + "/" + dataType.ServiceName
	handler.oper.CreateDCNode(dataNode, data, "PUT")

	dataNode2 := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.Namespace + "/" + handler.dataType + "/" + dataType.ServiceName
	handler.oper.CreateDCNode(dataNode2, data, "PUT")

	return nil
}

//Delete delete info
func (handler *ExpServiceHandler) Delete(data interface{}) error {
	dataType := data.(*lbtypes.ExportService)
	blog.V(3).Infof("ExportService %s-%s.%s handle delete Event", handler.ClusterID, dataType.Namespace, dataType.ServiceName)

	dataNode := "/bcsstorage/v1/mesos/watch/clusters/" + handler.ClusterID + "/namespaces/" + dataType.Namespace + "/" + handler.dataType + "/" + dataType.ServiceName
	handler.oper.DeleteDCNode(dataNode, "DELETE")

	dataNode2 := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.Namespace + "/" + handler.dataType + "/" + dataType.ServiceName
	handler.oper.DeleteDCNode(dataNode2, "DELETE")

	return nil
}

//Update update in zookeeper
func (handler *ExpServiceHandler) Update(data interface{}) error {
	dataType := data.(*lbtypes.ExportService)
	blog.V(3).Infof("ExportService %s-%s.%s handle update event", handler.ClusterID, dataType.Namespace, dataType.ServiceName)

	dataNode := "/bcsstorage/v1/mesos/watch/clusters/" + handler.ClusterID + "/namespaces/" + dataType.Namespace + "/" + handler.dataType + "/" + dataType.ServiceName
	handler.oper.CreateDCNode(dataNode, data, "PUT")

	dataNode2 := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.Namespace + "/" + handler.dataType + "/" + dataType.ServiceName
	handler.oper.CreateDCNode(dataNode2, data, "PUT")

	return nil
}
