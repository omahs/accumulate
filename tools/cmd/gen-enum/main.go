// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/accumulatenetwork/accumulate/tools/internal/typegen"
)

var flags struct {
	files typegen.FileReader

	Package    string
	Language   string
	Out        string
	ShortNames bool
}

func main() {
	cmd := cobra.Command{
		Use:  "gen-enum [file]",
		Args: cobra.MinimumNArgs(1),
		Run:  run,
	}

	cmd.Flags().StringVarP(&flags.Language, "language", "l", "Go", "Output language or template file")
	cmd.Flags().StringVar(&flags.Package, "package", "protocol", "Package name")
	cmd.Flags().BoolVar(&flags.ShortNames, "short-names", false, "Omit the type name from the enum value")
	cmd.Flags().StringVarP(&flags.Out, "out", "o", "enums_gen.go", "Output file")
	flags.files.SetFlags(cmd.Flags(), "enums")

	_ = cmd.Execute()
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		fatalf("%v", err)
	}
}

func run(_ *cobra.Command, args []string) {
	types, err := typegen.ReadMap[typegen.Enum](&flags.files, args, nil)
	check(err)
	ttypes := convert(types, flags.Package)

	w := new(bytes.Buffer)
	check(Templates.Execute(w, flags.Language, ttypes))
	check(typegen.WriteFile(flags.Out, w))
}
