#!/bin/bash
export GOROOT="${HOME}/prgs/go"
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${GOROOT}/bin"

dp0="`dirname \"$0\"`" # relative
dp0="`( cd \"${dp0}\" && pwd )`"
gogithub="src/github.com/google/go-github"

if [[ ! -d "${dp0}/deps/${gogithub}/.git" ]]; then git submodule update --init ; fi

if [[ "${GOPATH}" != "" && -d "${GOPATH}" ]]; then
	if [[ ! -d "${GOPATH}/${gogithub}" ]]; then
		ln -s ${dp0}/deps/${gogithub} ${GOPATH}/${gogithub}
		cd ${GOPATH}/${gogithub}
		go install
	fi 
fi
export GOPATH="${dp0}/deps"
export GOBIN="${dp0}/bin"
cd ${dp0}
go install
echo alias seec=${dp0}/bin/seec
