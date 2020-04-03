package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

func isUpper(s string) bool {
	return s[0] == strings.ToUpper(s)[0]
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

type xsymbol struct {
	name string
	pos  token.Pos
}

func getSymbol(file, key string) (*xsymbol, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	for _, d := range node.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			if gd.Tok == token.VAR {
			} else if gd.Tok == token.CONST {
			} else if gd.Tok == token.TYPE {
			} else if gd.Tok == token.FUNC {
			} else if gd.Tok == token.IMPORT {
				continue
			} else {
				fmt.Printf("Unrecognized token: %v\n", gd)
				continue
			}
			for _, s := range gd.Specs {
				if vs, ok := s.(*ast.ValueSpec); ok {
					for _, n := range vs.Names {
						if !strings.Contains(n.Name, key) {
							return &xsymbol{name: n.Name, pos: n.Pos()}, nil
						}
					}
				} else if ts, ok := s.(*ast.TypeSpec); ok {
					//fmt.Printf("type %v %v\n", ts.Name, ts.Pos())
					if !strings.Contains(ts.Name.String(), key) {
						return &xsymbol{name: ts.Name.String(), pos: ts.Pos()}, nil
					}
				} else {
					fmt.Printf("Unrecognized spec: %v\n", s)
				}
			}
		} else if fd, ok := d.(*ast.FuncDecl); ok {
			if fd.Name != nil {
				//fmt.Printf("func %v %v\n", fd.Name, fd.Pos())
				if fd.Recv != nil {
					// This is a method, we don't want to rename it because it should
					// not be prefixed with the new package name.
					continue
				}
				if !strings.Contains(fd.Name.String(), key) {
					return &xsymbol{name: fd.Name.String(), pos: fd.Name.Pos()}, nil
				}
				//fmt.Printf("func %v %v\n", fd.Name, fd.Pos())
			} else {
				//fmt.Printf("func lower %v\n", fd.Name)
			}
		} else {
			fmt.Printf("Unrecognized decl: %v (%v)\n", d, reflect.TypeOf(d))
		}
	}
	return nil, nil
}

func cmd(name string, args ...string) []string {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	err := cmd.Run()
	if err != nil {
		fmt.Println(errStr)
		panic(err)
	}
	return strings.Split(outStr, "\n")
}

type fileInfo struct {
	fi   os.FileInfo
	path string
}

func fixPath(cwd, p string) string {
	if path, err := filepath.Abs(p); err != nil {
		panic(err)
	} else if out, err := filepath.Rel(cwd, path); err != nil {
		panic(err)
	} else {
		return out
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: mv <src-go-file>[ <src-go-file2>...] <dest-go-dir>")
		os.Exit(1)
	}
	var files []fileInfo
	var dest string
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for i, f := range os.Args {
		if i == 0 {
		} else if i == len(os.Args)-1 {
			// dest
			if fi, err := os.Stat(f); err != nil {
				panic(err)
			} else if !fi.IsDir() {
				panic(f + " (destination) must be a directory")
			} else {
				dest = fixPath(cwd, f)
			}
		} else {
			if fi, err := os.Stat(f); err != nil {
				panic(err)
			} else if fi.IsDir() {
				panic(f + " must be a file")
			} else {
				files = append(files, fileInfo{fi: fi, path: fixPath(cwd, f)})
			}
		}
	}
	fmt.Printf("Moving file(s) [")
	for i, fi := range files {
		if i > 0 {
			fmt.Printf(" ,")
		}
		fmt.Printf("%s", fi.path)
	}
	fmt.Printf("] to location [%s]\n", dest)

	key := randomHex(20)
	// first rename xsymbols to something unique
	var symbols []*xsymbol
	fmt.Printf("Renaming symbols\n")
	for _, fi := range files {
		fmt.Printf("  Processing file [%s]\n", fi.path)
		for {
			sym, err := getSymbol(fi.path, key)
			if err != nil {
				panic(err)
			}
			if sym == nil {
				break
			}
			symbols = append(symbols, sym)
			newName := sym.name + "_" + key
			fmt.Printf("    Renaming %s (#%d) %s\n", sym.name, sym.pos, newName)
			cmd("gopls", "rename", "-w", fmt.Sprintf("%s:#%d", fi.path, sym.pos), newName)
			break
		}
	}
	return

	// See if there are any unexported xsymbols which must be exported
	fmt.Printf("Checking for xsymbols to export\n")
	for _, sym := range symbols {
		if isUpper(sym.name) {
			continue
		}
		stdout := cmd("grep", "-nr", sym.name+"_"+key, ".")
		for _, l := range stdout {
			for _, fi := range files {
				if strings.HasPrefix(l, fi.path) {
					// Not interested in xsymbols in the file which we're moving anyway
					continue
				}
			}
		}
	}

	// move the file
	//os.Rename(os.Args[1], os.Args[2])
}
