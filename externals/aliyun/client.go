package aliyun

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/mts"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/utils"
	"log"
)

var mtsClient *mts.Client

func GetMTSClient() *mts.Client {
	if mtsClient == nil {
		c := config.GetConfig().Aliyun
		_mtsClient, err := mts.NewClientWithAccessKey(c.Region, c.AccessKey, c.AccessSecret)
		if err != nil {
			log.Fatal(err)
		}
		mtsClient = _mtsClient
	}

	return mtsClient
}

func VideoSnapShot(input, output string, time int32) (err error) {
	lg := utils.GetLogger()
	client := GetMTSClient()
	c := config.GetConfig().Aliyun

	snapshotJob := mts.CreateSubmitSnapshotJobRequest()
	snapshotJob.Input = fmt.Sprintf(`{"Bucket":"%s", "Location": "%s","Object":"%s" }`, c.Bucket, c.OssLocation, input)
	snapshotJob.SnapshotConfig = fmt.Sprintf(`{"OutputFile": {"Bucket": "%s","Location":"%s","Object": "%s"},"Time": "%d"}`, c.Bucket, c.OssLocation, output, time)

	_, err = client.SubmitSnapshotJob(snapshotJob)
	if err != nil {
		lg.Error("Submit snapshot job error,", err)
		return
	}

	lg.Infof("Video snapshot job submitted, snapshot is at time %d ms", time)

	return
}

func VideoCovert(input, output string) (jobId string, err error) {
	client := GetMTSClient()
	c := config.GetConfig().Aliyun

	job := mts.CreateSubmitJobsRequest()
	job.Input = fmt.Sprintf(`{"Bucket":"%s", "Location": "%s","Object":"%s" }`, c.Bucket, c.OssLocation, input)
	job.OutputBucket = c.Bucket
	job.OutputLocation = c.OssLocation
	job.Outputs = fmt.Sprintf(`[{"OutputObject": "%s","TemplateId": "%s"}]`, output, c.MTSConvertTemplateId)
	job.PipelineId = c.MTSPipelineID

	resp, err := client.SubmitJobs(job)
	if err != nil {
		lg := utils.GetLogger()
		lg.Error("Submit video convert job error,", err)
		return
	}

	jobId = resp.JobResultList.JobResult[0].Job.JobId
	return
}
