package main

import (
	"fmt"
	"os"
	"regexp"
)

type ReReplace struct {
	Regex *regexp.Regexp
	Text  string
}

func NewReReplace(regex string, replace string) ReReplace {
	return ReReplace{
		Regex: regexp.MustCompile(regex),
		Text:  replace,
	}
}

func NewReDate(dateStr string) ReReplace {
	return ReReplace{
		Regex: regexp.MustCompile(fmt.Sprintf(`%s: string`, dateStr)),
		Text:  fmt.Sprintf(`%s: Date | string`, dateStr),
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide a file path as an argument")
		os.Exit(1)
	}

	path := os.Args[1]

	fmt.Printf("Processing %s\n", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", path)
		os.Exit(1)
	}

	text := "/* post-processed by ./scripts/process-types.go */\n"
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	text += string(data)

	replaces := [...]ReReplace{
		NewReReplace(` Repo`, " "),
		NewReReplace(` PaginationResultRepo`, " PaginationResult"),
		NewReReplace(` Services`, " "),
		NewReReplace(` V1`, " "),
		NewReReplace(`\?:`, ":"),
		NewReReplace(`(\w+):\s(.*null.*)`, "$1?: $2"), // make null union types optional
		NewReDate("createdAt"),
		NewReDate("updatedAt"),
		NewReDate("soldDate"),
		NewReDate("purchaseDate"),
		NewReDate("warrantyExpires"),
		NewReDate("expiresAt"),
		NewReDate("date"),
		NewReDate("completedDate"),
		NewReDate("scheduledDate"),
	}

	for _, replace := range replaces {
		fmt.Printf("Replacing '%v' -> '%s'\n", replace.Regex, replace.Text)
		text = replace.Regex.ReplaceAllString(text, replace.Text)
	}

	err = os.WriteFile(path, []byte(text), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
