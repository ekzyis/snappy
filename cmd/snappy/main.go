package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sn "github.com/ekzyis/snappy"
)

//go:embed banner.txt
var banner string

func usage() {
	fmt.Fprintf(os.Stderr, "%s\n\n", banner)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <command> [options]\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  %s query -author <username>      Query all items of a user.\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s query -territory <territory>  Query all items of a territory.\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s query -item <id>              Query a single item by id.\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s delete -item <id>             Delete a single item by id.\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  -type string\n")
	fmt.Fprintf(os.Stderr, "      Items to query: all, posts, or comments (default: posts)\n")
	fmt.Fprintf(os.Stderr, "  -limit int\n")
	fmt.Fprintf(os.Stderr, "      how many items to fetch (default: 100)\n")
}

func queryUsage(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", banner)
		fmt.Fprintf(os.Stderr, "Query a single item or items of a user or territory.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s query -author <username> [options]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s query -territory <territory> [options]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s query -item <id>\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -type string\n")
		fmt.Fprintf(os.Stderr, "      Items to query: all, posts, or comments (default: posts)\n")
		fmt.Fprintf(os.Stderr, "  -limit int\n")
		fmt.Fprintf(os.Stderr, "      how many items to fetch (default: 100)\n")
		fmt.Fprintf(os.Stderr, "  -before string\n")
		fmt.Fprintf(os.Stderr, "      only items created before this date, YYYY-MM-DD or RFC3339\n")
		fmt.Fprintf(os.Stderr, "  -after string\n")
		fmt.Fprintf(os.Stderr, "      only items created on or after this date, YYYY-MM-DD or RFC3339\n")
	}
}

func runQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = queryUsage(fs)

	authorFlag := fs.String("author", "", "Stacker News username")
	territoryFlag := fs.String("territory", "", "Stacker News territory")
	itemFlag := fs.Int("item", 0, "Stacker News item id")
	typeFlag := fs.String("type", "posts", "Items to query: all, posts, or comments")
	limitFlag := fs.Int("limit", 100, "")
	afterFlag := fs.String("after", "", "only items created on or after this date (YYYY-MM-DD or RFC3339)")
	beforeFlag := fs.String("before", "", "only items created before this date (YYYY-MM-DD or RFC3339)")

	fs.Parse(args)

	author := *authorFlag
	territory := *territoryFlag
	item := *itemFlag
	type_ := *typeFlag
	limit := *limitFlag
	after := *afterFlag
	before := *beforeFlag
	sort := "new"
	when := "forever"
	var from, to string

	if after != "" || before != "" {
		if territory != "" {
			fmt.Fprint(os.Stderr, "error: -before and -after cannot be used with -territory\n\n")
			fs.Usage()
			os.Exit(2)
		}
		if author == "" {
			fmt.Fprint(os.Stderr, "error: -before and -after require -author\n\n")
			fs.Usage()
			os.Exit(2)
		}
	}

	// TODO: I wanted to refactor this to check exclusive arguments in one place, but Golang does not
	// have an XOR operator ...
	if item != 0 {
		if author != "" || territory != "" {
			fmt.Fprintf(os.Stderr, "error: only one of -author, -territory and -item is allowed\n\n")
			fs.Usage()
			os.Exit(2)
		}
		runQueryItem(item)
		return
	}

	if author == "" && territory == "" {
		fmt.Fprint(os.Stderr, "error: -author or -territory is required\n\n")
		fs.Usage()
		os.Exit(2)
	}
	if author != "" && territory != "" {
		fmt.Fprint(os.Stderr, "error: only one of -author and -territory is allowed\n\n")
		fs.Usage()
		os.Exit(2)
	}
	if author != "" {
		// must use sort:user for API reasons, this is also why -author and
		// -territory isn't supported
		sort = "user"
	}
	if before != "" {
		t, err := parseDate(before)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: -before: %v\n\n", err)
			fs.Usage()
			os.Exit(2)
		}
		to = strconv.FormatInt(t.UnixMilli(), 10)
	}
	if after != "" {
		t, err := parseDate(after)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: -after: %v\n\n", err)
			fs.Usage()
			os.Exit(2)
		}
		from = strconv.FormatInt(t.UnixMilli(), 10)
	}
	if from != "" || to != "" {
		when = "custom"
	}
	if type_ != "all" && type_ != "posts" && type_ != "comments" {
		fmt.Fprint(os.Stderr, "error: -type must be all, posts, or comments\n\n")
		fs.Usage()
		os.Exit(2)
	}

	progress, close := openTTY()
	defer close()

	client := sn.NewClient()
	var (
		posts     []sn.Item
		cursor    string
		hasMore   = true
		// fetch max 100 items at once
		pageLimit = min(limit, 100)
		count     = 0
		pageNum   = 1
	)
	for hasMore {
		fmt.Fprintf(progress, "Fetching page %d...\n", pageNum)
		q := &sn.ItemsQuery{
			Sort:   sort,
			Sub:    territory,
			Name:   author,
			Type:   type_,
			By:     "new",
			When:   when,
			From:   from,
			To:     to,
			Cursor: cursor,
			// fetch max as many items as we still need
			Limit:  min(limit-count, pageLimit),
		}
		page, err := client.Items(q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Items: %v\n", err)
			os.Exit(1)
		}
		posts = append(posts, page.Items...)
		count = len(posts)
		cursor = page.Cursor
		hasMore = cursor != ""
		pageNum++
		if count >= limit {
			break
		}
	}

	fmt.Fprintln(progress, "Done.")

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(posts); err != nil {
		fmt.Fprintf(os.Stderr, "error: encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func runQueryItem(id int) {
	progress, close := openTTY()
	defer close()

	fmt.Fprintf(progress, "Fetching item %d...\n", id)

	client := sn.NewClient()
	item, err := client.Item(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Item: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(progress, "Done.")

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		fmt.Fprintf(os.Stderr, "error: encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func deleteUsage(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", banner)
		fmt.Fprintf(os.Stderr, "Delete a single item by id. The item is shown before deletion.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s delete -item <id> [options]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -f\n")
		fmt.Fprintf(os.Stderr, "      skip the confirmation prompt\n")
	}
}

func runDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = deleteUsage(fs)

	itemFlag := fs.Int("item", 0, "Stacker News item id")
	forceFlag := fs.Bool("f", false, "skip the confirmation prompt")

	fs.Parse(args)

	id := *itemFlag
	if id == 0 {
		fmt.Fprint(os.Stderr, "error: -item is required\n\n")
		fs.Usage()
		os.Exit(2)
	}

	progress, close := openTTY()
	defer close()

	client := sn.NewClient()

	if client.ApiKey == "" && client.Nsec == "" {
		fmt.Fprintln(os.Stderr, "authentication required: did you set SN_NSEC or SN_API_KEY in your environment?")
		os.Exit(1)
	}

	me, err := client.Me()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Me: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("logged in as %s\n", me.Name)

	fmt.Fprintf(progress, "Fetching item %d...\n", id)
	item, err := client.Item(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Item: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s / %d sats / %d replies\n", item.User.Name, item.Sats, item.NComments)
	fmt.Printf("%s\n\n", truncate(item.Text, 280))

	if !*forceFlag && !confirm(fmt.Sprintf("Delete item %d? [y/N] ", id)) {
		fmt.Fprintln(os.Stderr, "Aborted.")
		os.Exit(1)
	}

	fmt.Fprintf(progress, "Deleting item %d...\n", id)
	if _, err := client.DeleteItem(id); err != nil {
		fmt.Fprintf(os.Stderr, "error: DeleteItem: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(progress, "Deleted item %d.\n", id)
}

// parseDate accepts a plain YYYY-MM-DD date or a full RFC3339 timestamp.
func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD or RFC3339", s)
}

// truncate shortens s to at most max runes, noting how many were cut off.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return fmt.Sprintf("%s\n[truncated with %d chars left]", string(r[:max]), len(r)-max)
}

// confirm prompts on stderr and reads a yes/no answer from stdin, defaulting to
// no on anything other than "y"/"yes".
func confirm(prompt string) bool {
	fmt.Fprint(os.Stderr, prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
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
	case "delete":
		runDelete(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", args[0])
		usage()
		os.Exit(2)
	}
}
