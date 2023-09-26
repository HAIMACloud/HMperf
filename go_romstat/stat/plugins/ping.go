package plugins

import (
	"fmt"

	"romstat/stat/data"
	"romstat/stat/utils"
)

type NetworkPingPlugin struct {
	currentStatLst []*utils.PingStat
}

func (t *NetworkPingPlugin) Open() bool {
	t.currentStatLst = make([]*utils.PingStat, 0)
	go t.runPingSecond(1)
	return true
}

func (t *NetworkPingPlugin) Close() {
}
func (t *NetworkPingPlugin) runPingSecond(count int) error {
	shell := utils.NewAndroidShell()
	stat, err := shell.GetPingStat("www.baidu.com", count)
	if err != nil {
		utils.DebugLogger.Println("ERROR: %s", err.Error())
		return err
	}
	statLstSize := len(t.currentStatLst)
	if statLstSize > 1 { //save the last seconds stat data
		t.currentStatLst = t.currentStatLst[statLstSize-1 : statLstSize]
	}
	t.currentStatLst = append(t.currentStatLst, stat)
	return nil
}
func (t *NetworkPingPlugin) Run() {
}
func (t *NetworkPingPlugin) GetTypes() []*data.PluginType {
	return []*data.PluginType{
		{Name: "rtt", DisplayName: "rtt(ms)", IsCmdShow: true},
		{Name: "loss", DisplayName: "loss", IsCmdShow: true},
	}
}

func (t *NetworkPingPlugin) GetData() map[string]string {
	go t.runPingSecond(5)
	var avgRss, totalRss float64
	if len(t.currentStatLst) == 0 {
		return map[string]string{
			"rtt":  "0.0",
			"loss": "0",
		}
	}
	currentStat := t.currentStatLst[0]
	for _, rss := range currentStat.RssLst {
		totalRss += rss
	}
	if len(currentStat.RssLst) > 0 {
		avgRss = totalRss / float64(len(currentStat.RssLst))
	}
	return map[string]string{
		"rtt":  fmt.Sprintf("%.1f", avgRss),
		"loss": fmt.Sprintf("%d", currentStat.SendPackages-currentStat.RecvPackages),
	}
}
