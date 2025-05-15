package generator

// Strategy defines how fixes should be generated
type Strategy interface {
	GenerateFix(err error, context map[string]interface{}) (*Fix, error)
}

// Fix represents a generated code fix
type Fix struct {
	Description string
	Code       string
	Path       string
	Line       int
}

// BaseStrategy provides common functionality for fix strategies
type BaseStrategy struct {
	name string
}

func NewBaseStrategy(name string) *BaseStrategy {
	return &BaseStrategy{
		name: name,
	}
}
