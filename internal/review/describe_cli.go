package review

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tomasz-tomczyk/crit/internal/clicmd"
)

const describeUsage = `Usage: crit describe [options] [--title <text>] [--body <text>|--file <path>|-]
       crit describe --clear

Set the review header shown above the file list: a one-line title and a
markdown description (a summary of the CR or PR under review).

Options:
      --title <text>   One-line review title
      --body <text>    Markdown description
      --file <path>    Read the description from a file ("-" for stdin)
      --clear          Remove the title and description
      --session <id>   Target a specific review session
  -o, --output <dir>   Crit data root for reviews`

// MaxDescribeBodyBytes caps the description so a runaway agent cannot write a
// multi-megabyte blob into the review file, which the browser renders in full.
const MaxDescribeBodyBytes = 64 * 1024

type describeArgs struct {
	title, body, file, outputDir, sessionID string
	hasTitle, hasBody                       bool
	clear, stdin                            bool
}

var describeValueFlags = map[string]bool{
	"--title": true, "--body": true, "--file": true, "-o": true, "--output": true, "--session": true,
}

func parseDescribeArgs(args []string) (describeArgs, error) {
	var a describeArgs
	for i := 0; i < len(args); i++ {
		flag := args[i]
		value := ""
		if describeValueFlags[flag] {
			if i+1 >= len(args) {
				return a, describeUsageError(flag + " requires a value")
			}
			i++
			value = args[i]
		}
		switch flag {
		case "--title":
			a.title, a.hasTitle = value, true
		case "--body":
			a.body, a.hasBody = value, true
		case "--file":
			a.file = value
		case "-o", "--output":
			a.outputDir = value
		case "--session":
			a.sessionID = value
		case "--clear":
			a.clear = true
		case "-":
			a.stdin = true
		default:
			return a, describeUsageError("unknown argument " + flag)
		}
	}
	// Not `return a, a.validate()`: validate normalizes `--file -` into stdin,
	// and the evaluation order of a non-call operand against a call is not
	// specified, so the caller could get the un-normalized copy.
	err := a.validate()
	return a, err
}

func (a *describeArgs) validate() error {
	if a.file == "-" {
		a.stdin = true
		a.file = ""
	}
	sets := a.hasTitle || a.hasBody || a.file != "" || a.stdin
	if a.clear && sets {
		return describeUsageError("--clear cannot be combined with other options")
	}
	if !a.clear && !sets {
		return describeUsageError("nothing to set")
	}
	return nil
}

// resolveBody folds --file / stdin into the body, then trims and size-checks it.
func (a *describeArgs) resolveBody() error {
	switch {
	case a.file != "":
		data, err := os.ReadFile(a.file)
		if err != nil {
			return err
		}
		a.body, a.hasBody = string(data), true
	case a.stdin:
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		a.body, a.hasBody = string(data), true
	}

	a.body = strings.TrimSpace(a.body)
	a.title = strings.TrimSpace(a.title)
	if len(a.body) > MaxDescribeBodyBytes {
		return fmt.Errorf("description is %d bytes; the limit is %d", len(a.body), MaxDescribeBodyBytes)
	}
	return nil
}

func RunDescribe(args []string) error {
	a, err := parseDescribeArgs(args)
	if err != nil {
		return err
	}
	if err := a.resolveBody(); err != nil {
		return err
	}

	critPath, err := ResolveCommandReviewPathWithSession(a.sessionID, a.outputDir, "")
	if err != nil {
		return err
	}
	cj, err := LoadCritJSON(critPath)
	if err != nil {
		return err
	}
	switch {
	case a.clear:
		cj.Title, cj.Description = "", ""
	default:
		if a.hasTitle {
			cj.Title = a.title
		}
		if a.hasBody {
			cj.Description = a.body
		}
	}
	if err := SaveCritJSON(critPath, cj); err != nil {
		return err
	}

	if a.clear {
		fmt.Println("Cleared the review title and description.")
		return nil
	}
	if cj.Title != "" {
		fmt.Printf("Review title: %s\n", cj.Title)
	}
	if cj.Description != "" {
		fmt.Printf("Description: %d characters.\n", len([]rune(cj.Description)))
	}
	return nil
}

func describeUsageError(reason string) error {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", reason, describeUsage)
	return clicmd.ExitError{Code: 1, Err: errors.New("exit")}
}
