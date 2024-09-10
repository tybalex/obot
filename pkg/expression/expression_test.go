package expression

import (
	"context"
	"testing"

	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/storage/scheme"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestExpressions(t *testing.T) {
	objs := []runtime.Object{
		&v1.WorkflowStep{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: v1.WorkflowStepSpec{
				ParentWorkflowStepName: "parent",
				Step: v1.Step{
					Name: "test",
				},
			},
			Status: v1.WorkflowStepStatus{
				LastRunName: "testrun",
			},
		},
		&v1.WorkflowStep{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
				Namespace: "default",
			},
			Spec: v1.WorkflowStepSpec{
				ParentWorkflowStepName: "parent",
				Step: v1.Step{
					Name: "test2",
				},
			},
			Status: v1.WorkflowStepStatus{
				LastRunName: "testrun2",
			},
		},
		&v1.Run{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testrun",
				Namespace: "default",
			},
			Spec: v1.RunSpec{
				Input: "test run input",
			},
			Status: v1.RunStatus{
				Output: "test run output",
			},
		},
		&v1.Run{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testrun2",
				Namespace: "default",
			},
			Spec: v1.RunSpec{
				Input: "test run 2 input",
			},
			Status: v1.RunStatus{
				Output: `["test run 2 output"]`,
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objs...).Build()
	x, err := Eval(context.Background(), c, objs[0].(*v1.WorkflowStep), "steps.test2.output.json[0]")
	require.NoError(t, err)
	autogold.Expect("test run 2 output").Equal(t, x)
}
