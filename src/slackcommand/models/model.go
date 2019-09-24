package models

import (
	"github.com/astaxie/beego/orm"
	"testbedpool/tools"
	"strings"
	"encoding/json"
	"github.com/astaxie/beego"
)

type SlackUser struct {
	Id        int64
	Name      string
	Latest    float64
	ChannelId string
	AppToken  string
	TimeStamp string
}

type MegaData struct {
	Ok bool `json:"ok"`
	PA []PoolApi `json:"pools"`
	TA []TestbedApi `json:"testbeds"`
}

type PoolApi struct {
	Id        int64 `json:"id"`
	Title     string `json:"title"`
	Capacity  int `json:"capacity"`
	Available int `json:"available"`
	Exception string `json:"exception"`
}

type TestbedApi struct {
	TestbedName  string `json:"testbedname"`
	Vcip         string `json:"vcip"`
	Esxips       string `json:"esxips"`
	Vcbuild      string `json:"vcbuild"`
	Esxbuild     string `json:"esxbuild"`
	Changeset    string `json:"cln"`
	Lifetime     string `json:"lifetime"`
	PoolUniqueId string `json:"puid"`
	Exception    string `json:"exception"`
}

func CheckBlackList(cmd string) bool {
        var exist bool
        blackList := []string{".vi", ".vim", ".top", ".less", ".more"}
        for _, bl := range blackList {
                if strings.Contains(bl, cmd) {
                        exist = true
                        break
                }
        }
        return exist
}


func QueryDataFromSlackApi(url string) (megaData *MegaData, err error) {
	content, err0 := tools.GetContentFromUrl(url)
	if err0 != nil {
		err = err0
		return
	}
	data := &MegaData{}
	if err1 := json.Unmarshal(content, data); err1 != nil {
		err = err1
		return
	} else {
		megaData = data
	}
	return
}

func QuerySlackUserByNameFromDB(name string) (*SlackUser, error) {
	o := orm.NewOrm()
	o.Using("default")
	slackUser := &SlackUser{Name:name}
	err := o.Read(slackUser, "Name")
	return slackUser, err
}

func UpdateSlackUserLatestTimeByName(name string, lst float64) error {
	o := orm.NewOrm()
	o.Using("default")
	slackUser, err := QuerySlackUserByNameFromDB(name)
	if err != nil {
		return err
	}
	if o.Read(slackUser) == nil {
		slackUser.Latest = lst
		_, err := o.Update(slackUser)
		if err != nil {
			beego.Error(err)
		}
	}
	return nil
}

func UpdateProgramStampByName(name, timestamp string) error {
	o := orm.NewOrm()
	o.Using("default")
	slackUser, err := QuerySlackUserByNameFromDB(name)
	if err != nil {
		return err
	}
	if o.Read(slackUser) == nil {
		slackUser.TimeStamp = timestamp
		_, err := o.Update(slackUser)
		if err != nil {
			beego.Error(err)
		}
	}
	return nil
}

func CheckUserExist(name string) bool {
	userList, err := QueryUsersFromDB()
	if err != nil {
		return false
	}
	if len(userList) == 0 {
		return false
	}
	for _, user := range userList {
		if user.Name == name {
			return true
		}
	}
	return false
}

func InsertUserIntoDB(name, key string) error {
	o := orm.NewOrm()
	o.Using("default")
	slackUser := &SlackUser{
		Name:name,
		AppToken:key,
	}
	_, err := o.Insert(slackUser)
	return err
}

func QueryUsersFromDB() ([]*SlackUser, error) {
	o := orm.NewOrm()
	o.Using("default")
	slackUsers := make([]*SlackUser, 0)
	_, err := o.QueryTable("slack_user").All(&slackUsers)
	return slackUsers, err
}


