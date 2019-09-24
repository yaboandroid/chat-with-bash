package config

import (
	"strings"
	"testbedpool/slacknotifier"
	"slackcommand/plugin"
	"slackcommand/models"
)

func HelpMessage() map[string]string {
	/*
	This part is for generate help messages when user type .help
	Developer should add the new key:value in messages map
	For example:
	".xxx":"xxx xxxx",
	 */
	messages := map[string]string{
		".help":"show help message",
		".plist":"show all pools and relate pool id",
		".pshow":"show more detail info of pool, like: .show [pool id]",
		".pget":"get testbed from pool,like: .gt [pool id]",
		".pall":"show all testbeds of users",
		".fastsvs":"run fastsvs command, like .fastsvs [cln]",
		".svs":"run regular svs command, like .svs [cln]",
		".bldsvr":"run gobuild server for main branch command, like .bldsvr [cln]",
		".bldvc":"run gobuild vcenter-all for main branch command, like .bldvc [cln]",
		".sshhost":"run command in remote esxi host, like .sshhost HOST_IP CMD",
		".sshvc":"run command in remote vc, like .sshvc VC_IP CMD",
		".[shell command]":"simulate bash in dbc,like: .ls, .pwd, .ps -ef | grep xxx ...",
	}
	return messages
}

func HandleCommads(cmmds []string, slack *slacknotifier.Notifier) {
	switch cmmds[0] {
	default:
		// DO NOT MODIFY DEFAULT CONDITION UNLESS KNOW WHAT YOU ARE DOING
		if models.CheckBlackList(cmmds[0]) {
			slack.SendMessage("OOPS,edit or interation commands not support in sdbc.")
		} else {
			cmds := strings.Join(cmmds, " ")
			if strings.Contains(cmds, ".tail -f") {
				slack.SendMessage("OOPS,*.tail -f* not support in sdbc.")
			} else if strings.Contains(cmds, ".p4 change") {
				slack.SendMessage("OOPS, *.p4 change* not support in sdbc.")
			} else {
				if strings.Contains(cmds, "&gt;") {
					cmds = strings.Replace(cmds, "&gt;", ">", -1)
				} else if strings.Contains(cmds, "&amp;") {
					cmds = strings.Replace(cmds, "&amp;", "&", -1)
				} else if strings.Contains(cmds, "&lt;") {
					cmds = strings.Replace(cmds, "&lt;", "<", -1)
				}
				plugin.RunDbcCommand(cmds, slack)
			}
		}
	case ".help":
		plugin.ShowHelpMessage(HelpMessage(), slack)
	case ".plist":
		plugin.ListAllPools(slack)
	case ".pget":
		if len(cmmds) == 2 {
			plugin.GetTestbedFromPool(cmmds[1], slack)
		} else {
			slack.SendMessage("USAGE: .pget POOL_ID.")
		}
	case ".pshow":
		if len(cmmds) == 2 {
			plugin.ShowPoolDetailInfo(cmmds[1], slack)
		} else {
			slack.SendMessage("USAGE: .pshow POOL_ID.")
		}
	case ".pall":
		plugin.QueryTestbedsOfUsers(slack)
	case ".fastsvs":
		if len(cmmds) == 2 {
			plugin.TriggerFastSvs(cmmds[1], slack, true)
		} else {
			slack.SendMessage("USAGE: .fastsvs CLN.")
		}
	case ".svs":
		if len(cmmds) == 2 {
                        plugin.TriggerFastSvs(cmmds[1], slack, false)
                } else {
                        slack.SendMessage("USAGE: .svs CLN.")
                }
	case ".bldsvr":
		if len(cmmds) == 2 {
                        plugin.TriggerGoBuild(cmmds[1], slack, "server")
                } else {
                        slack.SendMessage("USAGE: .bldsvr CLN, build server with CLN.")
                }
	case ".bldvc":
		if len(cmmds) == 2 {
                        plugin.TriggerGoBuild(cmmds[1], slack, "vcenter-all")
                } else {
                        slack.SendMessage("USAGE: .bldvc CLN, build vcenter-all with CLN.")
                }
	case ".sshhost":
		cmds := strings.Join(cmmds, " ")
		plugin.TriggerRemoteCommand(cmds, "esxi", slack)
	case ".sshvc":
		cmds := strings.Join(cmmds, " ")
		plugin.TriggerRemoteCommand(cmds, "vc", slack)

	/*
	dev can add case condition here.
	conditon is specified command for the simulator DBC.
	for the command, Prefix must be ".", otherwise submit will be reject
	if command has parameters, dev can handle here or in corresponding response function
	Example:
	case ".xxx":
		plugin.FuncXXXX()
	 */
	}
	return
}



