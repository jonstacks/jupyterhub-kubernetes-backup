# jupyterhub-kubernetes-backup

[![codecov](https://codecov.io/gh/jonstacks/jupyterhub-kubernetes-backup/branch/master/graph/badge.svg)](https://codecov.io/gh/jonstacks/jupyterhub-kubernetes-backup)

Backup PVCs created by jupyterhub to S3(the only backend supported right now).

<!-- TOC -->

- [jupyterhub-kubernetes-backup](#jupyterhub-kubernetes-backup)
    - [Overview](#overview)
    - [Testing](#testing)
    - [Install](#install)
    - [Deploying](#deploying)

<!-- /TOC -->

## Overview

This project will launch a kubernetes job which determines all of the jupyterhub
user PVCs in the given namespace. It does this by looking for PVCs which have
the `claim-` prefix. This was found to be reliable enough for our installation
of [zero-to-jupyterhub](https://zero-to-jupyterhub.readthedocs.io/en/latest/).
Once it knows which PVCs should be backed up, it creates a kubernetes job for
each of these which backs up the user's home directory to the backend of
choice(only S3 for now).

## Testing

`make test`

## Install

`make install`

## Deploying

The easiest way to deploy this right now is with the helm chart,located in
`chart/jupyterhub-kubernetes-backup`. At this time, no docker image is published
for this project, so you will need to publish your own after running: `make
docker-image`. You'll then need to push that image to your docker registry of
choice and reference it below in the `values.yaml` override. A minimal
installation will override the following in a custom `values.yaml`:


``` yaml
image:
  repository: <where you pushed the docker image>
  tag: <the tag you pushed to>

backend:
  type: s3
  s3:
    bucket: <bucket you want the files backed up to>
    prefix: <the prefix in the bucket you want to backup to>
    accessKey: <an accessKey that has access to write to the bucket>
    secret: <the secret for the accessKey>
    region: <the region the bucket is located in>

cronJob:
  schedule: <the schedule that you want to run it on>
```