package otium

import (
	"fmt"
	"os"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// cmdVariables implements the "variables" command.
func cmdVariables(pcd *Procedure) {
	if len(pcd.bag.bag) == 0 {
		fmt.Println("no variables")
		return
	}

	keys := maps.Keys(pcd.bag.bag)
	slices.Sort(keys)

	for _, k := range keys {
		v := pcd.bag.bag[k]
		if v.set {
			fmt.Printf("%s (%s): %v\n", k, v.Desc, v.val)
		} else {
			fmt.Printf("%s (%s): <unset>\n", k, v.Desc)
		}
	}
}

type visitFn func(pcd *Procedure, step *Step) error

// cmdNext implements the "next" command.
func cmdNext(pcd *Procedure) error {
	return visitStep(pcd, visitAsNext)
}

func visitStep(pcd *Procedure, visitor visitFn) error {
	if pcd.stepIdx >= len(pcd.steps) {
		return fmt.Errorf("next: internal error: step index > len(steps)")
	}
	step := pcd.steps[pcd.stepIdx]

	fmt.Printf("\n## %d. %s %s\n\n", pcd.stepIdx+1, step.Icon(), step.Title)

	if step.Desc != "" {
		if err := renderTemplate(os.Stdout, step.Desc, pcd.bag.bag); err != nil {
			return fmt.Errorf("%s %w", err, ErrUnrecoverable)
		}
		fmt.Printf("\n\n")
	}

	if visitor != nil {
		if err := visitor(pcd, step); err != nil {
			return err
		}
	}

	pcd.stepIdx++
	return nil
}

func visitAsNext(pcd *Procedure, step *Step) error {
	// Prompt the user for the declared variables.
	for _, variable := range step.Vars {
		if _, err := pcd.bag.ask(variable.Name, pcd.term); err != nil {
			return err
		}
	}

	// Run the step.
	if step.Run != nil {
		if err := step.Run(pcd.bag, pcd.uctx); err != nil {
			return fmt.Errorf("step %d: %w", pcd.stepIdx+1, err)
		}
	}

	return nil
}
