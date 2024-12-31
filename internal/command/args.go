package command

type Arg struct {
	Name      string
	ShortHand string
	Value     string
	Usage     string
	Required  bool
}
