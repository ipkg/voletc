# vol-etc [![Build Status](https://travis-ci.org/ipkg/voletc.svg?branch=master)](https://travis-ci.org/ipkg/voletc) [![Release](https://img.shields.io/github/release/ipkg/voletc.svg)](https://github.com/ipkg/voletc/releases)

## Overview 

voletc (pronounced vol etc) is a Docker Volume Plugin that allows to create volumes containing application configurations that can be accessed on any of your docker nodes.  

Once created the application no longer needs to worry about obtaining application specific configurations.  The mounted volume will contain all of config file/s based on the templates and keys you've specified during volume creation.

The configuration data can be stored in consul or etcd but currently only consul support exists.

A config volume name must be in the following format
 
	<name>-<version>-<environment>


Based on this a volume is created per unique `name`, `version` and `environment`.  The layout looks like this:
- Each application has versions.
- Each version contains an associated template along with environments.
- Each environment contains its keys.
- Templates are shared across each environment and per environment keys are applied to the template

Volumes can be managed directly through [**docker**](#docker) and via the [**CLI**](#command-line).

## Docker 
The Docker CLI can be used to manage volumes using the `voletc` driver and driver specific options.

### Creating volumes

	docker volume create --name test-0.1.0-dev -d voletc \
		--opt=subpath/key=value \
		--opt=template:config.json='{"key":"${subpath/key}"}'

This creates a config volume with a `template` that will be called `config.json` and will also contain a key called `subpath/key` with a value of `value` persisting them to the backend.  When docker requests the volume to be mounted the configuration files will be dynamically generated and available on the newly created volume.

`template` is a keyword signifying that key contains template data rather than just key value data.


### Using volumes

	docker run --rm -it -v test-0.1.1-dev:/opt/myconfigs/ busbox


Your config should now be available at `/opt/myconfigs/config.json` in the running container.  If there are multiple config files they will all be located under `/opt/myconfigs`.  The naming of the config is controlled by what has been supplied as part of the `--opt=template:<name>.<ext>` argument at the time of creation.

### Removing volumes

	docker volume rm test-0.1.1-dev


Removing a volume removes all the associated keys from the backend for the given environment.  It does not remove the template (as it is associated to the version).

Currently, there is not a way via Docker to change volumes configs once they have been created.  You can either use the [**CLI**](#command-line) or destroy and re-create the volume.

## Command Line
The command line tool provides more functionality in regards to volume management than are available through docker.

	Usage:

	  voletc [options] <cmd> [name] [key=value] [key=value]

	  A tool to manage application configuration volumes.

	Commands:

	  ls        List volumes
	  create    Create new volume
	  edit      Edit volume configurations
	  info      Show volume info
	  rm        Destroy volume i.e. remove all keys
	  render    Show rendered volume templates
	  version   Show version

	Options:

	  -b string
	    	Service bind address (default "127.0.0.1:8989")
	  -dir string
	    	Data directory (default "/opt")
	  -prefix string
	    	Path prefix to store data under (default "voletc")
	  -uri string
	    	Backend uri (default "consul://localhost:8500")

Aside from the global options each command also has its specific options.

### Create a volume

	voletc create test-0.1.1-dev \
		db/name=dbname \
		db/user=dbuser \
		template:config.json=./config.json \
		template:inline.json='{"db_name": "${db/name}", "db_user": "${db/user}"}'

To simply simulate the creation rather than actually creating the volume, use the `-dryrun` flag.

### Get volume details

	voletc info test-0.1.1-dev

### List volumes

	voletc ls

### Edit a volume

	voletc edit test-0.1.1-dev db/user=new_user

To simply simulate the update rather than actually updating the volume configs, use the `-dryrun` flag.

### Render volume templates

	voletc render test-0.1.1-dev

### Remove a volume

	voletc rm test-0.1.1-dev

To remove the volume without being prompted include the `-y` flag.

## Installation
The current supported platforms are [Linux](#linux) and [OS X](#os-x).  Download the package from the [releases](https://github.com/ipkg/voletc/releases) page.

### Linux
Download the linux package from [here](https://github.com/ipkg/voletc/releases).  Once downloaded, untar and execute the installer.

	# Untar
	tar -zxvf voletc-0.1.6-linux.tgz
	
	# Run installer
	./voletc-installer.sh

This will install the binary and startup script.

- /usr/local/bin/voletc
- /etc/init/voletc.conf

You can now start the service as follows:

	# Start the agent
	start voletc

	# Make sure it is running
	status voletc

To troubleshoot the service check the log located at `/var/log/voletc.log`

### OS X
Download the darwin package from [here](https://github.com/ipkg/voletc/releases).  Once downloaded, untar the package.  You can now start using the `voletc` binary.  A detailed description on the usage can be found in the [**CLI**](#command-line) section.

### Roadmap

- Support for an encryption interface for stored data. 
- Support for etcd as a backend.
