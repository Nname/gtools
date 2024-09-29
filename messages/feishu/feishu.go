package feishu

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"strings"
)

type feiShuAdapter struct{}

var adapter = feiShuAdapter{}

func Adapter() *feiShuAdapter {
	return &adapter
}

// FeiShuBotMsg msgType(text, post, image, share_chat) content `json`
func (s *feiShuAdapter) FeiShuBotMsg(url, msgType, content string) (string, error) {
	data := gjson.New(g.Map{"msg_type": msgType, "content": gjson.New(content)}).String()
	post, err := g.Client().Post(context.Background(), url, data)
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}

// FeiShuBotMsgCard card `json`
func (s *feiShuAdapter) FeiShuBotMsgCard(url string, card string) (string, error) {
	data := gjson.New(g.Map{"msg_type": "interactive", "card": gjson.New(card)}).String()
	post, err := g.Client().Post(context.Background(), url, data)
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}

type AppService struct {
	AppId     string
	AppSecret string
}

func (s *AppService) AppGetTenantToken() (string, error) {
	data := gjson.New(g.Map{"app_id": s.AppId, "app_secret": s.AppSecret}).String()
	post, err := g.Client().Header(map[string]string{"Content-Type": "application/json; charset=utf-8"}).Post(
		context.Background(),
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		data,
	)
	if err != nil {
		return "", err
	}
	defer post.Close()
	resultJson := gjson.New(post.ReadAllString())
	resultJson.MapStrAny()
	token := resultJson.Get("tenant_access_token").String()
	if token == "" {
		return "", errors.New(fmt.Sprintf("tenant_access_token not exist, %s", post.ReadAllString()))
	}
	return token, nil
}

func (s *AppService) AppGetAccessToken() (string, error) {
	data := gjson.New(g.Map{"app_id": s.AppId, "app_secret": s.AppSecret}).String()
	post, err := g.Client().Header(map[string]string{"Content-Type": "application/json; charset=utf-8"}).Post(
		context.Background(),
		"https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal",
		data,
	)
	if err != nil {
		return "", err
	}
	defer post.Close()
	resultJson := gjson.New(post.ReadAllString())
	resultJson.MapStrAny()
	token := resultJson.Get("app_access_token").String()
	if token == "" {
		return "", errors.New(fmt.Sprintf("app_access_token not exist, %s", post.ReadAllString()))
	}
	return token, nil
}

// AppClient AppGetTenantToken
func (s *AppService) AppClient() (*gclient.Client, error) {
	token, err := s.AppGetTenantToken()
	if err != nil {
		return nil, err
	}
	client := g.Client().Header(map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	})
	return client, nil
}

// AppClientAccessToken AppClientAccessToken
func (s *AppService) AppClientAccessToken() (*gclient.Client, error) {
	token, err := s.AppGetAccessToken()
	if err != nil {
		return nil, err
	}
	client := g.Client().Header(map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	})
	return client, nil
}

// AppClientTenantClient AppClientTenantClient
func (s *AppService) AppClientTenantClient() (*gclient.Client, error) {
	token, err := s.AppGetTenantToken()
	if err != nil {
		return nil, err
	}
	client := g.Client().Header(map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	})
	return client, nil
}

func (s *AppService) AppGetUserDetail(user string) (string, error) {
	return "", nil
}

func (s *AppService) AppGetUid(emails, mobiles []string) (string, error) {
	url := "https://open.feishu.cn/open-apis/contact/v3/users/batch_get_id?user_id_type=user_id"
	data := gjson.New(g.Map{"emails": emails, "mobiles": mobiles})
	appClient, err := s.AppClient()
	if err != nil {
		return "", err
	}
	post, err := appClient.Post(context.Background(), url, data.String())
	if err != nil {
		return "", err
	}
	defer post.Close()
	resultJson := gjson.New(post.ReadAllString())
	userList := resultJson.Get("data.user_list").Array()
	if len(userList) == 0 {
		return "", errors.New(fmt.Sprintf("args error, %s", resultJson.String()))
	}
	return resultJson.Get("data.user_list").String(), nil
}

func (s *AppService) AppGetOpenId(emails, mobiles []string) (string, error) {
	url := "https://open.feishu.cn/open-apis/contact/v3/users/batch_get_id?user_id_type=open_id"
	data := gjson.New(g.Map{"emails": emails, "mobiles": mobiles})
	appClient, err := s.AppClient()
	if err != nil {
		return "", err
	}
	post, err := appClient.Post(context.Background(), url, data.String())
	if err != nil {
		return "", err
	}
	defer post.Close()
	resultJson := gjson.New(post.ReadAllString())
	userList := resultJson.Get("data.user_list").Array()
	if len(userList) == 0 {
		return "", errors.New(resultJson.String())
	}
	return resultJson.Get("data.user_list").String(), nil
}

func (s *AppService) AppGetUidPersonFormMobile(mobile string) (string, error) {
	uidData, err := s.AppGetUid([]string{}, []string{mobile})
	if err != nil {
		return "", err
	}
	dataJson := gjson.New(uidData)
	dataJson.MapStrAny()
	return dataJson.Get("0.user_id").String(), nil
}

func (s *AppService) AppGetUidPersonFormEmail(email string) (string, error) {
	uidData, err := s.AppGetUid([]string{email}, []string{})
	if err != nil {
		return "", err
	}
	dataJson := gjson.New(uidData)
	dataJson.MapStrAny()
	return dataJson.Get("0.user_id").String(), nil
}

func (s *AppService) AppGetOpenIdPersonFormMobile(mobile string) (string, error) {
	uidData, err := s.AppGetOpenId([]string{}, []string{mobile})
	if err != nil {
		return "", err
	}
	dataJson := gjson.New(uidData)
	dataJson.MapStrAny()
	return dataJson.Get("0.user_id").String(), nil
}

func (s *AppService) AppGetOpenIdPersonFormEmail(email string) (string, error) {
	uidData, err := s.AppGetOpenId([]string{email}, []string{})
	if err != nil {
		return "", err
	}
	dataJson := gjson.New(uidData)
	dataJson.MapStrAny()
	return dataJson.Get("0.user_id").String(), nil
}

func (s *AppService) AppGetUidUser(user string) (string, error) {
	if strings.Contains(user, "@") {
		return s.AppGetUidPersonFormEmail(user)
	}
	return s.AppGetUidPersonFormMobile(user)
}

func (s *AppService) AppGetUidUsers(users []string) ([]string, error) {
	var userList []string
	for i := 0; i < len(users); i++ {
		if strings.Contains(users[i], "@") {
			email, err := s.AppGetUidPersonFormEmail(users[i])
			if err != nil {
				return nil, err
			}
			if email != "" {
				userList = append(userList, email)
			}
		} else {
			mobile, err := s.AppGetUidPersonFormMobile(users[i])
			if err != nil {
				return nil, err
			}
			if mobile != "" {
				userList = append(userList, mobile)
			}
		}
	}
	return userList, nil
}

func (s *AppService) AppGetOpenIdUser(user string) (string, error) {
	if strings.Contains(user, "@") {
		return s.AppGetOpenIdPersonFormEmail(user)
	}
	return s.AppGetOpenIdPersonFormMobile(user)
}

func (s *AppService) AppGetOpenIdUsers(users []string) ([]string, error) {
	var userList []string
	for i := 0; i < len(users); i++ {
		if strings.Contains(users[i], "@") {
			email, err := s.AppGetOpenIdPersonFormEmail(users[i])
			if err != nil {
				return nil, err
			}
			if email != "" {
				userList = append(userList, email)
			}
		} else {
			mobile, err := s.AppGetOpenIdPersonFormMobile(users[i])
			if err != nil {
				return nil, err
			}
			if mobile != "" {
				userList = append(userList, mobile)
			}
		}
	}
	return userList, nil
}

// AppSendMsgUser msgType(text, post, image, share_chat, ...) content `json` urgent(phone, sms, app)
func (s *AppService) AppSendMsgUser(msgType, user string, content, urgent string) (string, error) {
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	userId, err := s.AppGetOpenIdUser(user)
	if err != nil {
		return "", err
	}
	data := gjson.New(g.Map{"receive_id": userId, "msg_type": msgType, "content": gjson.New(content).String()}).String()
	post, err := client.Post(context.Background(), "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id", data)
	if err != nil {
		return "", err
	}
	defer post.Close()
	result := post.ReadAllString()
	msgId := gjson.New(result).Get("data.message_id").String()
	if strings.Contains(urgent, "phone") {
		return s.AppUrgentPhone(userId, msgId)
	} else if strings.Contains(urgent, "sms") {
		return s.AppUrgentSms(userId, msgId)
	} else if strings.Contains(urgent, "app") {
		return s.AppUrgentApp(userId, msgId)
	} else {
		return result, nil
	}
}

// AppSendMsgUsers msgType(text, post, image, share_chat, ...) content `json`
func (s *AppService) AppSendMsgUsers(user []string, msgType, content string) (string, error) {
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	userIdList, err := s.AppGetUidUsers(user)
	if err != nil {
		return "", err
	}
	data := gjson.New(g.Map{"user_ids": userIdList, "msg_type": msgType, "content": gjson.New(content)}).String()
	post, err := client.Post(context.Background(), "https://open.feishu.cn/open-apis/message/v4/batch_send/", data)
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}

// AppSendMsgUrgentUsers msgType(text, post, image, share_chat, ...) content `json` urgent(phone, sms, app)
func (s *AppService) AppSendMsgUrgentUsers(msgType string, user []string, content, urgent string) (string, error) {
	for i := 0; i < len(user); i++ {
		_, err := s.AppSendMsgUser(msgType, user[i], content, urgent)
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

// AppUrgentPhone user(open_id)
func (s *AppService) AppUrgentPhone(user string, messageId string) (string, error) {
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	data := gjson.New(g.Map{"user_id_list": []string{user}}).String()
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages/%s/urgent_phone?user_id_type=open_id", messageId)
	patch, err := client.Patch(context.Background(), url, data)
	if err != nil {
		return "", err
	}
	defer patch.Close()
	return patch.ReadAllString(), nil
}

// AppUrgentSms user(open_id)
func (s *AppService) AppUrgentSms(user string, messageId string) (string, error) {
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	data := gjson.New(g.Map{"user_id_list": []string{user}}).String()
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages/%s/urgent_sms?user_id_type=open_id", messageId)
	patch, err := client.Patch(context.Background(), url, data)
	if err != nil {
		return "", err
	}
	defer patch.Close()
	return patch.ReadAllString(), nil
}

// AppUrgentApp user(open_id)
func (s *AppService) AppUrgentApp(user string, messageId string) (string, error) {
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	data := gjson.New(g.Map{"user_id_list": []string{user}}).String()
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages/%s/urgent_app?user_id_type=open_id", messageId)
	patch, err := client.Patch(context.Background(), url, data)
	if err != nil {
		return "", err
	}
	defer patch.Close()
	return patch.ReadAllString(), nil
}

// AppSendMsgUsersCard card `json`
func (s *AppService) AppSendMsgUsersCard(user []string, card string) (string, error) {
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	userIdList, err := s.AppGetOpenIdUsers(user)
	if err != nil {
		return "", err
	}
	data := gjson.New(g.Map{"open_ids": userIdList, "msg_type": "interactive", "card": gjson.New(card)}).String()
	post, err := client.Post(context.Background(), "https://open.feishu.cn/open-apis/message/v4/batch_send/", data)
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}

// AppGetDepartment id DepartmentId(default 0)
func (s *AppService) AppGetDepartment(id string) (string, error) {
	if id == "" {
		id = "0"
	}
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/contact/v3/departments/%s/children?department_id_type=department_id&fetch_child=true&page_size=50&user_id_type=user_id", id)
	get, err := client.Get(context.Background(), url)
	if err != nil {
		return "", err
	}
	defer get.Close()
	return get.ReadAllString(), nil
}

// AppGetDepartmentUser id DepartmentId
func (s *AppService) AppGetDepartmentUser(id string) (string, error) {
	UserData := User{}
	client, err := s.AppClient()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/contact/v3/users/find_by_department?department_id=%s&department_id_type=department_id&page_size=10&user_id_type=user_id", id)
	get, err := client.Get(context.Background(), url)
	if err != nil {
		return "", err
	}
	err = gjson.DecodeTo(get.ReadAllString(), &UserData)
	if err != nil {
		return "", err
	}
	for {
		pageData := User{}
		pageUrl := fmt.Sprintf("https://open.feishu.cn/open-apis/contact/v3/users/find_by_department?department_id=%s&department_id_type=department_id&page_size=10&user_id_type=user_id", id)
		if UserData.Data.HasMore {
			pageGet, err := client.Get(context.Background(), fmt.Sprintf("%s&page_token=%s", pageUrl, UserData.Data.PageToken))
			if err != nil {
				return "", err
			}
			err = gjson.DecodeTo(pageGet.ReadAllString(), &pageData)
			if err != nil {
				return "", err
			}
			UserData.Data.Items = append(UserData.Data.Items, pageData.Data.Items...)
		}
		if pageData.Data.HasMore == false {
			UserData.Data.HasMore = pageData.Data.HasMore
			UserData.Data.PageToken = pageData.Data.PageToken
			break
		}
	}
	defer get.Close()
	encodeString, err := gjson.EncodeString(UserData)
	if err != nil {
		return "", err
	}
	return encodeString, nil
}

// AppGetDepartmentUsers ids DepartmentIdList
func (s *AppService) AppGetDepartmentUsers(ids []string) (string, error) {
	var userList []interface{}
	for i := 0; i < len(ids); i++ {
		department, err := s.AppGetDepartmentUser(ids[i])
		if err != nil {
			return "", err
		}
		jsonDataItem := gjson.New(department)
		itemData := jsonDataItem.Get("data.items").Array()
		userList = append(userList, itemData...)
	}
	newJsonData := gjson.New(userList)
	return newJsonData.String(), nil
}

// AppExcelCreate title, folderToken
func (s *AppService) AppExcelCreate(title, folderToken string) (string, error) {
	url := "https://open.feishu.cn/open-apis/sheets/v3/spreadsheets"
	data := gjson.New(g.Map{"title": title, "folder_token": folderToken})
	appClient, err := s.AppClient()
	if err != nil {
		return "", err
	}
	post, err := appClient.Post(context.Background(), url, data.String())
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}

// AppExcelGetSheetId sheetToken
func (s *AppService) AppExcelGetSheetId(sheetToken string) (string, error) {
	url := "https://open.feishu.cn/open-apis/sheets/v3/spreadsheets/" + sheetToken + "/sheets/query"
	appClient, err := s.AppClient()
	if err != nil {
		return "", err
	}
	get, err := appClient.Get(context.Background(), url)
	if err != nil {
		return "", err
	}
	defer get.Close()
	return get.ReadAllString(), nil
}

// AppExcelInsertLine sheetToken, data
func (s *AppService) AppExcelInsertLine(sheetToken, data string) (string, error) {
	url := "https://open.feishu.cn/open-apis/sheets/v2/spreadsheets/" + sheetToken + "/values_prepend"
	data1 := gjson.New(data)
	appClient, err := s.AppClient()
	if err != nil {
		return "", err
	}
	post, err := appClient.Post(context.Background(), url, data1.String())
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}
