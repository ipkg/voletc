# voletc
voletc (pronounced vol etc) is a Docker Volume Plugin that allows to create volumes containing application configurations that can be accessed on any of your docker nodes.  

Once created the application no longer needs to worry about obtaining application specific configurations.  The mounted volume will contain all of config file/s based on the templates and keys you've specified during volume creation.

The configuration data can be stored in consul or etcd but currently only consul support exists.

## Overview

A config volume name must be in the following format
 
	<name>-<version>-<environment>


Based on this a volume is created per unique `name`, `version` and `environment`.  The layout looks like this:
- Each application has versions.
- Each version contains an associated template along with environments.
- Each environment contains its keys.
- Templates are shared across each environment and per environment keys are applied to the template

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


### Installation
Download the latest stable linux installer.  Once downloaded, untar and execute the installer.

	# Download installer
	curl -O -sL https://github.com/ipkg/voletc/releases/download/v0.1.1/voletc-0.1.1.tgz
	
	# Untar
	tar -zxvf voletc-0.1.1.tgz
	
	# Run installer
	./voletc-installer

	# Start the agent
	start voletc

	# Make sure it is running
	status voletc

If the service does not start check the log located at `/var/log/voletc.log`

#### Roadmap

- Support for an encryption interface for stored data. 
- Support for etcd as a backend.
