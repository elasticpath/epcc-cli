package misc

import "regexp"

var numericArgument = regexp.MustCompile(`^\s*-[0-9]+\s*$`)

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

	if !dashesAdded {
		if len(newArgs) >= 3 && newArgs[1] == "logs" {
			oldArgs := newArgs
			newArgs = make([]string, 0, len(args))

			for i := 0; i < len(oldArgs); i++ {
				if !dashesAdded && numericArgument.MatchString(oldArgs[i]) {
					dashesAdded = true
					newArgs = append(newArgs, "--")
				}

				newArgs = append(newArgs, oldArgs[i])
			}
		}
	}

	return newArgs

}
