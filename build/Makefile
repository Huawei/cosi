# usage: make -f Makefile VER={VER} PLATFORM={PLATFORM} RELEASE_VER=${RELEASE_VER}

# (required) [x.y.x]
VER=VER
# (required) [X86 ARM]
PLATFORM=PLATFORM
# (Optional) [2.5.RC1 2.5.RC2 ...] eSDK Version
RELEASE_VER=RELEASE_VER

export GO111MODULE=on
export GOPATH:=$(GOPATH):$(shell pwd)

Build_Version = github.com/huawei/cosi-driver/pkg/utils/version.buildVersion
Build_Arch = github.com/huawei/cosi-driver/pkg/utils/version.buildArch
flag = -ldflags '-w -s -linkmode "external" -extldflags "-Wl,-z,now" -X "${Build_Version}=${VER}" -X "${Build_Arch}=${PLATFORM}"' -buildmode=pie

# Platform [X86, ARM]
ifeq (${PLATFORM}, X86)
env = CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2" GOOS=linux GOARCH=amd64
else
env = CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2" GOOS=linux GOARCH=arm64
endif

all:PREPARE BUILD

PREPARE:
	rm -rf ./${PACKAGE_NAME}
	mkdir -p ./${PACKAGE_NAME}

BUILD:
	go mod tidy
# usage: [env] go build [-o output] [flags] packages
	${env} go build -o ./${PACKAGE_NAME}/huawei-cosi-driver ${flag} -buildmode=pie ../cmd/driver
	${env} go build -o ./${PACKAGE_NAME}/huawei-cosi-liveness-probe ${flag} -buildmode=pie ../cmd/livenessprobe

