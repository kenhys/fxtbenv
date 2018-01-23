all: scripts fxtbenvctl

scripts: fxtbenvctl
	rm -f bin/firefox
	sed -e 's/__CPRODUCT__/FIREFOX/g'		\
		-e 's/__PRODUCT__/firefox/g'		\
		examples/product.in > bin/firefox
	chmod +x bin/firefox
	rm -f bin/thunderbird
	sed -e 's/__CPRODUCT__/THUNDERBIRD/g'		\
		-e 's/__PRODUCT__/thunderbird/g'	\
		examples/product.in > bin/thunderbird
	chmod +x bin/thunderbird

fxtbenvctl:
	go build -o bin/fxtbenvctl fxtbenvctl.go

