<!--* freshness: { owner: 'kaue' reviewed: '2022-09-26' } *-->

# Text Proto Formatting

go/txtpbfmt

[TOC]

`txtpbfmt` applies a standard formatting to text proto files. It is required for
use by all METADATA files in google3.

This saves development (read/edit/review) time and enables automated edits like
[Inferred Presubmit](http://go/bluze/metadata), analogous to what go/buildifier
does for BUILD files.

See also the
[open source documentation](http://g3doc/third_party/txtpbfmt/README).

[TOC]

## How to format existing text proto files?

```shell
$ /google/bin/releases/text-proto-format/public/fmt <files>
```

Or just

```shell
$ txtpbfmt <files>
```

(install with `$ sudo apt install text-proto-format`)

Please add tag [#text-proto-format](http://cl/#search/&q=tag:text-proto-format)
to your CL(s).

## Which tools support it? How to format on save?

**g4 fix**: runs by default on
[certain file extensions](#which-file-extensions-are-supported-for-my-text-proto-file)
(also when triggered from Cider)

**hg fix**: runs by default on
[certain file extensions](#which-file-extensions-are-supported-for-my-text-proto-file)

**git5 fix**: runs by default on
[certain file extensions](#which-file-extensions-are-supported-for-my-text-proto-file)

**Cider**: available

*   auto-format on save:
    *   open "Settings", "Preferences"
    *   scroll down
    *   enable "Auto-formatting on save"
    *   update "Auto-format whitelist" to the desired extensions (eg "METADATA,
        textpb")
*   one-time for a CL: click on pending CL then "g4 fix"
*   one-time for a file: click on "Tools" then "Format file"

**Vim**: available

*   auto-format on save:

    ```vim
    autocmd FileType textpb AutoFormatBuffer text-proto-format
    ```

**Emacs**: available

*   auto-format on save:

    ```elisp
    (require 'protobuffer)
    (customize-set-variable 'protobuffer-format-before-save t)
    ```

*   one-time for a buffer:

    ```none
    M-x protobuffer-format
    ```

**IntelliJ**: available; the standard reformat file action delegates to txtpbfmt

## How to enforce with a presubmit?

```textproto
presubmit: {
  include_presubmit: "//depot/google3/third_party/txtpbfmt/check_text_proto_format.METADATA"
}
```

Note: formatting of the content of METADATA files is enforced by default
([cs link](http://source/search?q=CheckMETADATAFormat%20f:google3%2FMETADATA$&ssfr=1)).

## Which file extensions are supported for my text proto file?

The `.textproto` extension is the most used extension Google-wide (the second
most popular choice is `.textpb`). See other supported extensions
[here](http://google3/third_party/txtpbfmt/check_text_proto_format.METADATA).

Extensions matching `*ascii*` are not supported because the name 'ascii format'
is deprecated in favor of 'text format' ([source](http://screen/FsWiivZAc9X)).

Finally, the `.pb.txt` extension is discouraged because it is harder to support
(tools parse the extension as being just `.txt`).

## What does it do?

Main features:

![alt_text](https://screenshot.googleplex.com/Be6n9KwWks9pkKR.png "image_tooltip")

([screen](http://screen/Be6n9KwWks9pkKR))

## How to integrate with my tool?

Options:

*   call blade:fmtserver to format
    ([example](http://g/text-proto-format/Avo5CjskrwM/EUnDInTrBQAJ))
*   run the binary from BinFS as a subprocess
    ([examples](http://screen/4XyekZWEaozjLyM))
*   run the binary from a Blaze target as a subprocess
    ([example in Python](http://screen/ZQQRoQHSYTZTNoz), can run on Forge)
*   call the Go library directly
    ([code](http://google3/third_party/txtpbfmt/parser.go?type=cs&q=%22func+Format%28%22))

## Is there an API to edit text proto files while preserving comments?

Yes, see [ast.go](http://google3/third_party/txtpbfmt/ast.go).

For C++ users,
[textformat_patch](http://google3/net/proto2/contrib/textformat_patch/) is built
on top of ast.go.

For Java users,
[textformat_patch](http://google3/java/com/google/protobuf/contrib/textformat/)
is SWIG'd from C++.

For CLI usage, go/textprotoedit is also built on top of ast.go.

## How to disable it?

Please notify g/text-proto-format explaining the problem.

You can disable formatting for a whole file by adding a comment with "#
txtpbfmt: disable" to the top of the file (before the first non-empty
non-comment line), eg:

```textproto
# proto-file: google3/devtools/metadata/metadata.proto
# proto-message: MetaData
# File overview ...

# txtpbfmt: disable

content: { ... }
```

Partial disabling is discussed on b/71542323.

You can disable the presubmit check for a CL using
`DISABLE_CHECK_TEXT_PROTO_FORMAT=<reason>` at the end of the CL description
([example](http://cl/186476404)).

## How to use it locally on Mac?

You can run the blaze command from a google3 workspace

```
blaze build third_party/txtpbfmt:burrito
```

to build `txtpbfmt` into a burrito package. In the artifacts from the build,
find `txtpbfmt.burrito` and install it by running

```
sudo /usr/local/bin/mule install <path_to_dir>/txtpbfmt.burrito
```

Then, you can type `txtpbfmt` in the command line tool and use it.

## How can I make my text proto files better?

You can add the text proto header (go/textformat-schema) to document the source
.proto file and message, and also to enable auto completion in some IDEs.

You can check validity of the text proto when parsing into the appropriate
proto message using go/text-proto-test.

## How to file bugs?

See existing bugs on go/txtpbfmt-bug-list, create a new one using
go/txtpbfmt-bug.

## How to contribute changes?

google3 reviews are preferred to GitHub pull requests.

## How to find more info?

Usage: go/txtpbfmt-usage

Docs / Bugs

*   Text Proto Formatting and Inferred METADATA Presubmits on Eng News
    ([link](http://engdoc/eng/newsletter/g3doc/content/165/newsletter_165_inferred_presubmit))
*   Text Proto Formatting METADATA LSC (go/txtpbfmt-lsc-metadata)
*   Text Proto Formatting All LSC (go/txtpbfmt-lsc-all)
*   Standard Text Proto Format for Piper (go/std-text-proto-format-for-piper)
*   Format all text protos under //video (go/yt-text-proto-format-all-video)
*   Format all google3 METADATA files (b/66905217) and blocking bugs
*   METADATA Refactoring (go/metafactor)

Talks

*   Text Proto Formatting talk (2018 Feb, 5 min, Open Source Farmers Market) on
    [link](https://docs.google.com/presentation/d/1o3qUfKCrP1sRJ9nbZ_ErZ7jK3q0Asu3CH9Vvkk2hRmo/edit#slide=id.p)
*   Text Proto Format talk (2017 Dec, 5 min, YT DevExProd Eng / Product Reviews)
    on
    [link](https://docs.google.com/presentation/d/1TxJNLqQWy7LTFFXlfFujG9TmHQ6JuYCf-6ZbTPCwRL0/edit?hl=en#slide=id.p)

## How to ask questions?

Please ask g/text-proto-format.

## See also

* Text Format Language Specification (go/textformat-spec)
