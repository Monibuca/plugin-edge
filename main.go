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
	puller IPuller
}

func (p *EdgeConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig:
		if len(p.Origin) < 4 {
			plugin.Warn("origin config error plugin disabled")
			plugin.RawConfig["enable"] = false
		} else {
			switch p.Origin[:4] {
			case "http":
				p.puller = new(hdl.HDLPuller)
			case "rtmp":
				p.puller = new(rtmp.RTMPPuller)
			case "rtsp":
				p.puller = new(rtsp.RTSPPuller)
			default:
				plugin.Panic("origin config not support")
			}
		}
	case *Stream:
		err := plugin.Pull(v.Path, p.Origin+v.Path, p.puller, false)
		if err != nil {
			plugin.Error("pull", zap.Error(err))
		}
	}
}

var plugin = InstallPlugin(new(EdgeConfig))
