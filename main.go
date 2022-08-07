// SPDX-FileCopyrightText: Copyright 2022 The protobuf-tools Authors
// SPDX-License-Identifier: BSD-3-Clause

// Command protoc-gen-proxy generates RPC service proxy for .pb.go types.
package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/protobuf-tools/protoc-gen-proxy/proxy"
)

func main() {
	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = proxy.SupportedFeatures

		for _, f := range gen.Files {
			if f.Generate {
				proxy.GenerateFile(gen, f)
			}
		}

		return nil
	})
}
