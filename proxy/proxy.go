// SPDX-FileCopyrightText: Copyright 2022 The protobuf-tools Authors
// SPDX-License-Identifier: BSD-3-Clause

// Package proxy generates RPC service proxy.
package proxy

import (
	"fmt"
	"sort"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

// list of _proxy.pb.go files package dependencies.
const (
	contextPackage = protogen.GoImportPath("context")
	errorsPackage  = protogen.GoImportPath("errors")
	netPackage     = protogen.GoImportPath("net")

	emptyPackage = protogen.GoImportPath("google.golang.org/protobuf/types/known/emptypb")
	grpcPackage  = protogen.GoImportPath("google.golang.org/grpc")
)

// FileNameSuffix is the suffix added to files generated by deepcopy
const FileNameSuffix = "_proxy.pb.go"

// Config represents a protoc-gen-proxy config.
type Config struct {
	Standalone bool
	OutGopath  string
}

// GenerateFile generates DeepCopyInto() and DeepCopy() functions for .pb.go types.
func GenerateFile(p *protogen.Plugin, f *protogen.File, cfg *Config) *protogen.GeneratedFile {
	if len(f.Services) == 0 {
		return nil
	}

	filename := f.GeneratedFilenamePrefix + FileNameSuffix
	goImportPath := f.GoImportPath

	g := p.NewGeneratedFile(filename, goImportPath)

	g.QualifiedGoIdent(contextPackage.Ident(""))
	g.QualifiedGoIdent(errorsPackage.Ident(""))
	g.QualifiedGoIdent(netPackage.Ident(""))
	g.QualifiedGoIdent(grpcPackage.Ident(""))
	g.QualifiedGoIdent(emptyPackage.Ident(""))

	g.P(`// Code generated by protoc-gen-proxy. DO NOT EDIT.`)
	g.P()
	g.P(`package `, f.GoPackageName)
	g.P()

	g.P(`var _ `, emptyPackage.Ident("Empty"))
	for _, service := range f.Services {
		serviceName := service.GoName
		lowerServiceName := strcase.ToLowerCamel(serviceName)
		serverName := lowerServiceName + "Server"

		methods := make(map[string]string)
		for _, method := range service.Methods {
			input := method.Input.GoIdent.GoName
			output := method.Output.GoIdent.GoName
			if input == "Empty" {
				input = "emptypb.Empty"
			}
			if output == "Empty" {
				output = "emptypb.Empty"
			}
			methods[method.GoName] = fmt.Sprintf(`(ctx context.Context, req *%s) (*%s, error)`, input, output)
		}

		sortMethods := make([]string, len(methods))
		i := 0
		for fn := range methods {
			sortMethods[i] = fn
			i++
		}
		sort.Strings(sortMethods)

		g.P(`// Proxy allows to create `, serviceName, ` proxy servers.`)
		g.P(`type Proxy struct {`)
		for _, fn := range sortMethods {
			args := methods[fn]
			g.P(`	`, fn, ` func`, args)
		}
		g.P(`}`)
		g.P()
		g.P(`// Serve starts serving the proxy server on the given listener with the specified options.`)
		g.P(`func (p *Proxy) Serve(l net.Listener, opts ...grpc.ServerOption) error {`)
		g.P(`	srv := grpc.NewServer(opts...)`)
		g.P(`	Register`, serviceName, `Server(srv, &`, serverName, `{proxy: p})`)
		g.P()
		g.P(`	return srv.Serve(l)`)
		g.P(`}`)
		g.P()
		g.P(`var errNotSupported = errors.New("operation not supported")`)
		g.P()
		g.P(`type `, serverName, ` struct {`)
		g.P(`	proxy *Proxy`)
		g.P(`}`)
		g.P()
		for _, fn := range sortMethods {
			args := methods[fn]

			g.P(`func (s *`, serverName, `) `, fn, args, ` {`)
			g.P(`	fn := s.proxy.`, fn)
			g.P(` 	if fn == nil {`)
			g.P(` 		return nil, errNotSupported`)
			g.P(` 	}`)
			g.P()
			g.P(`return fn(ctx, req)`)
			g.P(`}`)
			g.P()
		}
	}

	return g
}
