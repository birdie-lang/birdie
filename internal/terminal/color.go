package terminal

import (
	"io"
	"os"
)

func ColorEnabled(out io.Writer) bool {
	file, ok := out.(*os.File)
	return ok && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" && IsTerminal(file)
}
