all: fxtbenvctl

fxtbenvctl:
	go build -o bin/fxtbenvctl fxtbenvctl.go
