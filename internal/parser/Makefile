fuzz-definitions:
	go-fuzz -workdir _fuzz/definitions/

fuzz-definitions-build:
	go-fuzz-build -func FuzzDefinition -tags fuzz

fuzz-definitions-clear:
	rm -f _fuzz/definitions/crashers/*
	rm -f _fuzz/definitions/suppressions/*

fuzz-schema:
	go-fuzz -workdir _fuzz/schema/

fuzz-schema-build:
	go-fuzz-build -func Fuzz -tags fuzz

fuzz-schema-clear:
	rm -f _fuzz/schema/crashers/*
	rm -f _fuzz/schema/suppressions/*

fuzz-schema-coordinator:
	go-fuzz -coordinator localhost:1105 -workdir _fuzz/schema/
