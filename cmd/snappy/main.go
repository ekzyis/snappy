package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	sn "github.com/ekzyis/snappy"
)

//go:embed banner.txt
var banner string

func usage() {
	fmt.Fprintf(os.Stderr, "%s\n\n", banner)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <command> [-type all|posts|comments]\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  %s query [-author <username>|-territory<territory>]  Query all items of a user or territory.\n\n", filepath.Base(os.Args[0]))
}

func queryUsage(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", banner)
		fmt.Fprintf(os.Stderr, "Query all items of a user or territory.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s query -author <username> [-type all|posts|comments]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s query -territory <territory> [-type all|posts|comments]\n\n", filepath.Base(os.Args[0]))
		fs.PrintDefaults()
	}
}

func runQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = queryUsage(fs)

	authorFlag := fs.String("author", "", "Stacker News username")
	territoryFlag := fs.String("territory", "", "Stacker News territory")
	typeFlag := fs.String("type", "posts", "Items to query: all, posts, or comments")

	fs.Parse(args)

	author := *authorFlag
	territory := *territoryFlag
	type_ := *typeFlag
	sort := "new"

	if author == "" && territory == "" {
		fmt.Fprintln(os.Stderr, "error: -author or -territory is required")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(2)
	}
	if author != "" && territory != "" {
		fmt.Fprintln(os.Stderr, "error: only one of -author and -territory is allowed")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(2)
	}
	if author != "" {
		// must use sort:user for API reasons, this is also why -author and
		// -territory isn't supported
		sort = "user"
	}
	if type_ != "all" && type_ != "posts" && type_ != "comments" {
		fmt.Fprintln(os.Stderr, "error: -type must be all, posts, or comments")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(2)
	}

	progress, close := openTTY()
	defer close()

	client := sn.NewClient()
	var (
		posts   []sn.Item
		cursor  string
		hasMore = true
		limit   = 100
		pageNum = 1
	)
	for hasMore {
		fmt.Fprintf(progress, "Fetching page %d...\n", pageNum)
		q := &sn.ItemsQuery{
			Sort:   sort,
			Sub:    territory,
			Name:   author,
			Type:   type_,
			By:     "new",
			When:   "forever",
			Cursor: cursor,
			Limit:  limit,
		}
		page, err := client.Items(q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Items: %v\n", err)
			os.Exit(1)
		}
		posts = append(posts, page.Items...)
		cursor = page.Cursor
		hasMore = cursor != ""
		pageNum++
	}

	fmt.Fprintln(progress, "Done.")

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(posts); err != nil {
		fmt.Fprintf(os.Stderr, "error: encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func openTTY() (w io.Writer, close func()) {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return io.Discard, func() {}
	}
	return tty, func() { tty.Close() }
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage()
		os.Exit(0)
	case "query":
		runQuery(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", args[0])
		usage()
		os.Exit(2)
	}
}
