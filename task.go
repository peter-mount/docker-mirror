package mirror

import "fmt"

type Task func() error

// RunTasks runs a series of Task's in sequence.
func RunTasks(tasks ...Task) error {
	if *verbose {
		fmt.Printf("Executing %d tasks\n", len(tasks))
	}

	for _, task := range tasks {
		err := task()
		if err != nil {
			return err
		}
	}

	return nil
}
