// Package builtin provides the default set of built-in tools.
package builtin

import (
	"github.com/WuSuBuDuoMing/aidev/internal/tool"
)

// NewDefaultRegistry creates a registry with all built-in tools.
func NewDefaultRegistry() *tool.Registry {
	r := tool.NewRegistry()
	r.Register(NewReadTool())
	r.Register(NewWriteTool())
	r.Register(NewEditTool())
	r.Register(NewGlobTool())
	r.Register(NewGrepTool())
	r.Register(NewBashTool())
	r.Register(NewGitStatusTool())
	r.Register(NewGitDiffTool())
	r.Register(NewGitCommitTool())
	r.Register(NewLsTool())
	return r
}
