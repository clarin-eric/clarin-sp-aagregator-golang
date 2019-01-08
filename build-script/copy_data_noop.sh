#!/bin/bash

init_data (){
    LOCAL=0
    if [ "$1" == "local" ]; then
        LOCAL=1
    fi

    if [ "${LOCAL}" -eq 0 ]; then
        #Remote / gitlab ci
        echo -n ""
    else
        #Local copy
        echo -n ""
    fi
}

cleanup_data () {
    echo -n ""
}