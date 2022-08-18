package ghoststring

type nullGhostifyer struct{}

func (g *nullGhostifyer) Namespace() string { return "" }

func (g *nullGhostifyer) Ghostify(*GhostString) (string, error) {
	return "", nil
}

func (g *nullGhostifyer) Unghostify(string) (*GhostString, error) {
	return &GhostString{}, nil
}
