// Based on the example at https://github.com/danslimmon/donothing/tree/main/example
package main

import (
	"fmt"
	"os"

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
		Text: `
# The magic of 8

This procedure implements a little arithmetic trick involving some manipulation
of the user's phone number.
`})

	pcd.AddStep(&otium.Step{
		Text: `
## Multiply your phone number by 8

Treating your phone number ('PhoneNumber') as a single integer, multiply
it by 8 ('PhoneNumberX8').
`,
		Run: func(bag otium.Bag) error {
			// This user input is needed also when step is automated.
			if _, err := bag.Get("PhoneNumber", "your phone number"); err != nil {
				return err
			}
			// This user input is needed only until the step is automated.
			if _, err := bag.Get("PhoneNumberX8", "the result"); err != nil {
				return err
			}

			return nil
		},
		//Run: func(bag otium.Bag) error {
		//	// FIXME this conversion should go in the otium loop!
		//	pNumber, err := strconv.Atoi(bag["PhoneNumber"])
		//	if err != nil {
		//		return err
		//	}
		//	pnx8 := pNumber * 8
		//	bag["PhoneNumberX8"] = strconv.Itoa(pnx8)
		//	return nil
		//},
	})

	pcd.AddStep(&otium.Step{
		Title: "Add up the digits",
		Desc: `
Given:
  - 'PhoneNumber':   {{.PhoneNumber}}
  - 'PhoneNumberX8': {{.PhoneNumberX8}}

Same instructions for both numbers:

A. 'SumPhoneNumber': add up all the digits in 'PhoneNumber', and then add
   8 to the result.
   If the sum has more than one digit, take that sum and add up its digits.
   Repeat until there's a single digit left. That digit should be 8.  

B. 'SumPhoneNumberX8': add up all the digits in 'PhoneNumberX8', and then add
   8 to the result.
   If the sum has more than one digit, take that sum and add up its digits.
   Repeat until there's a single digit left. That digit should be 8.  
`,
		Run: func(bag otium.Bag) error {
			// This user input is needed only until the step is automated.
			if _, err := bag.Get("SumPhoneNumber", "the result of A"); err != nil {
				return err
			}
			// This user input is needed only until the step is automated.
			if _, err := bag.Get("SumPhoneNumberX8", "the result of B"); err != nil {
				return err
			}

			return nil
		},
	})

	return pcd.Execute()
}
