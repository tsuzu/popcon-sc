package utility

func FunctionJoin(functions ...func()) func() {
	return func() {
		for i := range functions {
			functions[i]()
		}
	}
}
