load(
    "//bazel:envoy_build_system.bzl",
    "envoy_cc_library",
    "envoy_extension_package",
)

licenses(["notice"])  # Apache 2

envoy_extension_package()

envoy_cc_library(
    name = "zlib_base_lib",
    srcs = ["base.cc"],
    hdrs = ["base.h"],
    deps = [
        "//bazel/foreign_cc:zlib",
        "//source/common/buffer:buffer_lib",
    ],
)
