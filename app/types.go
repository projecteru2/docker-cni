package app

type OCIPhase int

const (
	CreatePhase OCIPhase = iota
	StartPhase
	DeletePhase
	OtherPhase
)

func parsePhase(args []string) OCIPhase {
	for _, arg := range args {
		switch arg {
		case "create":
			return CreatePhase
		case "start":
			return StartPhase
		case "delete":
			return DeletePhase
		}
	}
	return OtherPhase
}
