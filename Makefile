setup:
	go get -u github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...

godep-save:
	godep save ./...

bindata:
	go-bindata -pkg app -o ./app/bindata.go assets

test:
	go test ./app
