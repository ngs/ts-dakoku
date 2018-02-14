setup:
	go get -u github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...

godep-save:
	godep save ./...

bindata:
	go-bindata -pkg app -o ./app/bindata.go assets

test:
	go test -v ./app

test-apex:
	docker pull atsnngs/force-com
	docker run \
		-v $$(pwd)/apex/wsdl:/wsdl \
		-v $$(pwd)/apex/src/classes:/src/classes \
		--rm \
		--env SF_USERNAME=$$SF_USERNAME \
		--env SF_PASSWORD=$$SF_PASSWORD \
		--env SF_SERVER=$$SF_SERVER \
		atsnngs/force-com
