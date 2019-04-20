export VERSION=1.0.0
export RELEASE_PATH="releases/salary-${VERSION}"

all: build
publish: build build_mac build_linux build_windows gen_version

deps:
	-go get -u github.com/tealeg/xlsx
#	go get -u github.com/jteeuwen/go-bindata/...

testdeps: deps

#metas: deps
#	go-bindata -o resource/metas.go -pkg templates -prefix ../salary-meta ../salary-meta/...

clean:
	rm -rf out/*

build: deps
	go build -ldflags "-X 'github.com/sahara/salary-tools/exe.Version=${VERSION}'" -o out/salary main/main.go

#install: build
#	cp out/salary /usr/local/bin

build_mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'github.com/sahara/salary-tools/dmg.Version=${VERSION}'" -o out/salary main/main.go
	tar zcvf out/salary-macosx-${VERSION}-amd64.tgz -C out salary

build_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X 'github.com/sahara/salary-tools/pkg.Version=${VERSION}'" -o out/salary main/main.go
	tar zcvf out/salary-linux-${VERSION}-amd64.tgz -C out salary

build_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X 'github.com/sahara/salary-tools/exe.Version=${VERSION}'" -o salary.exe main/main.go
	zip -r out/salary-windows-${VERSION}-amd64.zip salary.exe
	rm salary.exe

gen_version:
	-rm out/version
	echo ${VERSION} >> out/version

git_release: clean build make_release_dir release_mac release_linux release_windows

make_release_dir:
	mkdir -p ${RELEASE_PATH}

release_mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'github.com/sahara/salary-tools/dmg.Version=${VERSION}'" -o out/salary main/main.go
	tar zcvf ${RELEASE_PATH}/salary-darwin-amd64.tar.gz -C out salary

release_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X 'github.com/sahara/salary-tools/pkg.Version=${VERSION}'" -o out/salary main/main.go
	tar zcvf ${RELEASE_PATH}/salary-linux-amd64.tar.gz -C out salary

release_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X 'github.com/sahara/salary-tools/exe.Version=${VERSION}'" -o salary.exe main/main.go
	zip -r ${RELEASE_PATH}/salary-windows-amd64.exe.zip salary.exe
	rm salary.exe

fmt:
	go fmt ./main/...

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./main/...
	go tool cover -html=coverage.txt -o coverage.html
