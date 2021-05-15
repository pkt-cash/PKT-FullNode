package chanacceptor

import (
	"strings"
	"testing"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/lnd/input"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/lnd/lnwallet/chancloser"
	"github.com/pkt-cash/pktd/lnd/lnwire"
	"github.com/stretchr/testify/require"
)

// TestValidateAcceptorResponse test validation of acceptor responses.
func TestValidateAcceptorResponse(t *testing.T) {
	var (
		customError = er.New("custom error")
		validAddr   = "bcrt1qwrmq9uca0t3dy9t9wtuq5tm4405r7tfzyqn9pp"
		addr, _     = chancloser.ParseUpfrontShutdownAddress(
			validAddr, &chaincfg.TestNet3Params,
		)
	)

	tests := []struct {
		name        string
		dustLimit   btcutil.Amount
		response    lnrpc.ChannelAcceptResponse
		accept      bool
		acceptorErr er.R
		error       er.R
		shutdown    lnwire.DeliveryAddress
	}{
		{
			name: "accepted with error",
			response: lnrpc.ChannelAcceptResponse{
				Accept: true,
				Error:  customError.String(),
			},
			accept:      false,
			acceptorErr: errChannelRejected.Default(),
			error:       errAcceptWithError.Default(),
		},
		{
			name: "custom error too long",
			response: lnrpc.ChannelAcceptResponse{
				Accept: false,
				Error:  strings.Repeat(" ", maxErrorLength+1),
			},
			accept:      false,
			acceptorErr: errChannelRejected.Default(),
			error:       errCustomLength.Default(),
		},
		{
			name: "accepted",
			response: lnrpc.ChannelAcceptResponse{
				Accept:          true,
				UpfrontShutdown: validAddr,
			},
			accept:      true,
			acceptorErr: nil,
			error:       nil,
			shutdown:    addr,
		},
		{
			name: "rejected with error",
			response: lnrpc.ChannelAcceptResponse{
				Accept: false,
				Error:  customError.String(),
			},
			accept:      false,
			acceptorErr: customError,
			error:       nil,
		},
		{
			name: "rejected with no error",
			response: lnrpc.ChannelAcceptResponse{
				Accept: false,
			},
			accept:      false,
			acceptorErr: errChannelRejected.Default(),
			error:       nil,
		},
		{
			name: "invalid upfront shutdown",
			response: lnrpc.ChannelAcceptResponse{
				Accept:          true,
				UpfrontShutdown: "invalid addr",
			},
			accept:      false,
			acceptorErr: errChannelRejected.Default(),
			error:       errInvalidUpfrontShutdown.Default(),
		},
		{
			name:      "reserve too low",
			dustLimit: 100,
			response: lnrpc.ChannelAcceptResponse{
				Accept:     true,
				ReserveSat: 10,
			},
			accept:      false,
			acceptorErr: errChannelRejected.Default(),
			error:       errInsufficientReserve.Default(),
		},
		{
			name:      "max htlcs too high",
			dustLimit: 100,
			response: lnrpc.ChannelAcceptResponse{
				Accept:       true,
				MaxHtlcCount: 1 + input.MaxHTLCNumber/2,
			},
			accept:      false,
			acceptorErr: errChannelRejected.Default(),
			error:       errMaxHtlcTooHigh.Default(),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			// Create an acceptor, everything can be nil because
			// we just need the params.
			acceptor := NewRPCAcceptor(
				nil, nil, 0, &chaincfg.TestNet3Params, nil,
			)

			accept, acceptErr, shutdown, err := acceptor.validateAcceptorResponse(
				test.dustLimit, test.response,
			)
			require.Equal(t, test.accept, accept)
			require.True(t, er.FuzzyEquals(test.acceptorErr, acceptErr))
			require.True(t, er.FuzzyEquals(test.error, err))
			require.Equal(t, test.shutdown, shutdown)
		})
	}
}
