package main

import (
	"context"
	"strconv"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnrpc/routerrpc"
	"github.com/pkt-cash/pktd/lnd/lnwire"
	"github.com/pkt-cash/pktd/lnd/routing/route"
	"github.com/urfave/cli"
)

var queryProbCommand = cli.Command{
	Name:      "queryprob",
	Category:  "Payments",
	Usage:     "Estimate a success probability.",
	ArgsUsage: "from-node to-node amt",
	Action:    actionDecorator(queryProb),
}

func queryProb(ctx *cli.Context) er.R {
	args := ctx.Args()

	if len(args) != 3 {
		return er.E(cli.ShowCommandHelp(ctx, "queryprob"))
	}

	fromNode, err := route.NewVertexFromStr(args.Get(0))
	if err != nil {
		return er.Errorf("invalid from node key: %v", err)
	}

	toNode, err := route.NewVertexFromStr(args.Get(1))
	if err != nil {
		return er.Errorf("invalid to node key: %v", err)
	}

	amtSat, errr := strconv.ParseUint(args.Get(2), 10, 64)
	if errr != nil {
		return er.Errorf("invalid amt: %v", errr)
	}

	amtMsat := lnwire.NewMSatFromSatoshis(
		btcutil.Amount(amtSat),
	)

	conn := getClientConn(ctx, false)
	defer conn.Close()

	client := routerrpc.NewRouterClient(conn)

	req := &routerrpc.QueryProbabilityRequest{
		FromNode: fromNode[:],
		ToNode:   toNode[:],
		AmtMsat:  int64(amtMsat),
	}
	rpcCtx := context.Background()
	response, errr := client.QueryProbability(rpcCtx, req)
	if errr != nil {
		return er.E(errr)
	}

	printRespJSON(response)

	return nil
}
