package manager

import (
	"os"
)

// func init() {
// 	checkSetupCondaFile()
// }

func checkSetupCondaFile() error {
	// write SETUP_CONDA_CONTENT to setup_conda.sh
	err := os.WriteFile("setup_conda.sh", []byte(SETUP_CONDA_CONTENT), 0777)
	if err != nil {
		return err
	}
	return nil
}

const (
	SETUP_CONDA_CONTENT = `
	#----------------------------------------------
	# DO NOT MODIFY!
	# This script is automatically generated by eternal-infer-worker!
	#----------------------------------------------
	#!/bin/bash
	
	target_env=$1
	environment_file=$2
	
	echo "target_env: $target_env"
	echo "environment_file: $environment_file"
	
	ENVS=$(conda env list | awk '{print $1}' | grep -w $target_env)
	
	conda init --all --dry-run --verbose
	
	if [[ "$ENVS" == "" ]] 
	then
		echo "Creating new environment"
		conda env create -n $target_env -f $environment_file
		conda run -n $target_env python -V
		exit
	else 
		echo $ENVS
		conda run -n $target_env python -V
	fi;
`
)
