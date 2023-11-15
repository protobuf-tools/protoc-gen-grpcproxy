// SPDX-FileCopyrightText: Copyright 2022 The protobuf-tools Authors
// SPDX-License-Identifier: BSD-3-Clause

// Command protoc-gen-grpcproxy generates RPC service proxy for .pb.go types.
package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/protobuf-tools/protoc-gen-grpcproxy/proxy"
)

func main() {
	cfg := &proxy.Config{}

	flags := flag.NewFlagSet("protoc-gen-grpcproxy", flag.ExitOnError)
	flags.BoolVar(&cfg.Standalone, "standalone", false, "standalone mode.")

	opts := protogen.Options{
		ParamFunc: flags.Set,
	}
	pluginFn := func(p *protogen.Plugin) error {
		for _, f := range p.Files {
			if f.Generate {
				proxy.GenerateFile(p, f, cfg)
			}
		}

		return nil
	}

	opts.Run(pluginFn)
}
