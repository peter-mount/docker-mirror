# docker-mirror

This is a small utility to mirror docker images between repositories.

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

