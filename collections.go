package main

//TODO: you have any more of them, generics?

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Unique(e []string) []string {
	r := []string{}

	for _, s := range e {
		if !Contains(r[:], s) {
			r = append(r, s)
		}
	}
	return r
}

func Contains(e []string, c string) bool {
	for _, s := range e {
		if s == c {
			return true
		}
	}
	return false
}
