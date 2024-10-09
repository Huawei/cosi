# eg: docker build --target xxx --platform linux/amd64 --build-arg VER=${VER} -f Dockerfile -t xxx:${VER} .
ARG VER

# build huawei-cosi-driver image
FROM busybox:stable-glibc as huawei-cosi-driver
LABEL version="${VER}"
LABEL maintainers="Huawei COSI Authors"
LABEL description="Huawei COSI Driver"

ARG binary=./huawei-cosi-driver
COPY ${binary} huawei-cosi-driver
ENTRYPOINT ["/huawei-cosi-driver"]

# build huawei-cosi-liveness-probe image
FROM busybox:stable-glibc as huawei-cosi-liveness-probe
LABEL version="${VER}"
LABEL maintainers="Huawei COSI Authors"
LABEL description="Huawei COSI Driver Liveness Probe"

ARG binary=./huawei-cosi-liveness-probe
COPY ${binary} huawei-cosi-liveness-probe
ENTRYPOINT ["/huawei-cosi-liveness-probe"]