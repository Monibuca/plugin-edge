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
	Origin string //源服务器地址
	config.Pull
}

func (p *EdgeConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig:
		if len(p.Origin) < 4 {
			plugin.Warn("origin config error plugin disabled")
			plugin.RawConfig["enable"] = false
		}
	case *Stream:
		var puller IPuller
		switch p.Origin[:4] {
		case "http":
			puller = new(hdl.HDLPuller)
		case "rtmp":
			puller = new(rtmp.RTMPPuller)
		case "rtsp":
			puller = new(rtsp.RTSPPuller)
		default:
			plugin.Panic("origin config not support")
		}
		err := plugin.Pull(v.Path, p.Origin+v.Path, puller, 0)
		if err != nil {
			plugin.Error("pull", zap.Error(err))
		}
	}
}

var plugin = InstallPlugin(new(EdgeConfig))
