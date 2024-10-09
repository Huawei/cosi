#!/bin/bash
#
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#
set -e
workdir=$(cd $(dirname $0); pwd)

export PACKAGE_NAME="eSDK_${RELEASE_VER}_COSI_V${VER}_${PLATFORM}_64"
export GOPROXY=http://mirrors.tools.huawei.com/goproxy/
export GOSUMDB=off
# shellcheck disable=SC2164
# shellcheck disable=SC2154
if [ "${isRelease}" == "true" ]; then
    echo "buildVersion=${ENV_RELEASE_VERSION}" > buildInfo.properties
else
    echo "buildVersion=${RELEASE_VER}.$(date "+%Y%m%d%H%M%S")" > buildInfo.properties
fi

### step 1: build the binary
# shellcheck disable=SC2164
cd ${workdir}
make -f Makefile VER="${VER}" PLATFORM="${PLATFORM}" RELEASE_VER="${RELEASE_VER}"

### step 2: load the image
if [ "${PLATFORM}" == "ARM" ]; then
    wget http://10.29.160.97/busybox-arm.tar
    docker load -i busybox-arm.tar
    docker tag busybox:1.36.1 busybox-arm:stable-glibc
    sed -i 's/busybox:stable-glibc/busybox-arm:stable-glibc/g' Dockerfile
else
    wget http://10.29.160.97/busybox-x86.tar
    docker load -i busybox-x86.tar
    docker tag busybox:1.36.1 busybox-x86:stable-glibc
    sed -i 's/busybox:stable-glibc/busybox-x86:stable-glibc/g' Dockerfile
fi

### step 3: build the image
function build_image() {
    # shellcheck disable=SC2164
    cp -rf Dockerfile ./"${PACKAGE_NAME}"/Dockerfile
    # shellcheck disable=SC2164
    cd ./"${PACKAGE_NAME}"/
    echo "create image dir"
    mkdir -p ../release/image/
    # shellcheck disable=SC2054
    local images=("huawei-cosi-driver" "huawei-cosi-liveness-probe")
    # shellcheck disable=SC2068
    for img in ${images[@]}; do
      echo "build the ${img} image"
      chmod +x "${img}"
      # shellcheck disable=SC2086
      docker build -f Dockerfile -t "${img}":${VER} --target "${img}" --build-arg VER=${VER} .
      docker save "${img}":"${VER}" -o "${img}"-"${VER}".tar
      mv "${img}"-"${VER}".tar ../release/image/
    done
}
build_image

### step 4: pack the package
echo "pack deploy files"
cp -rf ../../helm ../release

echo "pack example files"
cp -rf ../../examples ../release

# shellcheck disable=SC2164
cd ../release

# set version in values.yaml and Chart.yaml
sed -i "s/{{version}}/${VER}/g" helm/values.yaml

# parse ${VAR} to Semantic Version style
# Charts https://helm.sh/docs/topics/charts/
# Semantic Version https://semver.org/lang/zh-CN/`
# example:
#   2.0.0.B070 -> 2.0.0-B070
#   2.0.0 -> 2.0.0
chart_version=$(echo ${VER} | sed -e 's/\([0-9]\+\.[0-9]\+\.[0-9]\+\)\./\1-/')
sed -i "s/{{version}}/${chart_version}/g" helm/Chart.yaml

# shellcheck disable=SC2035
zip -rq -o eSDK_Huawei_Storage_"${RELEASE_VER}"_COSI_V"${VER}"_"${PLATFORM}"_64.zip *
mkdir ${workdir}/../../output
cp eSDK_Huawei_Storage_"${RELEASE_VER}"_COSI_V"${VER}"_"${PLATFORM}"_64.zip ${workdir}/../../output
