package resources

type ResourceGenerator interface {
	Generate(src, dest string) error
}

var generators = map[string]ResourceGenerator{
	"less":       &lessGenerator{},
	"sass":       &scssGenerator{},
	"scss":       &scssGenerator{},
	"typescript": &tsGenerator{},
}

func Generate(generatorName, src, dest string) error {
	generator := generators[generatorName]
	if generator == nil {
		return nil
	}
	return generator.Generate(src, dest)
}
