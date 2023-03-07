// The fmt command applies standard formatting to text proto files and preserves
// comments.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"flag"
	// Google internal base/go package, commented out by copybara
	log "github.com/golang/glog"
	"github.com/protocolbuffers/txtpbfmt/parser"
)

var (
	// Top level flags.
	dryRun                   = flag.Bool("dry_run", false, "Enable dry run mode.")
	expandAllChildren        = flag.Bool("expand_all_children", false, "Expand all children irrespective of initial state.")
	skipAllColons            = flag.Bool("skip_all_colons", false, "Skip colons whenever possible.")
	allowTripleQuotedStrings = flag.Bool("allow_triple_quoted_strings", false,
		`Allow Python-style """ or ''' delimited strings in input.`)
	stdinDisplayPath = flag.String("stdin_display_path", "<stdin>",
		"The path to display when referring to the content read from stdin.")
)

const stdinPlaceholderPath = "<stdin>"

func read(path string) ([]byte, error) {
	if path == stdinPlaceholderPath {
		return ioutil.ReadAll(bufio.NewReader(os.Stdin))
	}
	return ioutil.ReadFile(path)
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func contentForLogging(content []byte) string {
	res := string(content)
	if len(res) > 100 {
		res = res[:100] + " ... <snip> ..."
	}
	return res
}

func main() {
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		paths = append(paths, stdinPlaceholderPath)
	}
	log.Info("paths: ", paths)
	errs := 0
	for _, path := range paths {
		if strings.HasPrefix(path, "//depot/google3/") {
			path = strings.Replace(path, "//depot/google3/", "", 1)
		}
		displayPath := path
		if path == stdinPlaceholderPath {
			displayPath = *stdinDisplayPath
			log.Info("path ", path, " displayed as ", displayPath)
		} else {
			log.Info("path ", path)
		}

		content, err := read(path)
		if os.IsNotExist(err) {
			log.Error("Ignoring path: ", err)
			errs++
			continue
		} else if err != nil {
			log.Exit(err)
		}

		// Only pass the verbose logger if its level is enabled.
		var logger parser.Logger
		if l := log.V(2); l {
			logger = l
		}
		newContent, err := parser.FormatWithConfig(content, parser.Config{
			ExpandAllChildren:        *expandAllChildren,
			SkipAllColons:            *skipAllColons,
			AllowTripleQuotedStrings: *allowTripleQuotedStrings,
			Logger:                   logger,
		})
		if err != nil {
			errorf("parser.Format for path %v with content %q returned err %v", displayPath, contentForLogging(content), err)
			errs++
			continue
		}
		log.V(2).Infof("New content for path %s: %q", displayPath, newContent)

		if path == stdinPlaceholderPath {
			fmt.Print(string(newContent))
			continue
		}
		if bytes.Equal(content, newContent) {
			log.Info("No change for path ", displayPath)
			continue
		}
		if *dryRun {
			fmt.Println(string(newContent))
			continue
		}
		if err := ioutil.WriteFile(path, newContent, 0664); err != nil {
			log.Exit(err)
		}
	}
	if errs > 0 {
		log.Exit(errs, " error(s) encountered during execution")
	}
}
