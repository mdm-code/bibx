/*
Scan package handles the tokenization of BibTeX entries, abbreviations,
preamble elements and comments. The process is divided between two components:

* `scan.Reader`  -- emits one character at a time from the underlying io.Reader
* `scan.Scanner` -- emits one token at a time based on the underlying scan.Reader

Here is an example of how to use `scan.Reader` and `scan.Scanner` to obtain a
series of Item tokens.

Usage
	package main

	import (
		"fmt"
		"os"

		"github.com/mdm-code/bibx/internal/scan"
	)

	func main() {
		s := scan.NewScanner(scan.NewReader(os.Stdin))

		items := []scan.Item{}

		for {
			if i := s.Next(); i.T != scan.ItemErr && i.T != scan.ItemEOF {
				items = append(items, i)
			} else {
				break
			}
		}
		fmt.Println(items)
	}

BibTeX has one major problem: it appears that there is no formal specification
of the language, and all the specs of BibTeX are based are deduced from the
behaviour of the original implementation. Here is a good deductive base lexical
and syntactical grammar specification for BibTeX:

https://metacpan.org/release/AMBS/Text-BibTeX-0.66/view/btparse/doc/bt_language.pod
*/
package scan
