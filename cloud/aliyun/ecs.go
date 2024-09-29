package aliyun

import (
	"errors"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	sdkClient "github.com/alibabacloud-go/ecs-20140526/v2/client"
	sdkService "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gogf/gf/v2/encoding/gjson"
)

type AppService struct {
	Ak, Sk string
}

func (a *AppService) Client() (*sdkClient.Client, error) {
	config := &openapi.Config{AccessKeyId: tea.String(a.Ak), AccessKeySecret: tea.String(a.Sk)}
	config.Endpoint = tea.String("ecs-cn-hangzhou.aliyuncs.com")
	return sdkClient.NewClient(config)
}

func (a *AppService) GetRegionEcsList(region, pageToken string, pageSize int) (string, error) {
	if pageSize < 10 || pageSize > 100 {
		return "", errors.New("pageSize < 10 || pageSize > 100")
	}
	client, err := a.Client()
	if err != nil {
		return "", err
	}
	describeInstancesRequest := &sdkClient.DescribeInstancesRequest{
		MaxResults: tea.Int32(int32(pageSize)),
		RegionId:   tea.String(region),
		NextToken:  tea.String(pageToken),
	}
	runtime := &sdkService.RuntimeOptions{ConnectTimeout: tea.Int(50000), ReadTimeout: tea.Int(50000)}
	data, err := client.DescribeInstancesWithOptions(describeInstancesRequest, runtime)
	if err != nil {
		return "", err
	}
	return gjson.New(data.Body.String()).String(), nil
}

func (a *AppService) GetRegionEcsListPage(region string, pageSize int) ([]string, error) {
	if pageSize < 10 || pageSize > 100 {
		return nil, errors.New("pageSize < 10 || pageSize > 100")
	}
	var ecsPagesList []string
	pageData, err := a.GetRegionEcsList(region, "", pageSize)
	if err != nil {
		return ecsPagesList, err
	}
	pageDataJson := gjson.New(pageData)
	instances := pageDataJson.Get("Instances.Instance").Strings()
	nextToken := pageDataJson.Get("NextToken").String()
	ecsPagesList = append(ecsPagesList, instances...)
	for {
		if nextToken == "" {
			break
		}
		pageDataLoop, errLoop := a.GetRegionEcsList(region, nextToken, pageSize)
		if errLoop != nil {
			return ecsPagesList, errLoop
		}
		pageDataJsonLoop := gjson.New(pageDataLoop)
		instancesLoop := pageDataJsonLoop.Get("Instances.Instance").Strings()
		nextToken = pageDataJsonLoop.Get("NextToken").String()
		ecsPagesList = append(ecsPagesList, instancesLoop...)
	}
	return ecsPagesList, nil
}
