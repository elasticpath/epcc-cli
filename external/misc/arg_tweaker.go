package misc

func AddImplicitDoubleDash(args []string) []string {

	newArgs := make([]string, 0, len(args))
	dashesAdded := false

	// This adds a -- before a sort where the next argument starts with a dash.
	// -- is a standard shell idiom to turn of flag parsing.
	for i := 0; i < len(args); i++ {
		if args[i] == "sort" {
			if i < len(args)-1 {
				nextArg := args[i+1]
				if len(nextArg) > 0 {
					if nextArg[0] == '-' {
						if !dashesAdded {
							newArgs = append(newArgs, "sort", "--", nextArg)
							i++
							dashesAdded = true
							continue
						}
					}
				}
			}
		} else if args[i] == "--" {
			dashesAdded = true
		}
		newArgs = append(newArgs, args[i])

	}

	return newArgs

}
