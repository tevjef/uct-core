#!/bin/bash

format=protobuf #protobuf or json

camden() {
    CM_OLD=/tmp/uct/rutgers-cm-old.out
    CM_LATEST=/tmp/uct/rutgers-cm-latest.out
    CM_LOG=/var/log/uct/rutgers-cm.log
    if [ ! -f CM_LATEST ]; then
        rutgers -c CM -f protobuf | tee CM_LATEST | uct-db insert-all 2>&1 | tee -a CM_LOG
        exit 0
    fi
    cp CM_LATEST CM_OLD
    rutgers -c CM -f format | tee CM_LATEST | uct-diff -f format CM_OLD | uct-db -f format 2>&1 | tee -a CM_LOG
}

newark() {
    NK_OLD=/tmp/uct/rutgers-nk-old.out
    NK_LATEST=/tmp/uct/rutgers-nk-latest.out
    NK_LOG=/var/log/uct/rutgers-nk.log
    if [ ! -f NK_LATEST ]; then
        rutgers -c NK -f protobuf | tee NK_LATEST | uct-db insert-all 2>&1 | tee -a NK_LOG
        exit 0
    fi
    cp NK_LATEST NK_OLD
    rutgers -c NK -f format | tee NK_LATEST | uct-diff -f format NK_OLD | uct-db -f format 2>&1 | tee -a NK_LOG
}

newbrunswick() {
    NB_OLD=/tmp/uct/rutgers-nb-old.out
    NB_LATEST=/tmp/uct/rutgers-nb-latest.out
    NB_LOG=/var/log/uct/rutgers-nb.log
    if [ ! -f NB_LATEST ]; then
        rutgers -c NB -f protobuf | tee NB_LATEST | uct-db insert-all 2>&1 | tee -a NB_LOG
        exit 0
    fi
    cp NB_LATEST NB_OLD
    rutgers -c NB -f format | tee NB_LATEST | uct-diff -f format NB_OLD | uct-db -f format 2>&1 | tee -a NB_LOG
}

camden &
newark &
newbrunswick &

