#! /bin/bash

find . -type f -not -name "*.sha256sum" -exec sha256sum {} \;
