// SPDX-FileCopyrightText: Copyright 2022 The protobuf-tools Authors
// SPDX-License-Identifier: BSD-3-Clause

// Command protoc-gen-grpcproxy generates RPC service proxy for .pb.go types.
package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/protobuf-tools/protoc-gen-grpcproxy/proxy"
)

func main() {
	cfg := &proxy.Config{}

	flags := flag.NewFlagSet("protoc-gen-grpcproxy", flag.ExitOnError)
	flags.BoolVar(&cfg.Standalone, "standalone", false, "standalone mode.")

	opts := protogen.Options{
		ParamFunc: flags.Set,
	}
	pluginFn := func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if f.Desc.Syntax() != protoreflect.Proto3 {
				continue
			}
			if f.Generate {
				proxy.GenerateFile(gen, f, cfg)
			}
		}
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) | uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
		gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
		gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023

		return nil
	}

	opts.Run(pluginFn)
}
