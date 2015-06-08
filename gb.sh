#!/bin/bash
export GOROOT="${HOME}/prgs/go"
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${GOROOT}/bin"

updateGlobalDep () {
	dep=$1
	sub=$2
	if [ -d "${GOPATH}/src/${dep}" ] && ! [ -L "${GOPATH}/src/${dep}" ]; then
		mv "${GOPATH}/src/${dep}" "${GOPATH}/src/${dep}_latest"
	fi
	if [[ ! -d "${GOPATH}/src/${dep}" ]]; then
		mkdir -p "${GOPATH}/src/${dep}"
		rmdir "${GOPATH}/src/${dep}"
		ln -s ${dp0}/deps/src/${dep} ${GOPATH}/src/${dep}
	fi
	cd ${GOPATH}/src/${dep}/${sub}
	# uses the global GOPATH, before it is modified below
	go install
}

dp0="`dirname \"$0\"`" # relative
dp0="`( cd \"${dp0}\" && pwd )`"
if [[ ! -d "${dp0}/.git/modules" ]]; then git submodule update --init ; fi

if [[ "${GOPATH}" != "" && -d "${GOPATH}" ]]; then
	updateGlobalDep "github.com/VonC/godbg"
	updateGlobalDep "github.com/atotto/clipboard"
	updateGlobalDep "github.com/google/go-querystring" "query"
	updateGlobalDep "github.com/google/go-github" "github"
fi
export GOPATH="${dp0}/deps"
export GOBIN="${dp0}/bin"
cd ${dp0}
go install
echo alias seec=${dp0}/bin/seec

