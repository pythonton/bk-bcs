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

package types

import (
	"strconv"

	"bk-bcs/bcs-mesos/bcs-scheduler/src/mesosproto/mesos"
	mesos_master "bk-bcs/bcs-mesos/bcs-scheduler/src/mesosproto/mesos/master"

	commtypes "bk-bcs/bcs-common/common/types"
	//"fmt"
)

//executor or task default resources limits
const (
	CPUS_PER_EXECUTOR = 0.01
	CPUS_PER_TASK     = 1
	MEM_PER_EXECUTOR  = 64
	MEM_PER_TASK      = 64
	DISK_PER_EXECUTOR = 64
	DISK_PER_TASK     = 64
)

// operation operate
const (
	OPERATION_LAUNCH     = "LAUNCH"
	OPERATION_DELETE     = "DELETE"
	OPERATION_SCALE      = "SCALE"
	OPERATION_INNERSCALE = "INNERSCALE"
	OPERATION_ROLLBACK   = "ROLLBACK"
	OPERATION_RESCHEDULE = "RESCHEDULE"
	OPERATION_UPDATE     = "UPDATE"
)

//operation status
const (
	OPERATION_STATUS_INIT    = "INIT"
	OPERATION_STATUS_FINISH  = "FINISH"
	OPERATION_STATUS_FAIL    = "FAIL"
	OPERATION_STATUS_TIMEOUT = "TIMEOUT"
)

// extension for TaskState_TASK_...
const (
	Ext_TaskState_TASK_RESTARTING int32 = 101
)

//app status
const (
	APP_STATUS_STAGING       = "Staging"
	APP_STATUS_DEPLOYING     = "Deploying"
	APP_STATUS_RUNNING       = "Running"
	APP_STATUS_FINISH        = "Finish"
	APP_STATUS_ERROR         = "Error"
	APP_STATUS_OPERATING     = "Operating"
	APP_STATUS_ROLLINGUPDATE = "RollingUpdate"
	APP_STATUS_UNKNOWN       = "Unknown"
	APP_STATUS_ABNORMAL      = "Abnormal"
)

//app substatus
const (
	APP_SUBSTATUS_UNKNOWN            = "Unknown"
	APP_SUBSTATUS_ROLLINGUPDATE_DOWN = "RollingUpdateDown"
	APP_SUBSTATUS_ROLLINGUPDATE_UP   = "RollingUpdateUp"
)

//task status
const (
	TASK_STATUS_STAGING  = "Staging"
	TASK_STATUS_STARTING = "Starting"
	TASK_STATUS_RUNNING  = "Running"
	TASK_STATUS_FINISH   = "Finish"
	TASK_STATUS_ERROR    = "Error"
	TASK_STATUS_KILLING  = "Killing"
	TASK_STATUS_KILLED   = "Killed"
	TASK_STATUS_FAIL     = "Failed"
	TASK_STATUS_LOST     = "Lost"

	TASK_STATUS_RESTARTING = "Restarting"

	TASK_STATUS_UNKOWN = "Unknown"
)

//taskgroup status
const (
	TASKGROUP_STATUS_STAGING  = "Staging"
	TASKGROUP_STATUS_STARTING = "Starting"
	TASKGROUP_STATUS_RUNNING  = "Running"
	TASKGROUP_STATUS_FINISH   = "Finish"
	TASKGROUP_STATUS_ERROR    = "Error"
	TASKGROUP_STATUS_KILLING  = "Killing"
	TASKGROUP_STATUS_KILLED   = "Killed"
	TASKGROUP_STATUS_FAIL     = "Failed"
	TASKGROUP_STATUS_LOST     = "Lost"

	TASKGROUP_STATUS_RESTARTING = "Restarting"

	TASKGROUP_STATUS_UNKNOWN = "Unknown"
)

const (
	TASK_TEMPLATE_KEY_FORMAT      = "${%s}"
	TASK_TEMPLATE_KEY_PORT_FORMAT = "ports.%s"
	TASK_TEMPLATE_KEY_PROCESSNAME = "processname"
	TASK_TEMPLATE_KEY_INSTANCEID  = "instanceid"
	TASK_TEMPLATE_KEY_HOSTIP      = "hostip"
	TASK_TEMPLATE_KEY_NAMESPACE   = "namespace"
	TASK_TEMPLATE_KEY_WORKPATH    = "workPath"
	TASK_TEMPLATE_KEY_PIDFILE     = "pidFile"
)

const (
	APP_TASK_TEMPLATE_KEY_FORMAT      = "${%s}"
	APP_TASK_TEMPLATE_KEY_PORT_FORMAT = "bcs.ports.%s"
	APP_TASK_TEMPLATE_KEY_APPNAME     = "bcs.appname"
	APP_TASK_TEMPLATE_KEY_INSTANCEID  = "bcs.instanceid"
	APP_TASK_TEMPLATE_KEY_HOSTIP      = "bcs.hostip"
	APP_TASK_TEMPLATE_KEY_NAMESPACE   = "bcs.namespace"
	APP_TASK_TEMPLATE_KEY_PODID       = "bcs.taskgroupid"
	APP_TASK_TEMPLATE_KEY_PODNAME     = "bcs.taskgroupname"
)

//Version for api resources application or deployment
type Version struct {
	ID            string
	Name          string
	ObjectMeta    commtypes.ObjectMeta
	PodObjectMeta commtypes.ObjectMeta
	Instances     int32
	RunAs         string
	Container     []*Container
	//add  20180802
	Process       []*commtypes.Process
	Labels        map[string]string
	KillPolicy    *commtypes.KillPolicy
	RestartPolicy *commtypes.RestartPolicy
	Constraints   *commtypes.Constraint
	Uris          []string
	Ip            []string
	Mode          string
	// added  20181011, add for differentiate process/application
	Kind commtypes.BcsDataType
	// add  20181122
	RawJson *commtypes.ReplicaController `json:"raw_json,omitempty"`
}

//Resource discribe resources needed by a task
type Resource struct {
	//cpu核数
	Cpus   float64
	CPUSet int16
	//MB
	Mem  float64
	Disk float64
	//IOTps  uint32 //default times per second
	//IOBps  uint32 //default MB/s
}

//CheckAndDefaultResource check the resource of each container, if no resource, set default value
func (version *Version) CheckAndDefaultResource() error {
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			if process.Resources.Limits.Cpu == "" {
				process.Resources.Limits.Cpu = strconv.Itoa(CPUS_PER_TASK)
			}
			if process.Resources.Limits.Mem == "" {
				process.Resources.Limits.Mem = strconv.Itoa(MEM_PER_TASK)
			}
			if process.Resources.Limits.Storage == "" {
				process.Resources.Limits.Storage = strconv.Itoa(DISK_PER_TASK)
			}
		}
		return nil
	case commtypes.BcsDataType_APP, "":
		for index, container := range version.Container {
			if nil == container.DataClass {
				version.Container[index].DataClass = &DataClass{}
			}
			if nil == container.DataClass.Resources {
				version.Container[index].DataClass.Resources = &Resource{
					Cpus: float64(CPUS_PER_TASK),
					Mem:  float64(MEM_PER_TASK),
					Disk: float64(DISK_PER_TASK),
				}
			}
		}
		return nil
	}

	return nil
}

//check application constraints whether is valid
func (version *Version) CheckConstraints() bool {
	if version.Constraints == nil {
		return true
	}

	for _, constraint := range version.Constraints.IntersectionItem {
		if constraint == nil {
			continue
		}
		for _, oneData := range constraint.UnionData {
			if oneData == nil {
				continue
			}
			if oneData.Type == commtypes.ConstValueType_Scalar && oneData.Scalar == nil {
				return false
			}
			if oneData.Type == commtypes.ConstValueType_Text && oneData.Text == nil {
				return false
			}
			if oneData.Type == commtypes.ConstValueType_Set && oneData.Set == nil {
				return false
			}
			if oneData.Type == commtypes.ConstValueType_Range {
				for _, oneRange := range oneData.Ranges {
					if oneRange == nil {
						return false
					}
				}
			}
		}
	}

	return true
}

//AllCpus return taskgroup will use cpu resources
func (version *Version) AllCpus() float64 {
	var allCpus float64
	allCpus = 0

	// split process and containers
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			cpu, _ := strconv.ParseFloat(process.Resources.Limits.Cpu, 64)
			allCpus = allCpus + cpu
		}
	case commtypes.BcsDataType_APP, "":
		for _, container := range version.Container {
			allCpus = allCpus + container.DataClass.Resources.Cpus
		}
	}
	return allCpus
}

//AllMems return taskgroup will use memory resource
func (version *Version) AllMems() float64 {
	var allMem float64
	allMem = 0

	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			mem, _ := strconv.ParseFloat(process.Resources.Limits.Mem, 64)
			allMem = allMem + mem
		}
	case commtypes.BcsDataType_APP, "":
		for _, container := range version.Container {
			allMem = allMem + container.DataClass.Resources.Mem
		}
	}
	return allMem + float64(MEM_PER_EXECUTOR)
}

//AllDisk return taskgroup will use disk resources
func (version *Version) AllDisk() float64 {
	var allDisk float64
	allDisk = 0

	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			disk, _ := strconv.ParseFloat(process.Resources.Limits.Storage, 64)
			allDisk = allDisk + disk
		}
	case commtypes.BcsDataType_APP, "":
		for _, container := range version.Container {
			allDisk = allDisk + container.DataClass.Resources.Disk
		}
	}
	return allDisk + float64(DISK_PER_EXECUTOR)
}

//AllResource return  taskgroup used cpu, memory, disk resources
func (version *Version) AllResource() *Resource {
	return &Resource{
		Cpus: version.AllCpus(),
		Mem:  version.AllMems(),
		Disk: version.AllDisk(),
	}
}

//Container for Version
type Container struct {
	Type          string
	Docker        *Docker
	Volumes       []*Volume
	Resources     *Resource
	LimitResoures *Resource
	DataClass     *DataClass

	ConfigMaps []commtypes.ConfigMap
	Secrets    []commtypes.Secret

	HealthChecks []*commtypes.HealthCheck

	//network flow limit
	NetLimit *commtypes.NetLimit
}

//Docker for container
type Docker struct {
	Hostname        string
	ForcePullImage  bool
	Image           string
	ImagePullUser   string
	ImagePullPasswd string
	Network         string
	NetworkType     string
	Command         string
	Arguments       []string
	Parameters      []*Parameter
	PortMappings    []*PortMapping
	Env             map[string]string
	Privileged      bool
}

//Parameter for container
type Parameter struct {
	Key   string
	Value string
}

//PortMapping for container
type PortMapping struct {
	ContainerPort int32
	HostPort      int32
	Name          string
	Protocol      string
}

//Volume for container
type Volume struct {
	ContainerPath string
	HostPath      string
	Mode          string
}

//HealthCheck
//type HealthCheck struct {
//	ID                     string
//	Address                string
//	TaskID                 string
//	AppID                  string
//	Protocol               string
//	Port                   int32
//	PortIndex              int32
//	PortName               string
//	Command                *Command
//	Path                   string
//	MaxConsecutiveFailures uint32
//	GracePeriodSeconds     float64
//	IntervalSeconds        float64
//	TimeoutSeconds         float64
//	DelaySeconds           float64
//	ConsecutiveFailures    uint32
//}

//Command
type Command struct {
	Value string
}

//Task
type Task struct {
	Kind            commtypes.BcsDataType
	ID              string
	Name            string
	Hostame         string
	Command         string
	Arguments       []string
	Image           string
	ImagePullUser   string
	ImagePullPasswd string
	Network         string
	NetworkType     string
	PortMappings    []*PortMapping
	Privileged      bool
	Parameters      []*Parameter
	ForcePullImage  bool
	Volumes         []*Volume
	Env             map[string]string
	Labels          map[string]string
	DataClass       *DataClass
	HealthChecks    []*commtypes.HealthCheck
	// health check status
	HealthCheckStatus           []*commtypes.BcsHealthCheckStatus
	Healthy                     bool
	IsChecked                   bool
	ConsecutiveFailureTimes     uint32
	LocalMaxConsecutiveFailures uint32

	OfferId        string
	AgentId        string
	AgentHostname  string
	AgentIPAddress string
	Status         string
	LastStatus     string
	UpdateTime     int64
	StatusData     string
	AppId          string
	RunAs          string
	KillPolicy     *commtypes.KillPolicy
	Uris           []string
	LastUpdateTime int64
	Message        string
	//network flow limit
	NetLimit *commtypes.NetLimit
}

// taskgroup describes the implements of multiple tasks
type TaskGroup struct {
	Kind            commtypes.BcsDataType
	ID              string
	Name            string
	AppID           string
	RunAs           string
	ObjectMeta      commtypes.ObjectMeta
	AgentID         string
	ExecutorID      string
	Status          string
	LastStatus      string
	InstanceID      uint64
	Taskgroup       []*Task
	KillPolicy      *commtypes.KillPolicy
	RestartPolicy   *commtypes.RestartPolicy
	VersionName     string
	LastUpdateTime  int64
	Attributes      []*mesos.Attribute
	StartTime       int64
	UpdateTime      int64
	ReschededTimes  int
	LastReschedTime int64
	//we should replace the next three BcsXXX, using ObjectMeta.Labels directly
	//BcsAppID       string
	//BcsSetID       string
	//BcsModuleID    string
	HostName       string
	Message        string
	LaunchResource *Resource
	CurrResource   *Resource
	//BcsMessages map[int64]*BcsMessage
	BcsEventMsg *BcsMessage
}

//Application for container
type Application struct {
	Kind             commtypes.BcsDataType
	ID               string
	Name             string
	ObjectMeta       commtypes.ObjectMeta
	DefineInstances  uint64
	Instances        uint64
	RunningInstances uint64
	RunAs            string
	ClusterId        string
	Status           string
	SubStatus        string
	LastStatus       string
	Created          int64
	UpdateTime       int64
	Mode             string
	LastUpdateTime   int64
	//we should replace the next three BcsXXX, using ObjectMeta.Labels directly
	//BcsAppID    string
	//BcsSetID    string
	//BcsModuleID string
	Message string
	Pods    []*commtypes.BcsPodIndex
	// add  20181122
	RawJson *commtypes.ReplicaController `json:"raw_json,omitempty"`
}

//Operation for application
type Operation struct {
	ID             string
	RunAs          string
	AppID          string
	OperationType  string
	Status         string
	CreateTime     int64
	LastUpdateTime int64
	ErrorStr       string
}

type OperationIndex struct {
	Operation string
}

// mesos slave info
type Agent struct {
	Key          string
	LastSyncTime int64
	AgentInfo    *mesos_master.Response_GetAgents_Agent
}

type Check struct {
	ID          string   `json:"id"`
	Protocol    string   `json:"protocol"`
	Address     string   `json:"address"`
	Port        int      `json:"port"`
	Command     *Command `json:"command"`
	Path        string   `json:"path"`
	MaxFailures int      `json:"max_failures"`
	Interval    int      `json:"interval"`
	Timeout     int      `json:"timeout"`
	TaskID      string   `json:"task_id"`
	TaskGroupID string   `json:"taskgroup_id"`
	AppID       string   `json:"app_id"`
}

type ProcDef struct {
	ProcName   string           `json:"procName"`
	WorkPath   string           `json:"workPath"`
	PidFile    string           `json:"pidFile"`
	StartCmd   string           `json:"startCmd"`
	CheckCmd   string           `json:"checkCmd"`
	StopCmd    string           `json:"stopCmd"`
	RestartCmd string           `json:"restartCmd"`
	ReloadCmd  string           `json:"reloadCmd"`
	KillCmd    string           `json:"killCmd"`
	LogPath    string           `json:"logPath"`
	CfgPath    string           `json:"cfgPath"`
	Uris       []*commtypes.Uri `json:"uris"`
	// seconds
	StartGracePeriod int `json:"startGracePeriod"`
}

type DataClass struct {
	Resources      *Resource
	LimitResources *Resource
	Msgs           []*BcsMessage
	NetLimit       *commtypes.NetLimit
	//add for proc 20180730
	ProcInfo *ProcDef
}

type DeploymentDef struct {
	ObjectMeta commtypes.ObjectMeta      `json:"metadata"`
	Selector   map[string]string         `json:"selector,omitempty"`
	Version    *Version                  `json:"version"`
	Strategy   commtypes.UpgradeStrategy `json:"strategy"`
	// add  20181122
	RawJson *commtypes.BcsDeployment `json:"raw_json,omitempty"`
}

const (
	DEPLOYMENT_STATUS_DEPLOYING             = "Deploying"
	DEPLOYMENT_STATUS_RUNNING               = "Running"
	DEPLOYMENT_STATUS_ROLLINGUPDATE         = "Update"
	DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED  = "UpdatePaused"
	DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND = "UpdateSuspend"
	DEPLOYMENT_STATUS_DELETING              = "Deleting"
)

const (
	DEPLOYMENT_OPERATION_NIL    = ""
	DEPLOYMENT_OPERATION_DELETE = "DELETE"
	DEPLOYMENT_OPERATION_START  = "START"
)

type Deployment struct {
	ObjectMeta      commtypes.ObjectMeta        `json:"metadata"`
	Selector        map[string]string           `json:"selector,omitempty"`
	Strategy        commtypes.UpgradeStrategy   `json:"strategy"`
	Status          string                      `json:"status"`
	Application     *DeploymentReferApplication `json:"application"`
	ApplicationExt  *DeploymentReferApplication `json:"application_ext"`
	LastRollingTime int64                       `json:"last_rolling_time"`
	CurrRollingOp   string                      `json:"curr_rolling_operation"`
	IsInRolling     bool                        `json:"is_in_rolling"`
	CheckTime       int64                       `json:"check_time"`
	Message         string                      `json:"message"`
	// add  20181122
	RawJson       *commtypes.BcsDeployment `json:"raw_json,omitempty"`
	RawJsonBackup *commtypes.BcsDeployment `json:"raw_json_backup,omitempty"`
}

type DeploymentReferApplication struct {
	ApplicationName         string `json:"name"`
	CurrentTargetInstances  int    `json:"curr_target_instances"`
	CurrentRollingInstances int    `josn:"curr_rolling_instances"`
}

type AgentSchedInfo struct {
	HostName   string  `json:"host_name"`
	DeltaCPU   float64 `json:"delta_cpu"`
	DeltaMem   float64 `json:"delta_mem"`
	DeltaDisk  float64 `json:"delta_disk"`
	Taskgroups map[string]*Resource
}

type TaskGroupOpResult struct {
	ID     string
	Status string
	Err    string
}
