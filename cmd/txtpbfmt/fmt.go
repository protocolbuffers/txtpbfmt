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
	log "github.com/golang/glog"
	"github.com/protocolbuffers/txtpbfmt/parser"
)

var (
	// Top level flags.
	dryRun            = flag.Bool("dry_run", false, "Enable dry run mode.")
	expandAllChildren = flag.Bool("expand_all_children", false, "Expand all children irrespective of initial state.")
	skipAllColons     = flag.Bool("skip_all_colons", false, "Skip colons whenever possible.")
)

const stdinPath = "<stdin>"

func read(path string) ([]byte, error) {
	if path == stdinPath {
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
		paths = append(paths, stdinPath)
	}
	log.Info("paths: ", paths)
	errs := 0
	for _, path := range paths {
		if strings.HasPrefix(path, "//depot/google3/") {
			path = strings.Replace(path, "//depot/google3/", "", 1)
		}
		log.Info("path ", path)
		content, err := read(path)
		if os.IsNotExist(err) {
			log.Error("Ignoring path: ", err)
			errs++
			continue
		} else if err != nil {
			log.Exit(err)
		}

		newContent, err := parser.FormatWithConfig(content, parser.Config{
			ExpandAllChildren: *expandAllChildren,
			SkipAllColons:     *skipAllColons,
		})
		if err != nil {
			errorf("parser.Format for path %v with content %q returned err %v", path, contentForLogging(content), err)
			errs++
			continue
		}
		log.V(2).Infof("New content for path %s: %q", path, newContent)

		if path == stdinPath {
			fmt.Print(string(newContent))
			continue
		}
		if bytes.Equal(content, newContent) {
			log.Info("No change for path ", path)
			continue
		}
		if *dryRun {
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
