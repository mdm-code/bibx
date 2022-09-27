package main

import (
	"fmt"
	"os"

	"github.com/mdm-code/bibx/internal/parse"
	"github.com/mdm-code/bibx/internal/scan"
)

func main() {
	s := scan.NewScanner(scan.NewReader(os.Stdin))
	p := parse.NewParser(s)

	n, ok := p.Next()
	for ok {
		switch decl := n.(type) {
		case *parse.EntryDecl:
			fmt.Printf("Type: %s\n", decl)
			fmt.Printf("Cite key: %s\n", decl.CiteKey)
			fmt.Println("Comments:")
			for i, c := range decl.Comments.Values {
				fmt.Printf("%d: %s\n", i, c.Value)
			}
			fmt.Println("Fields:")
			for _, f := range decl.Fields {
				fmt.Printf("%s = %s\n", f.Key, f.Value)
			}
			fmt.Println()
		case *parse.PreambleDecl:
			fmt.Printf("Type: %s\n", decl)
			fmt.Println("Comments:")
			for i, c := range decl.Comments.Values {
				fmt.Printf("%d: %s\n", i, c.Value)
			}
			fmt.Println("Value:")
			fmt.Println(decl.Value)
		case *parse.AbbrevDecl:
			fmt.Printf("Type: %s\n", decl)
			fmt.Println("Comments:")
			for i, c := range decl.Comments.Values {
				fmt.Printf("%d: %s\n", i, c.Value)
			}
			fmt.Println("Field:")
			fmt.Printf("%s = %s\n", decl.Field.Key, decl.Field.Value)
		default:
			fmt.Println(decl)
		}
		n, ok = p.Next()
	}
}
