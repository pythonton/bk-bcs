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

package resthdrs

import (
	"fmt"

	m "bk-bcs/bcs-services/bcs-api/pkg/models"
	"bk-bcs/bcs-services/bcs-api/pkg/server/proxier"
	"bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/filters"
	"bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/utils"
	"bk-bcs/bcs-services/bcs-api/pkg/server/types"
	"bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-api/pkg/server/external-cluster/tke"
	"github.com/emicklei/go-restful"
	"time"
)

type UpdateCredentialsForm struct {
	RegisterToken   string `json:"register_token" validate:"required"`
	ServerAddresses string `json:"server_addresses" validate:"required,apiserver_addresses"`
	CaCertData      string `json:"cacert_data" validate:"required"`
	UserToken       string `json:"user_token" validate:"required"`
}

// UpdateCredentials updates the current cluster credentials, a valid registerToken is required to performe
// a credentials updating.
func UpdateCredentials(request *restful.Request, response *restful.Response) {
	form := UpdateCredentialsForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	clusterId := request.PathParameter("cluster_id")

	// FIXME: validate the registerToken here?

	// validate if the registerToken is correct
	token := sqlstore.GetRegisterToken(clusterId)
	if token == nil {
		message := fmt.Sprintf("errcode: %d, no valid register token found for cluster", common.BcsErrApiBadRequest)
		WriteClientError(response, "RTOKEN_NOT_FOUND", message)
		return
	}
	if token.Token != form.RegisterToken {
		message := fmt.Sprintf("errcode: %d, invalid register token given", common.BcsErrApiBadRequest)
		WriteClientError(response, "INVALID_RTOKEN", message)
		return
	}

	// validate if CaCertData or UserToken changed, if changed, refresh handler of clusterId
	clusterCredentials := sqlstore.GetCredentials(clusterId)
	if clusterCredentials != nil {
		if clusterCredentials.CaCertData != form.CaCertData || clusterCredentials.UserToken != form.UserToken {
			proxier.DefaultReverseProxyDispatcher.DelHandlerStoreByClusterId(clusterId)
		}
	}

	err = sqlstore.SaveCredentials(clusterId, form.ServerAddresses, form.CaCertData, form.UserToken, "")
	if err != nil {
		message := fmt.Sprintf("errcode: %d, can not update credentials, error: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "CANNOT_UPDATE_CREDENTIALS", message)
		return
	}
	response.WriteEntity(types.EmptyResponse{})
}

// GetCredentials lists the credentials for current user
func GetCredentials(request *restful.Request, response *restful.Response) {
	cluster := filters.GetCluster(request)

	credentials := sqlstore.GetCredentials(cluster.ID)
	if credentials == nil {
		message := fmt.Sprintf("errcode: %d, no credentials found", common.BcsErrApiBadRequest)
		WriteClientError(response, "CREDENTIALS_NOT_FOUND", message)
		return
	}
	response.WriteEntity(credentials)
}

// ClientCredentials 包含 kubectl 等客户端访问一个集群的必要信息
type ClientCredentials struct {
	ClusterID         string `json:"cluster_id"`
	ServerAddressPath string `json:"server_address_path"`
	UserToken         string `json:"user_token"`
	CaCertData        string `json:"cacert_data"`
}

// GetClientCredentials list the credentials, clients like kubectl can use this credentials to connect to cluster
func GetClientCredentials(request *restful.Request, response *restful.Response) {
	user := filters.GetUser(request)
	cluster := filters.GetCluster(request)
	userTokenType := filters.GetUserTokenType(request)

	credentials := sqlstore.GetCredentials(cluster.ID)
	if credentials == nil {
		message := fmt.Sprintf("errcode: %d, no credentials found", common.BcsErrApiBadRequest)
		WriteClientError(response, "CREDENTIALS_NOT_FOUND", message)
		return
	}
	// Create a user token if not exists
	userToken, err := sqlstore.GetOrCreateUserToken(user, uint(userTokenType), "")
	if err != nil {
		blog.Warnf("Unable to create user token of type UserTokenTypeKubeConfig for user %s: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, can not create user token: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteServerError(response, "CANNOT_CREATE_USER_RTOKEN", message)
		return
	}
	clientCredential := ClientCredentials{
		ClusterID:         cluster.ID,
		ServerAddressPath: fmt.Sprintf("/tunnels/clusters/%s/", cluster.Identifier),
		UserToken:         userToken.Value,
		CaCertData:        credentials.CaCertData,
	}
	response.WriteEntity(clientCredential)
}

// SyncTkeClusterCredentials sync the tke cluster credentials from tke
func SyncTkeClusterCredentials(request *restful.Request, response *restful.Response) {
	cluster := filters.GetCluster(request)

	externalClusterInfo := sqlstore.QueryBCSClusterInfo(&m.BCSClusterInfo{
		ClusterId: cluster.ID,
	})
	if externalClusterInfo == nil {
		message := fmt.Sprintf("errcode: %d, external cluster info not exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "EXTERNAL_CLUSTER_NOT_EXISTS", message)
		return
	}
	if externalClusterInfo.ClusterType != utils.BcsTkeCluster {
		message := fmt.Sprintf("errcode: %d, cluster %s is not tke cluster", common.BcsErrApiBadRequest, cluster.ID)
		WriteClientError(response, "NOT_TKE_CLUSTER", message)
		return
	}

	tkeCluster := tke.NewTkeCluster(cluster.ID, externalClusterInfo.TkeClusterId, externalClusterInfo.TkeClusterRegion)

	err := tkeCluster.BindClusterLb()
	if err != nil {
		message := err.Error()
		WriteServerError(response, "CANNOT_BIND_TKE_CLUSTER_LB", message)
		return
	}

	var isBinded bool
	// wait for BindMasterVipLoadBalancer to be successful
	for i := 0; i < 6; i++ {
		time.Sleep(15 * time.Second)
		isBinded, err = tkeCluster.GetMasterVip()
		if err != nil {
			message := err.Error()
			WriteServerError(response, "GET_TKE_MASTER_VIP_FAILED", message)
			return
		}
		if isBinded {
			break
		}
	}

	if !isBinded {
		message := fmt.Sprintf("failed to bind lb for tke cluster: %s", cluster.ID)
		WriteServerError(response, "BIND_TKE_CLUSTER_LB_FAILED", message)
		return
	}

	err = tkeCluster.SyncClusterCredentials()
	if err != nil {
		message := err.Error()
		WriteServerError(response, "CANNOT_SYNC_TKE_CREDENTIALS", message)
		return
	}
	response.WriteEntity(types.EmptyResponse{})
}
