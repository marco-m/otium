package otium

import (
	"fmt"
	"os"
	"strings"
)

func cmdNext(pcd *Procedure) error {
	if pcd.stepIdx >= len(pcd.steps) {
		return fmt.Errorf("next: internal error: step index > len(steps)")
	}
	step := pcd.steps[pcd.stepIdx]
	fmt.Printf("\n## (%d) %s\n\n", pcd.stepIdx+1, strings.TrimSpace(step.Title))

	if step.Desc != "" {
		if err := render(os.Stdout, strings.TrimSpace(step.Desc),
			pcd.bag); err != nil {
			return fmt.Errorf("%s %w", err, ErrUnrecoverable)
		}
		fmt.Printf("\n\n")
	}

	if err := step.Run(pcd.bag); err != nil {
		return fmt.Errorf("step %d: %w", pcd.stepIdx+1, err)
	}
	pcd.stepIdx++

	return nil
}
