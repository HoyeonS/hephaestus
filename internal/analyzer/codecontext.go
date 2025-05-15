package analyzer

// CodeContext represents the context around a code segment
type CodeContext struct {
	FilePath    string
	LineNumber  int
	LineContent string
	Context     []string // surrounding lines
}

// NewCodeContext creates a new code context
func NewCodeContext(filePath string, lineNumber int) *CodeContext {
	return &CodeContext{
		FilePath:   filePath,
		LineNumber: lineNumber,
	}
}

// AddContext adds surrounding lines to the context
func (c *CodeContext) AddContext(lines []string) {
	c.Context = lines
}

// SetLineContent sets the content of the target line
func (c *CodeContext) SetLineContent(content string) {
	c.LineContent = content
}
