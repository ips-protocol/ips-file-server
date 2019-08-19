package aliyun

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/mts"
	"github.com/ipweb-group/file-server/config"
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

func VideoSnapShot(input, output string) {
	client := GetMTSClient()
	c := config.GetConfig().Aliyun

	snapshotJob := mts.CreateSubmitSnapshotJobRequest()
	snapshotJob.Input = fmt.Sprintf(`{"Bucket":"%s", "Location": "%s","Object":"%s" }`, c.Bucket, c.OssLocation, input)
	snapshotJob.SnapshotConfig = fmt.Sprintf(`{"OutputFile": {"Bucket": "%s","Location":"%s","Object": "%s"},"Time": "5"}`, c.Bucket, c.OssLocation, output)

	resp, err := client.SubmitSnapshotJob(snapshotJob)
	if err != nil {
		log.Fatal(err)
	}

	// TODO 记录转换 ID 到数据库
	fmt.Println(resp)
}

func VideoCovert(input, output string) {
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
		log.Fatal(err)
	}

	// TODO 记录转换 ID 到数据库
	fmt.Println(resp)
}
