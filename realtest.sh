#!/bin/bash

cd /home/nrochon/terraform-provider-st2138
go build -o terraform-provider-st2138 .
export TF_CLI_CONFIG_FILE=/home/nrochon/terraform-provider-st2138/examples/catena-test/dev.tfrc
cd examples/catena-test
tofu apply