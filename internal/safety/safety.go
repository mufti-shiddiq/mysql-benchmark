package safety

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ConfirmNonEmpty(in io.Reader, out io.Writer, database string, tableCount int, force bool) (bool, error) {
	if tableCount == 0 || force {
		return true, nil
	}
	_, _ = fmt.Fprintf(out, "\nWARNING\n\nDatabase %q is not empty.\nDetected tables: %d\n\nBenchmark setup may drop and recreate benchmark-owned tables.\nContinue? [y/N]: ", database, tableCount)
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}
