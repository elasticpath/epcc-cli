BINARY=epcc
VERSION=NIGHTLY

build:
	go build -o ./bin/${BINARY} ./

release-some:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/${VERSION}/darwin_amd64/${BINARY}
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/${VERSION}/darwin_arm64/${BINARY}
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/${VERSION}/windows_amd64/${BINARY}.exe
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/${VERSION}/linux_amd64/${BINARY}


release: release-some
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o ./bin/${VERSION}/windows_386/${BINARY}.exe
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -o ./bin/${VERSION}/freebsd_386/${BINARY}
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o ./bin/${VERSION}/freebsd_amd64/${BINARY}
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -o ./bin/${VERSION}/freebsd_arm/${BINARY}
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o ./bin/${VERSION}/linux_386/${BINARY}
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./bin/${VERSION}/linux_arm/${BINARY}
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386 go build -o ./bin/${VERSION}/openbsd_386/${BINARY}
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -o ./bin/${VERSION}/openbsd_amd64/${BINARY}
	CGO_ENABLED=0 GOOS=solaris GOARCH=amd64 go build -o ./bin/${VERSION}/solaris_amd64/${BINARY}

clean:
	rm -rf bin || true