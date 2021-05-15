package main

import (
	"context"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnrpc/routerrpc"

	"github.com/urfave/cli"
)

var resetMissionControlCommand = cli.Command{
	Name:     "resetmc",
	Category: "Payments",
	Usage:    "Reset internal mission control state.",
	Action:   actionDecorator(resetMissionControl),
}

func resetMissionControl(ctx *cli.Context) er.R {
	conn := getClientConn(ctx, false)
	defer conn.Close()

	client := routerrpc.NewRouterClient(conn)

	req := &routerrpc.ResetMissionControlRequest{}
	rpcCtx := context.Background()
	_, errr := client.ResetMissionControl(rpcCtx, req)
	return er.E(errr)
}
