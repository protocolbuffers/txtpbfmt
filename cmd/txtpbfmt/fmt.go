// The fmt command applies standard formatting to text proto files and preserves
// comments.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"flag"
	// Google internal base/go package, commented out by copybara
	log "github.com/golang/glog"
	"github.com/protocolbuffers/txtpbfmt/config"
	"github.com/protocolbuffers/txtpbfmt/logger"
	"github.com/protocolbuffers/txtpbfmt/parser"
)

var (
	// Top level flags.
	dryRun                                 = flag.Bool("dry_run", false, "Enable dry run mode.")
	expandAllChildren                      = flag.Bool("expand_all_children", false, "Expand all children irrespective of initial state.")
	skipAllColons                          = flag.Bool("skip_all_colons", false, "Skip colons whenever possible.")
	sortFieldsByFieldName                  = flag.Bool("sort_fields_by_field_name", false, "Sort fields by field name.")
	sortRepeatedFieldsByContent            = flag.Bool("sort_repeated_fields_by_content", false, "Sort adjacent scalar fields of the same field name by their contents.")
	sortRepeatedFieldsBySubfield           = flag.String("sort_repeated_fields_by_subfield", "", "Sort adjacent message fields of the given field name by the contents of the given subfield.")
	removeDuplicateValuesForRepeatedFields = flag.Bool("remove_duplicate_values_for_repeated_fields", false, "Remove lines that have the same field name and scalar value as another.")
	allowTripleQuotedStrings               = flag.Bool("allow_triple_quoted_strings", false, `Allow Python-style """ or ''' delimited strings in input.`)
	stdinDisplayPath                       = flag.String("stdin_display_path", "<stdin>", "The path to display when referring to the content read from stdin.")
	wrapStringsAtColumn                    = flag.Int("wrap_strings_at_column", 0, "Max columns for string field values. (0 means no wrap.)")
	wrapHTMLStrings                        = flag.Bool("wrap_html_strings", false, "Wrap strings containing HTML tags. (Requires wrap_strings_at_column > 0.)")
	wrapStringsAfterNewlines               = flag.Bool("wrap_strings_after_newlines", false, "Wrap strings after newlines.")
	wrapStringsWithoutWordwrap             = flag.Bool("wrap_strings_without_wordwrap", false, "Wrap strings at the given column only.")
	preserveAngleBrackets                  = flag.Bool("preserve_angle_brackets", false, "Preserve angle brackets instead of converting to curly braces.")
	smartQuotes                            = flag.Bool("smart_quotes", false, "Use single quotes around strings that contain double but not single quotes.")
)

const stdinPlaceholderPath = "<stdin>"

func read(path string) ([]byte, error) {
	if path == stdinPlaceholderPath {
		return io.ReadAll(bufio.NewReader(os.Stdin))
	}
	return os.ReadFile(path)
}

func errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func contentForLogging(content []byte) string {
	res := string(content)
	if len(res) > 100 {
		res = res[:100] + " ... <snip> ..."
	}
	return res
}

func processPath(path string) error {
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
		return fmt.Errorf("path not found")
	}
	if err != nil {
		return err
	}

	// Only pass the verbose logger if its level is enabled.
	var logger logger.Logger
	if l := log.V(2); l {
		logger = l
	}
	newContent, err := parser.FormatWithConfig(content, config.Config{
		ExpandAllChildren:                      *expandAllChildren,
		SkipAllColons:                          *skipAllColons,
		SortFieldsByFieldName:                  *sortFieldsByFieldName,
		SortRepeatedFieldsByContent:            *sortRepeatedFieldsByContent,
		SortRepeatedFieldsBySubfield:           strings.Split(*sortRepeatedFieldsBySubfield, ","),
		RemoveDuplicateValuesForRepeatedFields: *removeDuplicateValuesForRepeatedFields,
		AllowTripleQuotedStrings:               *allowTripleQuotedStrings,
		WrapStringsAtColumn:                    *wrapStringsAtColumn,
		WrapHTMLStrings:                        *wrapHTMLStrings,
		WrapStringsAfterNewlines:               *wrapStringsAfterNewlines,
		WrapStringsWithoutWordwrap:             *wrapStringsWithoutWordwrap,
		PreserveAngleBrackets:                  *preserveAngleBrackets,
		SmartQuotes:                            *smartQuotes,
		Logger:                                 logger,
	})
	if err != nil {
		errorf("parser.Format for path %v with content %q returned err %v", displayPath, contentForLogging(content), err)
		return fmt.Errorf("parser.Format failed")
	}
	log.V(2).Infof("New content for path %s: %q", displayPath, newContent)

	return write(path, content, newContent)
}

func write(path string, content, newContent []byte) error {
	if path == stdinPlaceholderPath {
		fmt.Print(string(newContent))
		return nil
	}
	if bytes.Equal(content, newContent) {
		log.Info("No change for path ", path)
		return nil
	}
	if *dryRun {
		fmt.Println(string(newContent))
		return nil
	}
	if err := os.WriteFile(path, newContent, 0664); err != nil {
		return err
	}
	return nil
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
		if err := processPath(path); err != nil {
			if err.Error() == "path not found" || err.Error() == "parser.Format failed" {
				errs++
				continue
			}
			log.Exit(err)
		}
	}
	if errs > 0 {
		log.Exit(errs, " error(s) encountered during execution")
	}
}
