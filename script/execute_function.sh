#!/bin/bash

# Number of times to execute the kubectl command
num_executions=500

# Command to execute
command="kubectl execute default/func1 '{\"x\":1, \"y\": 2}'"

# Execute the command multiple times
for ((i=1; i<=$num_executions; i++)); do
    echo "Executing command: $command (Execution $i)"
    eval $command
done