# usage: make -f Makefile VER={VER} PLATFORM={PLATFORM} RELEASE_VER=${RELEASE_VER}

# (required) [x.y.x]
VER=VER
# (required) [X86 ARM]
PLATFORM=PLATFORM

export GO111MODULE=on

Build_Version = github.com/huawei/cosi-driver/pkg/utils/version.buildVersion
Build_Arch = github.com/huawei/cosi-driver/pkg/utils/version.buildArch
flag = -ldflags '-w -s -bindnow -X "${Build_Version}=${VER}" -X "${Build_Arch}=${PLATFORM}"' -buildmode=pie

# Platform [X86, ARM]
ifeq (${PLATFORM}, X86)
env = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
else
env = CGO_ENABLED=0 GOOS=linux GOARCH=arm64
endif

all:PREPARE BUILD

PREPARE:
	rm -rf ${TMP_DIR_PATH}
	mkdir -p ${TMP_DIR_PATH}

BUILD:
	go mod tidy
# usage: [env] go build [-o output] [flags] packages
	${env} go build -o ${TMP_DIR_PATH}/huawei-cosi-driver ${flag} -buildmode=pie ./cmd/driver
	${env} go build -o ${TMP_DIR_PATH}/huawei-cosi-liveness-probe ${flag} -buildmode=pie ./cmd/livenessprobe

