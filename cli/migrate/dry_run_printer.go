package migrate

import (
	"bytes"
	"text/tabwriter"

	"github.com/hasura/graphql-engine/cli/util"
)

func getDryRunPrinter(migration <-chan Migration) (func(<-chan Migration), *bytes.Buffer) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)
	w := util.NewPrefixWriter(out)
	w.Write(util.LEVEL_0, "VERSION\tNAME\tDIRECTION\n")

	var printer = func(migration <-chan Migration) {
		for m := range migration {
			w.Write(util.LEVEL_0, "%d\t%s\t%s\t%s\n",
				m.Version,
				m.FileName,
				m.Identifier,
			)

		}
	}

	return printer, buf

}
