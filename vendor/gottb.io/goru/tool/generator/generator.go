//go:generate goru packer template generator -out=.
package generator

type Generator func(args []string, options map[string]string)

const (
	generatedRoot     string = "generated"
	generatedRootFile string = "generated.go"
)
