bindata:
		go-bindata --pkg frontend -o frontend/bindata.go --prefix frontend frontend/

godeps_save:
		godeps save ./...
