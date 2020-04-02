package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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

type export struct {
	name string
	pos  token.Pos
}

func getExport(file, key string) (*export, error) {
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
						if isUpper(n.Name) {
							//fmt.Printf("var/const %v %v\n", n.Name, n.Pos())
							if !strings.Contains(n.Name, key) {
								return &export{name: n.Name, pos: n.Pos()}, nil
							}
						} else {
							//fmt.Printf("  lower %v\n", n.Name)
						}
					}
				} else if ts, ok := s.(*ast.TypeSpec); ok {
					//fmt.Printf("type %v %v\n", ts.Name, ts.Pos())
					if !strings.Contains(ts.Name.String(), key) {
						return &export{name: ts.Name.String(), pos: ts.Pos()}, nil
					}
				} else {
					fmt.Printf("Unrecognized spec: %v\n", s)
				}
			}
		} else if fd, ok := d.(*ast.FuncDecl); ok {
			if fd.Name != nil && isUpper(fd.Name.Name) {
				//fmt.Printf("func %v %v\n", fd.Name, fd.Pos())
				if fd.Recv != nil {
					// This is a method, we don't want to rename it because it should
					// not be prefixed with the new package name.
					continue
				}
				if !strings.Contains(fd.Name.String(), key) {
					return &export{name: fd.Name.String(), pos: fd.Name.Pos()}, nil
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

func main() {
	file := "./wire/protocol.go"
	key := randomHex(20)
	for {
		export, err := getExport("./wire/protocol.go", key)
		if err != nil {
			panic(err)
		}
		if export == nil {
			return
		}
		newName := export.name + "_" + key
		fmt.Printf("Renaming %s (#%d) %s\n", export.name, export.pos, newName)
		cmd := exec.Command("gopls", "rename", "-w", fmt.Sprintf("%s:#%d", file, export.pos), newName)
		//cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
}
