package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	yomitan "github.com/yomidevs/yomitan-import"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] input-path output-path\n", path.Base(os.Args[0]))
	fmt.Fprint(os.Stderr, "https://github.com/yomidevs/yomitan-import/\n\n")
	fmt.Fprint(os.Stderr, "Parameters:\n")
	flag.PrintDefaults()
}

func main() {
	var (
		format   = flag.String("format", yomitan.DefaultFormat, "dictionary format [edict|enamdict|epwing|kanjidic|rikai]")
		language = flag.String("language", yomitan.DefaultLanguage, "dictionary language (if supported)")
		title    = flag.String("title", yomitan.DefaultTitle, "dictionary title")
		stride   = flag.Int("stride", yomitan.DefaultStride, "dictionary bank stride")
		pretty   = flag.Bool("pretty", yomitan.DefaultPretty, "output prettified dictionary JSON")
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	if err := yomitan.ExportDb(flag.Arg(0), flag.Arg(1), *format, *language, *title, *stride, *pretty); err != nil {
		log.Fatal(err)
	}
}
