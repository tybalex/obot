package workflowstep

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_getItemCount(t *testing.T) {
	vm := goja.New()
	err := vm.Set("count", 2)
	require.NoError(t, err)
	// Test javascript function
	v, err := vm.RunString(`
function getItemCount() {
	return 2
}
count + getItemCount()`)
	require.NoError(t, err)
	i := v.Export().(int64)
	assert.Equal(t, int(i), 4)

	vm.ToValue().ToObject()
	vm.NewProxy()
}
