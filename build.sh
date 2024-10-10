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

# usage: bash build.sh {VER} {PLATFORM}

# [x.y.z]
VER=$1
# [X86 ARM]
PLATFORM=$2

set -e
workdir=$(cd $(dirname $0); pwd)

# tmp dir is used to build binary files and images
export TMP_DIR_PATH="${workdir}/eSDK_COSI_V${VER}_${PLATFORM}_64"
# release dir is used to assemble the release package
release_dir_path="${workdir}/release"

# init tmp dir and release dir
rm -rf "${TMP_DIR_PATH}"
mkdir -p "${TMP_DIR_PATH}"
rm -rf "${release_dir_path}/image"
mkdir -p "${release_dir_path}/image"

echo "Start to make with Makefile"
make -f Makefile VER=$1 PLATFORM=$2

echo "Platform confirmation"
if [[ "${PLATFORM}" == "ARM" ]];then
  PULL_FLAG="--platform=arm64"
  BUILD_FLAG="--platform linux/arm64"
elif [[ "${PLATFORM}" == "X86" ]];then
  PULL_FLAG="--platform=amd64"
  BUILD_FLAG="--platform linux/amd64"
else
  echo "Wrong PLATFORM, support [X86, ARM]"
  exit
fi

echo "Start to pull busybox image with architecture"
docker pull ${PULL_FLAG} busybox:stable-glibc

# build the image
function build_image() {
    cp -rf Dockerfile "${TMP_DIR_PATH}"/Dockerfile

    # cd to tmp dir to build image
    cd "${TMP_DIR_PATH}"
    local images=("huawei-cosi-driver" "huawei-cosi-liveness-probe")
    # shellcheck disable=SC2068
    for img in ${images[@]}; do
      echo "build the ${img} image"
      chmod +x "${img}"
      docker build ${BUILD_FLAG} -f Dockerfile -t ${img}:${VER} --target ${img} --build-arg VER=${VER} .
      docker save ${img}:${VER} -o ${img}-${VER}.tar
      mv ${img}-${VER}.tar ${release_dir_path}/image
    done
}
build_image

# pack the package
echo "pack deploy files"
cp -rf "${workdir}"/helm "${release_dir_path}"

echo "pack example files"
cp -rf "${workdir}"/examples "${release_dir_path}"

# cd to release dir to pack the package
cd "${release_dir_path}"

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

# zip the release package and move it to workdir
zip -rq -o eSDK_Huawei_Storage_COSI_V"${VER}"_"${PLATFORM}"_64.zip ./*
mv eSDK_Huawei_Storage_COSI_V"${VER}"_"${PLATFORM}"_64.zip "${workdir}"

# cd to workdir to remove tmp files
cd "${workdir}"
rm -rf "${TMP_DIR_PATH}"
rm -rf "${release_dir_path}"
