# Command Line


Volume names must be in the format of `<name>-<verison>-<env>`.

- **uri**: URI to backend
- **dir**: Data directory to store under
- **prefix**: Key and path prefix for backend and folder structure

### Create volume

	voletc create <name> template:<key>=<value> <key>=<value> <key=value> [-dryrun]

- **-dryrun**: Do not actually create the volume

### Volume information

	voletc info <name>

### List volumes

	voletc ls

