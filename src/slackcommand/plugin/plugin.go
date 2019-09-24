package plugin

import (
	"testbedpool/slacknotifier"
	"strings"
	"testbedpool/tools"
	"slackcommand/models"
	"fmt"
)


func SimulateDbc(cmds string, slack *slacknotifier.Notifier) {
	cmdArr := strings.Split(cmds, ".")
	shellCmd := cmdArr[1]
	if shellCmd == "" {
		return
	}
	var msg string
	cmd := strings.Join(cmdArr[1:], ".")
	if err, stdout, stderr := tools.ExecuteShellCommand(cmd); err == nil {
		msg = fmt.Sprintf(`-bash-4.1$  %s\n%s\n%s\n`, cmd, stdout, stderr)
	} else {
		msg = fmt.Sprintf("Hit issue: %s %s", err.Error(), stderr)
	}
	slack.SendMessage(msg)
}


func RunDbcCommand(cmds string, slack *slacknotifier.Notifier) {
	go SimulateDbc(cmds, slack)
	slack.SendMessage("Executing command, maybe occurs delay...")
}

func TriggerFastSvs(cln string, slack *slacknotifier.Notifier, fast bool) {
	var triggerCmd string
	if fast{
		triggerCmd = fmt.Sprintf("./mts/git/vsan-tools/bin/fastsvs -c %s", cln)
	}else{
		triggerCmd = fmt.Sprintf(".svs precheckin -q -c %s", cln)
	}
	go SimulateDbc(triggerCmd, slack)
	slack.SendMessage("It may cost about 3 minutes,when job done,result will be sent soon.")
}

func TriggerGoBuild(cln string, slack *slacknotifier.Notifier, opt string) {
	var triggerCmd string
	switch opt{
	case "server":
		triggerCmd = fmt.Sprintf(".gobuild sandbox queue server --changeset %s --branch=main --buildtype release --accept-defaults --store-tree", cln)
	case "vcenter-all":
		triggerCmd = fmt.Sprintf(".gobuild sandbox queue vcenter-all --changeset %s --branch=main --buildtype release --accept-defaults --store-tree", cln)
	}
	go SimulateDbc(triggerCmd, slack)
	slack.SendMessage("It may cost minutes,when job done,result will be sent soon.")
}

func TriggerRemoteCommand(cmds, opt string, slack *slacknotifier.Notifier){
	var pwd string
	user := "root"
	switch opt{
	case "esxi":
		pwd = "ca$hc0w"
	case "vc":
		pwd = "vmware"
	}
        cmdArr := strings.Split(cmds, ".")
        var msg string
        cmd := strings.Join(cmdArr[1:], ".")
        remoteCmds := strings.Split(cmd, " ")
	if len(remoteCmds) < 3{
		slack.SendMessage("USAGE: .sshhost 10.0.0.127 ls -l")
		return
	}
	host := remoteCmds[1]
	rcs := remoteCmds[2:]
	rc := strings.Join(rcs, " ")
	if models.CheckBlackList(rcs[0]){
		msg = fmt.Sprintf("Cmd : %s can not be executed in remote server.\n", rc)
		slack.SendMessage(msg)
		return
	}
	fmt.Println(pwd)
        if err, stdout,_ := tools.ExecuteRemoteCommand(host, user, pwd, rc); err == nil {
                msg = fmt.Sprintf(`-bash-4.1$ ssh %s@%s:  %s\n%s\n`, user, host, rc, stdout)
        } else {
                msg = fmt.Sprintf("Hit issue: %s", err.Error())
        }
        slack.SendMessage(msg)

}

func QueryTestbedsOfUsers(slack *slacknotifier.Notifier) {
	var msg string
	user := slack.Name
	url := fmt.Sprintf("http://10.161.184.86:8080/slackcommand/api/testbed/list/%s", user)
	result, err := models.QueryDataFromSlackApi(url)
	if err == nil {
		if result.Ok {
			msg = "*Testbed*        *Vc*        *Expired*\n"
			for _, tb := range result.TA {
				msg += fmt.Sprintf("%s       %s        %s\n", tb.TestbedName, tb.Vcip, tb.Lifetime)
			}
		} else {
			msg = result.TA[0].Exception
		}
	} else {
		msg = fmt.Sprintf("Hit issue:%s", err.Error())
	}
	slack.SendMessage(msg)
}

func ShowPoolDetailInfo(id string, slack *slacknotifier.Notifier) {
	var msg string
	url := fmt.Sprintf("http://10.161.184.86:8080/slackcommand/api/pool/show/%s", id)
	result, err := models.QueryDataFromSlackApi(url)
	if err == nil {
		if result.Ok {
			msg = fmt.Sprintf("Id:%d\nTitle:%s\nCapacity:%d\nAvailable:%d\n", result.PA[0].Id, result.PA[0].Title, result.PA[0].Capacity, result.PA[0].Available)
		} else {
			msg = result.PA[0].Exception
		}
	} else {
		msg = fmt.Sprintf("Hit issue:%s", err.Error())
	}
	slack.SendMessage(msg)
}

func GetTestbedFromPool(id string, slack *slacknotifier.Notifier) {
	var msg string
	url := fmt.Sprintf("http://10.161.184.86:8080/slackcommand/api/testbed/get/%s/%s", slack.Name, id)
	result, err := models.QueryDataFromSlackApi(url)
	if err == nil {
		if result.Ok {
			messageTemplate := `
        Dear %s, you got a deployed testbed,

        *Testbed Name*: %s
        *VC Host*: %s
        *Esxi Host*: %s
        *VC Build*: %s
        *Expired in*: %s

        DUE TO NIMBUS RESOURCE CONSTRAINED, PLEASE GET TESTBED ON DEMAND, DO NOT WASTE THEM.

        Thanks for your understanding and great support!!!
        `
			msg = fmt.Sprintf(messageTemplate, slack.Name, result.TA[0].TestbedName, result.TA[0].Vcip, result.TA[0].Esxips, result.TA[0].Vcbuild, result.TA[0].Lifetime)
		} else {
			msg = result.TA[0].Exception
		}
	} else {
		msg = fmt.Sprintf("Hit issue:%s", err.Error())
	}
	slack.SendMessage(msg)
}

func ShowHelpMessage(cmdhelp map[string]string, slack *slacknotifier.Notifier) {
	msg := ""
	for k, v := range cmdhelp {
		msg += fmt.Sprintf("%s -> %s\n", k, v)
	}
	slack.SendMessage(msg)
}

func ListAllPools(slack *slacknotifier.Notifier) {
	var msg string
	url := "http://10.161.184.86:8080/slackcommand/api/pool/get/"
	result, err := models.QueryDataFromSlackApi(url)
	if err == nil {
		if result.Ok {
			msg = fmt.Sprintf("You also can visit web page:<%s|Testbed Pool>\n", "http://10.161.154.52:8080/")
			for _, mypool := range result.PA {
				msg += fmt.Sprintf("%d    %s\n", mypool.Id, mypool.Title)
			}
		} else {
			msg = result.TA[0].Exception
		}
	} else {
		msg = fmt.Sprintf("Hit issue:%s", err.Error())
	}
	slack.SendMessage(msg)
}

/*
Dev can define Handler function here, which to be called in config.go
It's better to receive *slacknotifier.Notifier as a handler to send message to slack
Suggest do not write business or service code here, instead of use rest api to deal with request
When need to parse the json from rest api, you can call similar function models.QueryDataFromSlackApi(url
	func QueryDataFromSlackApi(url string) (megaData *MegaData, err error) {
dev can extend the MegaData struct to fill new request
MegaData struct:
type MegaData struct {
	Ok bool `json:"ok"`
	PA []PoolApi `json:"pools"`
	TA []TestbedApi `json:"testbeds"`
}
dev need to define new struct to receive the result which your service provided
For example:
type YourStruct Struct{
	XXX string `json:"xxx"`
	...
}
type MegaData struct {
	Ok bool `json:"ok"`
	PA []PoolApi `json:"pools"`
	TA []TestbedApi `json:"testbeds"`
	XXX []YourStruct `json:"xxx"`
}
 */
