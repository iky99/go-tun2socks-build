package tun2socks

import (
	"context"
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/eycorsican/go-tun2socks/core"
	xsession "github.com/xtls/xray-core/common/session"
	xcore "github.com/xtls/xray-core/core"
)

var x *xcore.Instance

type PacketFlow interface {
	WritePacket(packet []byte)
}

func InputPacket(data []byte) {
	lwipStack.Write(data)
}

var lwipStack core.LWIPStack

func StartV2Ray(packetFlow PacketFlow, configBytes []byte) {
	if packetFlow == nil {
		return
	}

	lwipStack = core.NewLWIPStack()
	x, err := xcore.StartInstance("json", configBytes)
	if err != nil {
		log.Fatalf("start V instance failed: %v", err)
	}
	debug.SetGCPercent(5)
	ctx := context.Background()
	content := xsession.ContentFromContext(ctx)
	if content == nil {
		content = new(xsession.Content)
		ctx = xsession.ContextWithContent(ctx, content)
	}
	core.RegisterTCPConnHandler(xray.NewTCPHandler(ctx, x))
	core.RegisterUDPConnHandler(xray.NewUDPHandler(ctx, x, 30*time.Second))
	core.RegisterOutputFn(func(data []byte) (int, error) {
		packetFlow.WritePacket(data)
		runtime.GC()
		debug.FreeOSMemory()
		return len(data), nil
	})
}
