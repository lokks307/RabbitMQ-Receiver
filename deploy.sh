#!/bin/bash

# subscriber list와 version 배열 선언
subs=("careease-api" "careplanner-admin-api" "careplanner-api" "careplanner-mobile-api" "msv-saas-appt" "msv-saas-message" "msv-saas-checkup" "msv-saas-phr")
version=("0.0.1" "0.0.1" "0.0.1" "0.0.1" "0.0.1" "0.0.1" "0.0.1" "0.0.1")
runmode="test"

for ((i=0; i<${#subs[@]}; i++)); do
    dir="/home/user/logs/${subs[i]}-${runmode}-logs"
    dockerimage="${subs[i]}-subs:${version[i]}"
    sudo docker stop ${subs[i]}-subs
    sudo docker rm ${subs[i]}-subs
    
    sudo mkdir -p $dir

    sudo docker build --build-arg SUBS="${subs[i]}" --build-arg RUNMODE="${runmode}" -t $dockerimage .
    sudo docker tag ${dockerimage} "${subs[i]}-subs:latest"

    sudo docker run -d --name "${subs[i]}-subs" \
    -v $dir:/app/logs/"${subs[i]}-${runmode}-logs" \
    $dockerimage
done
