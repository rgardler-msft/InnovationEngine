This document is to show the hierarchy of environment variables 


<!---
```variables
export MY_RESOURCE_GROUP=setInComments
export MY_VARIABLE_NAME=commentVariable
```
--->

This first check echoes whatever values were supplied via comments or inherited environment.

```bash
echo $MY_RESOURCE_GROUP
echo $MY_VARIABLE_NAME
```

# The following will now declare variables locally which will overwrite comment variables

Declare new values locally to override the initial state.

```bash
export MY_RESOURCE_GROUP=RGSetLocally
export MY_VARIABLE_NAME=LocallySetVariable
```

Confirm the overrides are now visible to subsequent commands.

```bash
echo $MY_RESOURCE_GROUP
echo $MY_VARIABLE_NAME
```
