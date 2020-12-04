package zpay32

import (
	"strconv"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnwire"
)

var (
	// toMSat is a map from a unit to a function that converts an amount
	// of that unit to millisatoshis.
	toMSat = map[byte]func(uint64) (lnwire.MilliSatoshi, er.R){
		'm': mBtcToMSat,
		'u': uBtcToMSat,
		'n': nBtcToMSat,
		'p': pBtcToMSat,
	}

	// fromMSat is a map from a unit to a function that converts an amount
	// in millisatoshis to an amount of that unit.
	fromMSat = map[byte]func(lnwire.MilliSatoshi) (uint64, er.R){
		'm': mSatToMBtc,
		'u': mSatToUBtc,
		'n': mSatToNBtc,
		'p': mSatToPBtc,
	}
)

// mBtcToMSat converts the given amount in milliBTC to millisatoshis.
func mBtcToMSat(m uint64) (lnwire.MilliSatoshi, er.R) {
	return lnwire.MilliSatoshi(m) * 100000000, nil
}

// uBtcToMSat converts the given amount in microBTC to millisatoshis.
func uBtcToMSat(u uint64) (lnwire.MilliSatoshi, er.R) {
	return lnwire.MilliSatoshi(u * 100000), nil
}

// nBtcToMSat converts the given amount in nanoBTC to millisatoshis.
func nBtcToMSat(n uint64) (lnwire.MilliSatoshi, er.R) {
	return lnwire.MilliSatoshi(n * 100), nil
}

// pBtcToMSat converts the given amount in picoBTC to millisatoshis.
func pBtcToMSat(p uint64) (lnwire.MilliSatoshi, er.R) {
	if p < 10 {
		return 0, er.Errorf("minimum amount is 10p")
	}
	if p%10 != 0 {
		return 0, er.Errorf("amount %d pBTC not expressible in msat",
			p)
	}
	return lnwire.MilliSatoshi(p / 10), nil
}

// mSatToMBtc converts the given amount in millisatoshis to milliBTC.
func mSatToMBtc(msat lnwire.MilliSatoshi) (uint64, er.R) {
	if msat%100000000 != 0 {
		return 0, er.Errorf("%d msat not expressible "+
			"in mBTC", msat)
	}
	return uint64(msat / 100000000), nil
}

// mSatToUBtc converts the given amount in millisatoshis to microBTC.
func mSatToUBtc(msat lnwire.MilliSatoshi) (uint64, er.R) {
	if msat%100000 != 0 {
		return 0, er.Errorf("%d msat not expressible "+
			"in uBTC", msat)
	}
	return uint64(msat / 100000), nil
}

// mSatToNBtc converts the given amount in millisatoshis to nanoBTC.
func mSatToNBtc(msat lnwire.MilliSatoshi) (uint64, er.R) {
	if msat%100 != 0 {
		return 0, er.Errorf("%d msat not expressible in nBTC", msat)
	}
	return uint64(msat / 100), nil
}

// mSatToPBtc converts the given amount in millisatoshis to picoBTC.
func mSatToPBtc(msat lnwire.MilliSatoshi) (uint64, er.R) {
	return uint64(msat * 10), nil
}

// decodeAmount returns the amount encoded by the provided string in
// millisatoshi.
func decodeAmount(amount string) (lnwire.MilliSatoshi, er.R) {
	if len(amount) < 1 {
		return 0, er.Errorf("amount must be non-empty")
	}

	// If last character is a digit, then the amount can just be
	// interpreted as BTC.
	char := amount[len(amount)-1]
	digit := char - '0'
	if digit >= 0 && digit <= 9 {
		btc, err := strconv.ParseUint(amount, 10, 64)
		if err != nil {
			return 0, er.E(err)
		}
		return lnwire.MilliSatoshi(btc) * mSatPerBtc, nil
	}

	// If not a digit, it must be part of the known units.
	conv, ok := toMSat[char]
	if !ok {
		return 0, er.Errorf("unknown multiplier %c", char)
	}

	// Known unit.
	num := amount[:len(amount)-1]
	if len(num) < 1 {
		return 0, er.Errorf("number must be non-empty")
	}

	am, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		return 0, er.E(err)
	}

	return conv(am)
}

// encodeAmount encodes the provided millisatoshi amount using as few characters
// as possible.
func encodeAmount(msat lnwire.MilliSatoshi) (string, er.R) {
	// If possible to express in BTC, that will always be the shortest
	// representation.
	if msat%mSatPerBtc == 0 {
		return strconv.FormatInt(int64(msat/mSatPerBtc), 10), nil
	}

	// Should always be expressible in pico BTC.
	pico, err := fromMSat['p'](msat)
	if err != nil {
		return "", er.Errorf("unable to express %d msat as pBTC: %v",
			msat, err)
	}
	shortened := strconv.FormatUint(pico, 10) + "p"
	for unit, conv := range fromMSat {
		am, err := conv(msat)
		if err != nil {
			// Not expressible using this unit.
			continue
		}

		// Save the shortest found representation.
		str := strconv.FormatUint(am, 10) + string(unit)
		if len(str) < len(shortened) {
			shortened = str
		}
	}

	return shortened, nil
}
