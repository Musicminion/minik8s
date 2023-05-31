package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// 获取单个的Job
// @Summary 获取单个的Job
// "/apis/v1/namespaces/:namespace/jobs/:name"
func GetJob(c *gin.Context) {

	// 解析里面的参数
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := "GetJob: namespace=" + namespace + ", name=" + name
	k8log.InfoLog("APIServer", logStr)

	// 完整路径：/registry/jobs/<namespace>/<job-name>
	key := fmt.Sprintf(serverconfig.EtcdJobPath+"%s/%s", namespace, name)

	// 从etcd中获取数据
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "get job err, not find job",
		})
		return
	}

	// 处理Res，如果有多个返回的，报错
	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job err, find more than one job",
		})
		return
	}

	targetJob := res[0].Value
	c.JSON(http.StatusOK, gin.H{
		"data": targetJob,
	})
}

// "/apis/v1/namespaces/:namespace/jobs"
func GetJobs(c *gin.Context) {
	// TODO
	namespace := c.Param(config.URL_PARAM_NAMESPACE)

	if namespace == "" {
		// c.JSON(http.StatusBadRequest, gin.H{
		// 	"error": "namespace is empty",
		// })
		// return
		namespace = config.DefaultNamespace
	}

	// 完整路径：/registry/jobs/<namespace>
	// 从etcd中获取数据
	logStr := "GetJobs: namespace=" + namespace
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdJobPath+"%s", namespace)
	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get jobs failed " + err.Error(),
		})
		return
	}

	// 处理Res，如果有多个返回的，报错
	targetJobs := make([]string, 0)

	for _, pod := range res {
		targetJobs = append(targetJobs, pod.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetJobs),
	})
}

// 创建Job
// "/apis/v1/namespaces/:namespace/jobs"
func AddJob(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddJob")

	// 从body中获取Job的信息
	var job apiObject.Job

	if err := c.ShouldBind(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parse job err, " + err.Error(),
		})

		k8log.ErrorLog("APIServer", "parse job err, "+err.Error())
		return
	}

	// 检查Job的合法性
	newJobName := job.Metadata.Name
	if newJobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job name is empty",
		})

		k8log.ErrorLog("APIServer", "job name is empty")
		return
	}

	// 检查Job的namespace是否存在
	if job.GetObjectNamespace() == "" {
		job.Metadata.Namespace = config.DefaultNamespace
	}

	// 检查Job是否存在
	key := fmt.Sprintf(serverconfig.EtcdJobPath+"%s/%s", job.GetObjectNamespace(), newJobName)

	// 从etcd中获取数据
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job err, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job already exist",
		})
		k8log.ErrorLog("APIServer", "job already exist")
		return
	}

	// 给Job设置UUID，用于后续的操作
	// 哪怕用户自己设置了UUID，也会被覆盖
	job.Metadata.UUID = uuid.NewUUID()

	// 将Job转化为JobStore
	jobStore := job.ToJobStore()
	randomSuffix := "-" + stringutil.GenerateRandomStr(6)
	// 在文件名的文件类型前加入随机字符串，防止文件名重复
	re := regexp.MustCompile(`(.*)(\.[^.]+)`)
	match := re.FindStringSubmatch(jobStore.Spec.OutputFile)

	if len(match) >= 3 {
		jobStore.Spec.OutputFile = match[1] + randomSuffix + match[2]
	} else {
		fmt.Println("无法解析文件名")
	}

	match = re.FindStringSubmatch(jobStore.Spec.ErrorFile)

	if len(match) >= 3 {
		jobStore.Spec.ErrorFile = match[1] + randomSuffix + match[2]
	} else {
		fmt.Println("无法解析文件名")
	}

	// 将JobStore转化为json
	jobStoreJson, err := json.Marshal(jobStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal job err, " + err.Error(),
		})
		return
	}

	// 将JobStore存入etcd
	key = fmt.Sprintf(serverconfig.EtcdJobPath+"%s/%s", job.GetObjectNamespace(), newJobName)

	// 将JobStore存入etcd
	err = etcdclient.EtcdStore.Put(key, jobStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "put job err, " + err.Error(),
		})
		return
	}

	// 返回
	c.JSON(http.StatusCreated, gin.H{
		"message": "create job success",
	})

	k8log.InfoLog("APIServer", "create job success")

	// 后面可能要发消息给controller，
	/*
		// 发送消息给controller，然后controller会去etcd中获取job的信息
		// 之后会根据job的信息创建pod，然后执行任务
	*/
}

// "/apis/v1/namespaces/:namespace/jobs/:name"
func DeleteJob(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "DeleteJob")

	// 获取namespace和jobName
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	// 检查参数是否为空
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := "DeleteJob: namespace=" + namespace + ", name=" + name
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdJobPath+"%s/%s", namespace, name)
	// 删除Job
	err := etcdclient.EtcdStore.Del(key)

	if err != nil {
		k8log.DebugLog("APIServer", "delete job err, "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "delete job err, " + err.Error(),
		})
		return
	}

	// 返回
	c.JSON(http.StatusNoContent, gin.H{
		"message": "delete job success",
	})

	k8log.InfoLog("APIServer", "delete job success")

	// 后面要不要垃圾回收？
	/*
		【TODO】
	*/
}

// 获取Job的状态
// "/apis/v1/namespaces/:namespace/jobs/:name/status"
func GetJobStatus(c *gin.Context) {
	jobNamespace := c.Param(config.URL_PARAM_NAMESPACE)
	jobName := c.Param(config.URL_PARAM_NAME)

	// 检查参数是否为空
	if jobNamespace == "" || jobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取Job
	logStr := "GetJobStatus: namespace=" + jobNamespace + ", name=" + jobName
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdJobPath+"%s/%s", jobNamespace, jobName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job err, " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "job not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job err, job is not unique",
		})
		return
	}

	// 获取res[0]的JobStore
	jobStore := &apiObject.JobStore{}
	err = json.Unmarshal([]byte(res[0].Value), jobStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unmarshal job err, " + err.Error(),
		})
		return
	}

	// 返回jobStore的Status
	c.JSON(http.StatusOK, gin.H{
		"data": jobStore.Status,
	})

}

// 更新Job的状态
func UpdateJobStatus(c *gin.Context) {
	jobName := c.Param(config.URL_PARAM_NAME)
	jobNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	// 检查参数是否为空
	if jobNamespace == "" || jobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := "UpdateJobStatus: namespace=" + jobNamespace + ", name=" + jobName
	k8log.InfoLog("APIServer", logStr)

	// 从etcd中获取Job
	key := fmt.Sprintf(serverconfig.EtcdJobPath+"%s/%s", jobNamespace, jobName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job err, " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "job not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job err, job is not unique",
		})
		return
	}

	// 获取res[0]的JobStore
	jobStore := &apiObject.JobStore{}
	err = json.Unmarshal([]byte(res[0].Value), jobStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unmarshal job err, " + err.Error(),
		})
		return
	}

	// 获取status
	jobStatus := &apiObject.JobStatus{}

	err = c.ShouldBind(jobStatus)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bind job status err, " + err.Error(),
		})
		return
	}

	// 更新jobStore的Status
	selectiveUpdateJobStatus(jobStore, jobStatus)

	// 将jobStore转化为json
	jobStoreJson, err := json.Marshal(jobStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal job err, " + err.Error(),
		})
		return
	}

	// 更新etcd中的job
	err = etcdclient.EtcdStore.Put(key, jobStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "update job err, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update job status success",
	})
}

func selectiveUpdateJobStatus(oldJob *apiObject.JobStore, newJob *apiObject.JobStatus) {
	if newJob.Account != "" {
		oldJob.Status.Account = newJob.Account
	}

	if newJob.AllocCPUS != "" {
		oldJob.Status.AllocCPUS = newJob.AllocCPUS
	}

	if len(newJob.Error) != 0 {
		oldJob.Status.Error = newJob.Error
	}

	if newJob.ExitCode != "" {
		oldJob.Status.ExitCode = newJob.ExitCode
	}

	if newJob.JobID != "" {
		oldJob.Status.JobID = newJob.JobID
	}

	if len(newJob.Output) != 0 {
		oldJob.Status.Output = newJob.Output
	}

	if newJob.Partition != "" {
		oldJob.Status.Partition = newJob.Partition
	}

	if newJob.State != "" {
		oldJob.Status.State = newJob.State
	}

	oldJob.Status.UpdateTime = time.Now()
}

// "/apis/v1/namespaces/:namespace/jobs/:name/file"
func GetJobFile(c *gin.Context) {
	jobName := c.Param(config.URL_PARAM_NAME)
	jobNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	// 检查参数是否为空
	if jobNamespace == "" || jobName == "" {
		k8log.ErrorLog("APIServer", "GetJobFile: namespace or name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取Job
	logStr := "GetJobFile: namespace=" + jobNamespace + ", name=" + jobName
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdJobFilePath+"%s/%s", jobNamespace, jobName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job file err, " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "job file not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job file err, job file is not unique",
		})
		return
	}

	// 获取res[0]的JobFile
	targetFile := res[0].Value

	c.JSON(http.StatusOK, gin.H{
		"data": targetFile,
	})
}

// 一定要注意选择性的更新JobFile
// JobFile包含了Job的Error、Output、用户上传的原文件
// "/apis/v1/namespaces/:namespace/jobfiles"
func AddJobFile(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddJobFile")

	var jobFile apiObject.JobFile

	if err := c.ShouldBind(&jobFile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bind job file err, " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "bind job file err, "+err.Error())
		return
	}

	fmt.Println(jobFile)
	// 打印请求体c
	body, _ := io.ReadAll(c.Request.Body)
	k8log.InfoLog("APIServer", "request body: "+string(body))

	k8log.InfoLog("APIServer", "api version: "+jobFile.APIVersion)
	// 检查参数是否为空
	newJobName := jobFile.GetJobName()

	if newJobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job name is empty",
		})
		k8log.ErrorLog("APIServer", "job name is empty")
		return
	}

	if jobFile.GetJobNamespace() == "" {
		jobFile.Basic.Metadata.Namespace = config.DefaultNamespace
	}

	// 检查Jobfile是否存在
	key := fmt.Sprintf(serverconfig.EtcdJobFilePath+"%s/%s", jobFile.GetJobNamespace(), newJobName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job file err, " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "get job file err, "+err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job file already exists",
		})
		k8log.ErrorLog("APIServer", "job file already exists")
		return
	}

	// 设置uuid
	jobFile.Metadata.UUID = uuid.NewUUID()

	// 将jobFile转化为json
	jobFileJson, err := json.Marshal(jobFile)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal job file err, " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "marshal job file err, "+err.Error())
		return
	}

	key = fmt.Sprintf(serverconfig.EtcdJobFilePath+"%s/%s", jobFile.GetJobNamespace(), newJobName)
	// 将jobFile存入etcd
	err = etcdclient.EtcdStore.Put(key, jobFileJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "put job file err, " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "put job file err, "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "add job file success",
	})
	/*
		首次创建了JobFile，需要给消息队列发送消息
	*/
	message.PublishUpdateJobFile(&jobFile.Basic)

}

// "/apis/v1/namespaces/:namespace/jobfiles/:name"
func UpdateJobFile(c *gin.Context) {
	jobName := c.Param(config.URL_PARAM_NAME)
	jobNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	// 检查参数是否为空
	if jobNamespace == "" || jobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取Job
	logStr := "UpdateJobFile: namespace=" + jobNamespace + ", name=" + jobName

	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdJobFilePath+"%s/%s", jobNamespace, jobName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job file err, " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "job file not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get job file err, job file is not unique",
		})
		return
	}

	// 获取res[0]的JobFile
	oldJobFile := &apiObject.JobFile{}
	err = json.Unmarshal([]byte(res[0].Value), oldJobFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unmarshal job file err, " + err.Error(),
		})
		return
	}

	// 解析请求体里面的JobFile
	reqJobFile := &apiObject.JobFile{}
	err = c.ShouldBind(reqJobFile)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bind job file err, " + err.Error(),
		})
		return
	}

	// 选择性的只更新UserUploadFile、OutputFile、ErrorFile
	selectiveUpdateJobFile(oldJobFile, reqJobFile)

	// 将jobFile转化为json
	jobFileJson, err := json.Marshal(oldJobFile)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal job file err, " + err.Error(),
		})
		return
	}

	key = fmt.Sprintf(serverconfig.EtcdJobFilePath+"%s/%s", jobNamespace, jobName)

	err = etcdclient.EtcdStore.Put(key, jobFileJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "put job file err, " + err.Error(),
		})
		return
	}

	// 返回成功
	c.JSON(http.StatusOK, gin.H{
		"message": "update job file success",
	})

}

// 为了简单处理，只更新 UserUploadFile、OutputFile、ErrorFile
func selectiveUpdateJobFile(oldJobFile *apiObject.JobFile, newJobFile *apiObject.JobFile) {
	if len(newJobFile.UserUploadFile) != 0 {
		oldJobFile.UserUploadFile = newJobFile.UserUploadFile
	}

	if len(newJobFile.OutputFile) != 0 {
		oldJobFile.OutputFile = newJobFile.OutputFile
	}

	if len(newJobFile.ErrorFile) != 0 {
		oldJobFile.ErrorFile = newJobFile.ErrorFile
	}
}
