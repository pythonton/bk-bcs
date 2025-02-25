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

package sqlstore

import (
	m "bk-bcs/bcs-services/bcs-api/pkg/models"
)

func SaveTkeLbSubnet(clusterRegion, subnetId string) error {
	var tkeLbSubnet m.TkeLbSubnet
	dbScoped := GCoreDB.Where(m.TkeLbSubnet{ClusterRegion: clusterRegion}).Assign(
		m.TkeLbSubnet{
			SubnetId: subnetId,
		},
	).FirstOrCreate(&tkeLbSubnet)

	return dbScoped.Error
}

func GetSubnetByClusterRegion(clusterRegion string) *m.TkeLbSubnet {
	tkeLbSubnet := m.TkeLbSubnet{}
	GCoreDB.Where(&m.TkeLbSubnet{ClusterRegion: clusterRegion}).First(&tkeLbSubnet)
	if tkeLbSubnet.ID != 0 {
		return &tkeLbSubnet
	}
	return nil
}
