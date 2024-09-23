package workflow

import (
	"github.com/acorn-io/baaah/pkg/randomtoken"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func PopulateIDs(manifest v1.WorkflowManifest) v1.WorkflowManifest {
	manifest = *manifest.DeepCopy()
	ids := map[string]struct{}{}
	for i, step := range manifest.Steps {
		manifest.Steps[i] = populateStepID(ids, step)
	}
	return manifest
}

func FindStep(manifest *v1.WorkflowManifest, id string) *v1.Step {
	if manifest == nil || id == "" {
		return nil
	}
	return findInSteps(manifest.Steps, id)
}

func findInSteps(steps []v1.Step, id string) *v1.Step {
	for _, step := range steps {
		if step.ID == id {
			return &step
		}
		if step.While != nil {
			if found := findInSteps(step.While.Steps, id); found != nil {
				return found
			}
		}
		if step.If != nil {
			if found := findInSteps(step.If.Steps, id); found != nil {
				return found
			}
			if found := findInSteps(step.If.Else, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func nextID(seen map[string]struct{}) string {
	for {
		next, err := randomtoken.Generate()
		if err != nil {
			panic(err)
		}
		id := "si1" + next[:5]
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			return id
		}
	}
}

func populateStepID(seen map[string]struct{}, step v1.Step) v1.Step {
	if step.ID == "" {
		step.ID = nextID(seen)
	} else if _, ok := seen[step.ID]; ok {
		step.ID = nextID(seen)
	}

	if step.While != nil {
		step.While = populateWhileID(seen, *step.While)
	}
	if step.If != nil {
		step.If = populateIfID(seen, *step.If)
	}
	return step
}

func populateWhileID(seen map[string]struct{}, while v1.While) *v1.While {
	for i, step := range while.Steps {
		while.Steps[i] = populateStepID(seen, step)
	}
	return &while
}

func populateIfID(seen map[string]struct{}, ifStep v1.If) *v1.If {
	for i, step := range ifStep.Steps {
		ifStep.Steps[i] = populateStepID(seen, step)
	}
	for i, step := range ifStep.Else {
		ifStep.Else[i] = populateStepID(seen, step)
	}
	return &ifStep
}
