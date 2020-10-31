# docker-mirror

This is a small utility to mirror docker images between repositories.

Please note this is beta quality so it might work for you or not. So far it works but will probably require more testing.

Prebuilt binaries will come at a future date.

## Prerequisites

Because this utility uses the [docker manifest](https://docs.docker.com/engine/reference/commandline/manifest/) command
you need to enable experimental features to the Docker CLI.
  
## Mirroring a repository

Simply compile this utility then:

    docker-mirror -d docker.example.com hello-world

Here we set the mirror repository as docker.example.com and ask it to mirror the hello-world image.

As long as you have permissions to upload to `docker.example.com` then it will pull down every image for all architectures and then push them to the new repository.

You can then access that image with:

    docker run -it --rm docker.example.com/library/hello-world

Note: the `library/` prefix is there because the `hello-world` image is in reality `library/hello-world`.
If you had another image, say `area51/jenkins` then you could mirror it with:

    docker-mirror -d docker.example.com area51/jenkins

and then pull it from your local repository as:

    docker pull docker.example.com/area51/jenkins

### Adding a prefix in the local repository

As my local Nexus3 repository has non-public images I prefer to keep the mirrored images with a mirror/ prefix.

This is simple to implement. Using the above examples:

    docker-mirror -d docker.example.com/mirror hello-world area51/jenkins

Then those two images are accessible in the local repository as `docker.example.com/mirror/library/hello-world` & `docker.example.com/mirror/area51/jenkins`

## Why this utility exists

On November 2 2020 Docker started enforcing a pull rate limit on the number of images you can pull from the main
central repository.

At the time of writing this limit was:
* 100 container image pulls every 6 hours for unauthenticated users
* 200  container image pulls every 6 hours for authenticated users

For most developers this isn't a problem but for serious developers who have CI builds running on multiple machines and
multiple platforms it's potentially an issue - depending on how many builds you are using it's possible that you will
hit this limit and have builds randomly failing.

What this utility does is it allows you to pull an image from the central repository and push it to a local one.
I already use [Sonatype Nexus 3](https://www.sonatype.com/nexus/repository-oss) locally for docker, Java/Maven, NodeJS
& APT repositories so it was a no-brainer to use it as the local mirror.

## Doesn't docker already support a local mirror

Yes it does, however it doesn't work. Nexus3 can support proxying the central repository,
however a bug with docker [#30880](https://github.com/moby/moby/issues/30880) causes it to fail for [non-root users](https://github.com/moby/moby/issues/30880#issuecomment-670150369).

## Isn't this just a pull, tag & push?

Yes & for simple images that would work. For example, you could mirror an image with:

    docker pull golang:alpine
    docker tag golang:alpine docker.example.org/golang:alpine
    docker push docker.example.org/golang:alpine

The downside is that would only push the image for the platform you run the commands on.
So if you had that image for both amd64 & arm architectures then those commands would only push amd64 if that was the
platform you run them under. 

What this utility does is it works with the manifests and would ensure that all platforms are mirrored.

Note: The utility deliberately limits the images mirrored with those who's OS = 'linux'.
I did this deliberately as I found that for hello-world it refused to pull the image with Windows as the os.
I also don't have a Windows instance so I cannot test against that Operating system.
