package main

import (
	"github.com/astaxie/beego/orm"
	"testbedpool/slacknotifier"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"github.com/astaxie/beego"
	"strconv"
	"strings"
	"testbedpool/tools"
	"slackcommand/models"
	"slackcommand/config"
	"flag"
	"os"
	"time"
)

const (
	_DB_NAME = "slack.db"
	_SQLITES_DRIVER = "sqlite3"
)

func init() {
	if !tools.CheckFileExist(_DB_NAME) {
		tools.CreateFile(tools.GetCurrentDir(), _DB_NAME)
	}
	orm.RegisterModel(new(models.SlackUser))
	orm.RegisterDriver(_SQLITES_DRIVER, orm.DRSqlite)
	orm.RegisterDataBase("default", _SQLITES_DRIVER, _DB_NAME, 10000)
	orm.RunSyncdb("default", false, true)
}

func MonitorCmd(user *models.SlackUser, slack *slacknotifier.Notifier, ch chan bool) {
	_, msgs := slack.QueryHistory()
	if len(msgs) != 0 {
		latestMsg := msgs[0]
		text := latestMsg.Text
		timevalue := latestMsg.Ts
		lst, _ := strconv.ParseFloat(timevalue, 64)
		// check new command detected
		if lst > user.Latest {
			// check command is legal
			valid, cmds := RecognizeSlackCommand(text)
			if valid {
				config.HandleCommads(cmds, slack)
			} else {
				slack.SendMessage("invalid command,please type .help to show usage.")
			}
			models.UpdateSlackUserLatestTimeByName(user.Name, lst)
		}
	}
	ch <- true
}

func RecognizeSlackCommand(text string) (valid bool, CommandArray []string) {
	CmdStr := strings.Split(text, " ")
	if strings.HasPrefix(CmdStr[0], ".") {
		valid = true
		CommandArray = CmdStr
	}
	return
}

func GetHostName() (name string) {
	cmd := "whoami"
	if err, stdout, _ := tools.ExecuteShellCommand(cmd); err == nil {
		name = strings.Replace(stdout, " ", "", -1)
		name = strings.Replace(name, "\n", "", -1)
	}
	return
}

func CheckProgramNeedUpgrade(user *models.SlackUser) bool {
	timeStamp := ""
	cmd := "ls -lr /mts/home4/lsui/sdbc/slackcommand | awk {'print $6 $7 $8'}"
	if err, stdout, _ := tools.ExecuteShellCommand(cmd); err == nil {
		timeStamp = strings.Replace(stdout, " ", "", -1)
		timeStamp = strings.Replace(timeStamp, "\n", "", -1)
	} else {
		beego.Warning(err)
	}
	if timeStamp == "" {
		return false
	}
	if timeStamp != user.TimeStamp {
		models.UpdateProgramStampByName(user.Name, timeStamp)
		time.Sleep(2 * time.Second)
		return true
	}
	return false
}


func main() {
	key := flag.String("k", "", "Bot User OAuth Access Token of your slack app")
	local := flag.Bool("l", false, "use token in db")
	help := flag.Bool("h", false, "show help message")
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("type -h to see usgae.")
		return
	}
	if *help {
		fmt.Println("-k, Bot User OAuth Access Token of your slack app.")
		return
	}
	if *key == "" {
		if !*local {
			fmt.Println("type -h to see usgae.")
			return
		}
	}else{
		if !strings.Contains(*key,"-"){
			fmt.Println("Slack Auth Token Invalid.")
			return
		}
	}
	// initialize
	name := GetHostName()
	if name == "" {
		beego.Warning("Execute whoami fail")
		return
	}
	if !models.CheckUserExist(name) {
		if !*local {
			if err := models.InsertUserIntoDB(name, *key); err != nil {
				beego.Warning(fmt.Sprintf("User: %s is invalid", name))
				return
			}
		}
	}
	for {
		users, _ := models.QueryUsersFromDB()
		dbChan := make(chan bool, len(users))
		for _, user := range users {
			notifer := slacknotifier.NewSlackNotifier(user.Name, user.AppToken)
			if err := notifer.Initial(); err != nil {
				dbChan <- true
				continue
			}
			currentMinute := time.Now().Minute()
			if currentMinute % 30 == 0 {
				if CheckProgramNeedUpgrade(user) {
					sdbc_cmd := fmt.Sprintf("sh /mts/home4/lsui/sdbc/sdbc %s", user.AppToken)
					msg := fmt.Sprintf("New SDBC release version detected, please run below command in your worker!\n%s\n",sdbc_cmd)
					notifer.SendMessage(msg)
				}
			}
			go MonitorCmd(user, notifer, dbChan)
		}
		for i := 0; i < len(users); i++ {
			<-dbChan
		}
	}
}

