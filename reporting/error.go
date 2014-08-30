package reporting

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

/*func getErrorCols(lineNo int, span parse.FileSpan) (int, int) {
	start := 0
	end := 0
	if span.Start.Line == lineNo {
		start == span.Start.Col
	}
}*/

func PrintPosAsError(path string, errLine, errCol int) {
	r, err := os.Open(path)
	brdr := bufio.NewReader(r)
	if err != nil {
		fmt.Printf("Cannot print error span %s\n", err)
		return
	}
	lineNo := 0
	for {
		lineNo += 1
		line, err := brdr.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Printf("Cannot print error span %s\n", err)
			return
		}
		if lineNo == errLine {
			line = strings.Replace(line, "\t", "    ", -1)
			//remove newline then readd with println. This makes eof case work properly.
			line = strings.Replace(line, "\n", "", -1)
			fmt.Println(line)
			for i := 0; i < len(line); i++ {
				if i == errCol-1 {
					fmt.Printf("^")
				} else {
					fmt.Printf(" ")
				}
			}
			fmt.Printf("\n")
			break
		}
	}
}
