# Docker CNI

This repo aims to integrate [CNI](https://github.com/containernetworking/cni) with Dockerd.

There is, [according to CNI repo](https://github.com/containernetworking/cni/blob/master/scripts/docker-run.sh), an approach to integrate by running a [pause] equivalent container ahead of the application container, but that's too pod-like for those who resent pod models.

Let's figure out yet another solution.

# Usage

## 1. Configure docker-cni

## 2. Configure dockerd
