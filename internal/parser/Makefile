fuzz-definitions:
	go-fuzz -workdir _fuzz

fuzz-definitions-build:
	go-fuzz-build -func FuzzDefinition -tags fuzz

fuzz-definitions-clear:
	rm -f _fuzz/crashers/*
	rm -f _fuzz/suppressions/*
