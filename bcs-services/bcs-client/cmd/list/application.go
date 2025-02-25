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

package list

import (
	"fmt"
	"net/url"
	"sort"

	commonTypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-client/cmd/utils"
	"bk-bcs/bcs-services/bcs-client/pkg/storage/v1"
)

func listApplication(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace); err != nil {
		return err
	}

	storage := v1.NewBcsStorage(utils.GetClientOption())

	// get namespace
	condition := url.Values{}
	condition.Add(FilterNamespaceTag, c.Namespace())

	if c.IsAllNamespace() {
		var err error
		if condition, err = getNamespaceFilter(storage, c.ClusterID()); err != nil {
			return err
		}
	}

	list, err := storage.ListApplication(c.ClusterID(), condition)
	if err != nil {
		return fmt.Errorf("failed to list application: %v", err)
	}

	sort.Sort(list)
	return printListApplication(list)
}

func printListApplication(list v1.ApplicationList) error {
	if len(list) == 0 {
		fmt.Printf("Found no application\n")
		return nil
	}

	fmt.Printf("%-50s %-10s %-10s %-17s %-25s %-35s %-30s\n",
		"NAME",
		"STATUS",
		"INSTANCE",
		"RUNNING_INSTANCE",
		"MESSAGE",
		"CREATETIME",
		"NAMESPACE")
	for _, status := range list {
		if status.Data.Kind != "" && status.Data.Kind != commonTypes.BcsDataType_APP {
			continue
		}
		if len(status.Data.Message) > 22 {
			status.Data.Message = status.Data.Message[:22] + "..."
		}
		fmt.Printf("%-50s %-10s %-10d %-17d %-25s %-35s %-30s\n",
			status.Data.Name,
			status.Data.Status,
			status.Data.Instance,
			status.Data.RunningInstance,
			status.Data.Message,
			status.Data.CreateTime,
			status.Data.ObjectMeta.NameSpace)
	}
	return nil
}
