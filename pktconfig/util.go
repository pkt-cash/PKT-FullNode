// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package pktconfig

import (
	"crypto/rand"
	"encoding/base64"
	flags "github.com/jessevdk/go-flags"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkt-cash/pktd/btcutil/er"
)

// CreateDefaultConfigFile copies the file sample-pktd.conf to the given destination path,
// and populates it with some randomly generated RPC username and password.
func CreateDefaultConfigFile(destinationPath, sampleFile string) er.R {
	// Create the destination directory if it does not exists
	errr := os.MkdirAll(filepath.Dir(destinationPath), 0700)
	if errr != nil {
		return er.E(errr)
	}

	// We generate a random user and password
	var userPass [2]string
	for i := 0; i < 2; i++ {
		randomBytes := make([]byte, 20)
		if _, errr = rand.Read(randomBytes); errr != nil {
			return er.E(errr)
		} else {
			userPass[i] = base64.StdEncoding.EncodeToString(randomBytes)
		}
	}

	dest, errr := os.OpenFile(destinationPath,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if errr != nil {
		return er.E(errr)
	}
	defer dest.Close()

	// We copy every line from the sample config file to the destination,
	// only replacing the two lines for rpcuser and rpcpass
	for _, line := range strings.Split(sampleFile, "\n") {
		if strings.Contains(line, "rpcuser=") {
			line = "rpcuser=" + userPass[0] + "\n"
		} else if strings.Contains(line, "rpcpass=") {
			line = "rpcpass=" + userPass[1] + "\n"
		}
		if _, errr := dest.WriteString(line + "\n"); errr != nil {
			return er.E(errr)
		}
	}

	return nil
}

type userpass struct {
	Username string `long:"rpcuser"`
	Password string `long:"rpcpass"`
	OldUsername string `long:"username"`
	OldPassword string `long:"password"`
}

// ReadUserPass reads out the username and password from a config file
func ReadUserPass(filePath string) ([]string, er.R) {
	cfg := userpass{}
	parser := flags.NewParser(&cfg, flags.IgnoreUnknown)
	if errr := flags.NewIniParser(parser).ParseFile(filePath); errr != nil {
		return nil, er.E(errr)
	}
	if cfg.Username == "" {
		cfg.Username = cfg.OldUsername
	}
	if cfg.Password == "" {
		cfg.Password = cfg.OldPassword
	}
	return []string{cfg.Username, cfg.Password}, nil
}
