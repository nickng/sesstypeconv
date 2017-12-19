// Copyright 2017 Nicholas Ng <nickng@nickng.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"go.nickng.io/sesstype"
	"go.nickng.io/sesstype/local"
	"go.nickng.io/sesstypeconv"
)

var (
	outfile = flag.String("o", "", "Output file (default: stdout)")

	reader = os.Stdin
	writer = os.Stdout
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sesstype2aut [options] local.mpst\n\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	if len(flag.Args()) > 0 {
		var infile string
		if len(flag.Args()) > 0 {
			infile = flag.Arg(0)
		}
		rdFile, err := os.Open(infile)
		if err != nil {
			log.Fatal(err)
		}
		defer rdFile.Close()
		reader = rdFile
	} else {
		fmt.Fprintf(os.Stderr, "Reading from stdin\n")
	}

	if *outfile != "" {
		wrFile, err := os.OpenFile(*outfile, os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer wrFile.Close()
		writer = wrFile
	}

	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	l, err := local.Parse(tee)
	if err != nil {
		if err, ok := err.(*sesstype.ErrParse); ok {
			diag := err.Pos.CaretDiag(buf.Bytes())
			fmt.Fprintf(os.Stderr, "%v:\n%s", err, diag)
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	a := sesstypeconv.ToAut(l)

	fmt.Fprintf(writer, "%s\n", a.String())
}
