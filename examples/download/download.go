package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/marco-m/otium"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Download a file and calculate its checksum",
		Desc: `
This is a silly example of how to use otium.

To follow this example, you will need a separate terminal to invoke shell commands.
`})

	pcd.AddStep(&otium.Step{
		Title: "Collect inputs",
		Desc: `
a. Choose your working directory (in another terminal).
b. Choose the URL to download.

NOTE: you can also set these parameters directly as CLI flags; try executing
this program with '-h'.
`,
		Vars: []otium.Variable{
			{Name: "pwd", Desc: "The output of 'pwd'"},
			{
				Name: "URL",
				Desc: "URL to download",
				// NOTE this shows how to use the validator function to set
				// new k/v pairs in the Bag.
				Fn: func(val string) error {
					_, file := path.Split(val)
					pcd.Put("file", file)
					return nil
				},
			},
		},
	})

	pcd.AddStep(&otium.Step{
		Title: "Download the file",
		Desc: `
Download the previous URL and put it in the pwd directory.

    curl --location -O {{.URL}}
`,
		//Run: func(bag otium.Bag) error {
		//	// actually download the URL and save it to bag[pwd]/bag[file]
		//},
	})

	pcd.AddStep(&otium.Step{
		Title: "Calculate the checksum",
		Desc: `
Calculate the checksum of the downloaded file.

    sha256sum {{.pwd}}/{{.file}}
`,
		Run: func(bag otium.Bag, uctx any) error {
			file, err := bag.Get("file")
			if err != nil {
				return err
			}
			pwd, err := bag.Get("pwd")
			if err != nil {
				return err
			}
			file = filepath.Join(pwd, file)

			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			h := sha256.New()
			if _, err := io.Copy(h, f); err != nil {
				return err
			}
			result := fmt.Sprintf("%x", h.Sum(nil))
			bag.Put("FileSHA256", result)
			fmt.Println(result, file)

			return nil
		},
	})

	return pcd.Execute(os.Args)
}
