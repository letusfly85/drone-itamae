# drone-itamae
Drone plugin for publishing Docker images using itamae


## Docker

Build the Docker container:

```sh
docker build --rm=true -t plugins/drone-itamae .
```

Build and Publish a Docker container

```sh
docker run -i --privileged -v $(pwd):/drone/src plugins/drone-itamae <<EOF
{
	"clone": {
		"dir": "/drone/src"
	},
	"commit" : {
		"sha": "9f2849d5",
		"branch": "master"
	},
	"vargs": {
	        "itamae_target_image": "ubuntu:14.04",
		"recipes": ["localte.rb"],
		"username": "kevinbacon",
		"password": "pa$$word", 
		"email": "foo@bar.com", 
		"repo": "foo/bar",
		"tag": "001",
		"storage_driver": "aufs"
	}
}
EOF
```
