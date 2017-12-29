#!/bin/bash

# 更新程序之前先根据监听端口与pid检测下原应用是否正确，因为可能存在其它应用占用相同端口或者相同pid.

module=$1
addr=$2

old_pid=$3
old_port=$4

function ssh_connect_test() {
    ssh -n -o StrictHostKeyChecking=no root@$addr "hostname"
}

ssh_connect_test

