package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/marco-m/otium"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
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
		Title: "Change directory",
		Desc: `
Change directory to $HOME/tmp.

    chdir ~/tmp
`})

	pcd.AddStep(&otium.Step{
		Title: "Download a file",
		Desc: `
Download a file of your choice and put it in the tmp directory.

    curl --location -O https://en.wikipedia.org/wiki/Special:Random 
`,
		Run: func(bag otium.Bag) error {
			// This user input is needed also when step is automated.
			URL, err := bag.Get("URL", "URL to download")
			if err != nil {
				return err
			}
			_, file := path.Split(URL)
			bag.Put("File", file)

			return nil
		},
	})

	pcd.AddStep(&otium.Step{
		Title: "Calculate the checksum",
		Desc: `
Calculate the checksum of the downloaded file.

    sha256sum {{.File}}
`,
		Run: func(bag otium.Bag) error {
			file, err := bag.Get("File", "")
			if err != nil {
				return err
			}

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

	return pcd.Execute()
}
