package weixin

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"strings"
)

type weiXinAdapter struct{}

var adapter = weiXinAdapter{}

func Adapter() *weiXinAdapter {
	return &adapter
}

func (s *weiXinAdapter) BotMsg(bot string, mobileList []string, msgType, content string) (string, error) {
	contentJson := gjson.New(content)
	_ = contentJson.Set("mentioned_mobile_list", mobileList)
	data := gjson.New(g.Map{"msgtype": msgType, msgType: contentJson}).String()
	post, err := g.Client().Header(map[string]string{"Content-Type": "application/json"}).Post(context.Background(), bot, data)
	if err != nil {
		return "", err
	}
	defer post.Close()
	return post.ReadAllString(), nil
}

type AppService struct {
	AppId  string
	CorpId string
	Secret string
}

func (w *AppService) GetToken() (string, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", w.CorpId, w.Secret)
	get, err := g.Client().Get(context.Background(), url)
	if err != nil {
		return "", err
	}
	dataJson := gjson.New(get.ReadAllString())
	return dataJson.Get("access_token").String(), nil
}

func (w *AppService) GetTokenCache() (string, error) {
	dirPath := gfile.Temp()
	filePath := gfile.Join(dirPath, "1676516581000")
	if gfile.Exists(filePath) {
		token := gfile.GetContents(filePath)
		checkToken, err := g.Client().Get(context.Background(), fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/tag/list?access_token=%s", token))
		if err != nil {
			return "", err
		}
		checkTokenJson := gjson.New(checkToken.ReadAllString())
		if checkTokenJson.Get("errcode").String() == "0" {
			return token, nil
		} else {
			token, err := w.GetToken()
			if err != nil {
				return "", err
			}
			if err != nil {
				return "", err
			}
			return token, nil
		}
	}
	token, err := w.GetToken()
	if err != nil {
		return "", err
	}
	err = gfile.PutContents(filePath, token)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetUid user = email || mobile
func (w *AppService) GetUid(token, user string) (string, error) {
	if user == "@all" {
		return user, nil
	}
	dataJson := gjson.New(``)
	var url string
	if strings.Contains(user, "@") {
		dataJson.Set("email", user)
		url = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/get_userid_by_email?access_token=%s", token)
	} else {
		dataJson.Set("mobile", user)
		url = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserid?access_token=%s", token)
	}
	post, err := g.Client().Header(map[string]string{"Content-Type": "application/json"}).Post(context.Background(), url, dataJson.String())
	if err != nil {
		return "", err
	}
	postJson := gjson.New(post.ReadAllString())
	return postJson.Get("userid").String(), nil
}

// SendMsg ([]string{"18888888888"}, "text", `{"content":"东方快递使命必达!"}`)
func (w *AppService) SendMsg(toUser []string, msgType, content string) (string, error) {
	cacheToken, err := w.GetTokenCache()
	if err != nil {
		return "", err
	}
	var uidList []string
	for i := 0; i < len(toUser); i++ {
		uid, err := w.GetUid(cacheToken, toUser[i])
		if err != nil {
			return "", err
		}
		uidList = append(uidList, uid)
	}
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", cacheToken)
	data := gjson.New(g.Map{"touser": strings.Join(uidList, "|"), "msgtype": msgType, "agentid": w.AppId, msgType: gjson.New(content)})
	post, err := g.Client().Header(map[string]string{"Content-Type": "application/json"}).Post(context.Background(), url, data)
	if err != nil {
		return "", err
	}
	return post.ReadAllString(), nil
}
