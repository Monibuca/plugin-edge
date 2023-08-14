package edge

import (
	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/plugin/hdl/v4"
	"m7s.live/plugin/rtmp/v4"
	"m7s.live/plugin/rtsp/v4"
)

type EdgeConfig struct {
	config.Pull
	Origin string //源服务器地址
}

func (p *EdgeConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig:
		if len(p.Origin) < 4 {
			EdgePlugin.Warn("origin config error plugin disabled")
			EdgePlugin.Disabled = true
		}
	case InvitePublish:
		var puller IPuller
		switch p.Origin[:4] {
		case "http":
			puller = hdl.NewHDLPuller()
		case "rtmp":
			puller = new(rtmp.RTMPPuller)
		case "rtsp":
			puller = new(rtsp.RTSPPuller)
		default:
			EdgePlugin.Panic("origin config not support")
		}
		err := EdgePlugin.Pull(v.Target, p.Origin+v.Target, puller, 0)
		if err != nil {
			EdgePlugin.Error("pull", zap.Error(err))
		}
	}
}

var EdgePlugin = InstallPlugin(new(EdgeConfig))
