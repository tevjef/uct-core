#!/bin/bash

format=protobuf #protobuf or json
path=/home/tevin/go/bin/

campus() {
    code=$1
    OLD=/tmp/uct/rutgers-${code}-old.out
    LATEST=/tmp/uct/rutgers-${code}-latest.out
    LOG=/var/log/uct/rutgers-${code}.log
    if [ ! -f LATEST ]; then
        ${path}rutgers -c ${code} -f protobuf | tee ${LATEST} | ${path}db insert-all 2>&1 | tee -a ${LOG}
        exit 0
    fi
    cp NB_LATEST NB_OLD
    ${path}rutgers -c ${code} -f format | tee ${LATEST} | ${path}diff -f format ${OLD} | ${path}db -f format 2>&1 | tee -a ${LOG}
}

campus CM &
campus NK &
campus NB &

wait $(jobs -p)