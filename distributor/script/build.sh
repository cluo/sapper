#!/bin/bash

set -e

url=$1

framework_package="github.com/dearcode/sapper/goapi"

function convert_url() {
    if [[ "$url" =~ "http://" ]]
    then
        url=`echo $url|sed 's/http:\/\//git@/g'|sed 's/\//:/'|sed 's/$/.git/'`
    fi
}

function create_path() {
    base_path=`echo $url|awk -F'[@:/]' '{print "src/"$2"/"$3}'`
    rm -rf $base_path
    mkdir -p $base_path
}

function clone_source() {
    cd $base_path;
    git clone $url;
    cd -;

    project=`echo $url|sed 's/.*://'|sed 's/\.git//'`;
    app=`echo $url|xargs basename -s .git`;
    base_path=$base_path/$app;

    cd $base_path;

    git_hash=`git log --pretty=format:'%H' -1`
    git_time=`git log --pretty=format:'%ci' -1`
    git_message=`git log --pretty=format:'%cn %s %b' -1`

    rm -rf .git

    cd -;
}

function create_dockerfile() {
    package=`echo $url|sed 's/.*@//'|sed 's/\.git//'|sed 's/:/\//'`;
    package_in_vendor="$package/vendor/$framework_package"
    cp Dockerfile.tpl Dockerfile
    sed -i "s#{{.APP}}#\"$app\"#" Dockerfile
    sed -i "s#{{.PROJECT}}#\"$project\"#" Dockerfile
    sed -i "s#{{.BASE_PATH}}#\"$base_path\"#" Dockerfile
    local ldflags="\'-X \"$package_in_vendor/debug.GitHash=$git_hash\" -X \"$package_in_vendor/debug.GitTime=$git_time\" -X \"$package_in_vendor/debug.GitMessage=$git_message\"\'"
    sed -i "s#{{.LDFLAGS}}#$ldflags#" Dockerfile
}

function build() {
    version=`date -d "$git_time" +%Y%m%d.%H%M`
    image="$project:$version"
    docker build --no-cache -t $image .
    docker run -i --rm -v $PWD/bin:/base $image bash -c 'cp $GOPATH/bin/* /base/' 
}


convert_url

create_path

clone_source

create_dockerfile

build
