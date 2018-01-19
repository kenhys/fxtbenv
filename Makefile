all: scripts fxtbenvctl

scripts: fxtbenvctl
	sed -e 's/__REFPRODUCT__/$$PRODUCT/g'		\
		-e 's/__VARPRODUCT__/PRODUCT/g'		\
		-e 's/__PRODUCT__/firefox/g'		\
		examples/product.in > bin/firefox
	chmod +x bin/firefox
	sed -e 's/__REFPRODUCT__/$$PRODUCT/g'		\
		-e 's/__VARPRODUCT__/PRODUCT/g'		\
		-e 's/__PRODUCT__/thunderbird/g'	\
		examples/product.in > bin/thunderbird
	chmod +x bin/thunderbird

fxtbenvctl:
	go build -o bin/fxtbenvctl fxtbenvctl.go

