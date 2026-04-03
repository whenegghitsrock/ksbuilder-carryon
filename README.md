# Introduction

ksbuilder is a CLI tool to create, publish, and manage KubeSphere extensions.

## Install

Download the [latest ksbuilder release](https://github.com/kubesphere/ksbuilder/releases) and then install it to `/usr/local/bin`:
```shell
tar xvzf ksbuilder_<version>_<arch>.tar.gz -C /usr/local/bin/
```

## Create your first KubeSphere extension

You can use `ksbuilder create` to create a KubeSphere extension interactively.

```
$ cd <project-directory>
$ ksbuilder create

Please input extension name: test
✔ ai-machine-learning
Please input extension author: ia
Please input Email (optional):
Please input author's URL (optional):
Directory: /path/test

The extension charts has been created.
```

The extension directory created looks like below (when both frontend and backend are enabled; frontend-only or backend-only omits the other chart and the matching `values.yaml` section):

```
.
├── README.md
├── README_zh.md
├── applicationclass.yaml
├── charts
│   ├── backend
│   │   ├── Chart.yaml
│   │   ├── templates
│   │   │   ├── NOTES.txt
│   │   │   ├── deployment.yaml
│   │   │   ├── extensions.yaml
│   │   │   ├── helps.tpl
│   │   │   ├── service.yaml
│   │   │   └── tests
│   │   │       └── test-connection.yaml
│   │   └── values.yaml
│   └── frontend
│       ├── Chart.yaml
│       ├── templates
│       │   ├── NOTES.txt
│       │   ├── deployment.yaml
│       │   ├── extensionentry.yaml
│       │   ├── helps.tpl
│       │   ├── jsbundle.yaml
│       │   ├── reverseproxy.yaml
│       │   ├── service.yaml
│       │   └── tests
│       │       └── test-connection.yaml
│       └── values.yaml
├── extension.yaml
├── frontend
│   ├── package.json
│   ├── yarn.lock
│   ├── configs
│   │   └── console.config.js
│   └── extensions
│       └── <extension-name>
│           ├── Dockerfile
│           ├── Makefile
│           ├── README.md
│           ├── package.json
│           └── src
│               └── ...
├── permissions.yaml
├── static
│   ├── favicon.svg
│   └── screenshots
│       └── screenshot.png
└── values.yaml
```

The `frontend/` directory is a KubeSphere Console-style extensions workspace (similar to [extension-samples `extensions-frontend`](https://github.com/kubesphere/extension-samples/tree/master/extensions-frontend)). Extension UI code lives under `frontend/extensions/<extension-name>/`. See that folder’s `README.md` for build and local development (`yarn install` at `frontend/`, `make build-frontend` from the extension root, and so on).

Then you can customize your extension like below:

- Specify the default backend and frontend images in `values.yaml`. Only the blocks for enabled capabilities are generated (for example, a frontend-only extension has no `backend:` section).

```
frontend:
  enabled: true
  image:
    repository: <YOUR_REPO>/<extension-name>
    tag: latest

backend:
  enabled: true
  image:
    repository: <YOUR_REPO>/<extension-name>
    tag: latest
```

- In `extension.yaml`, `installationMode` is `Multicluster` when both frontend and backend are present, and `HostOnly` for backend-only extensions.

- Add `APIService` definition to the backend `extensions.yaml`

```yaml
apiVersion: extensions.kubesphere.io/v1alpha1
kind: APIService
metadata:
  name: v1alpha1.<extension-name>.kubesphere.io
spec:
  group: <extension-name>.kubesphere.io
  version: v1alpha1                                      
  url: http://<extension-name>-backend.default.svc:8080
status:
  state: Available
```

- The frontend chart ships `jsbundle.yaml` (and related resources) for you. A typical `JSBundle` points at the service-served bundle, for example:

```yaml
apiVersion: extensions.kubesphere.io/v1alpha1
kind: JSBundle
metadata:
  name: <extension-name>
spec:
  rawFrom:
    url: http://<extension-name>-frontend.default.svc/dist/index.js
status:
  state: Available
  link: /dist/<extension-name>/index.js
```

## Publish/Unpublish your KubeSphere extension

You can publish/unpublish KubeSphere extension to KubeSphere cluster once it's ready:

```shell
ksbuilder publish/unpublish <extension-name>
```

## Push and submit your extension to KubeSphere Cloud

### Create API access token

1. Register an account on [KubeSphere Cloud](https://kubesphere.cloud).
2. Open [KubeSphere Marketplace](https://kubesphere.co/marketplace/), click on "Become a service provider," sign the agreement, and become an extension service provider (i.e., developer).
3. Open [https://kubesphere.cloud/user/setting/](https://kubesphere.cloud/user/setting/), click on "Security," then click "Create Token," check "Extension Component," and click "Generate." The generated token is the cloud API key, formatted like `kck-xxx`. Please keep it safe.

### Login to KubeSphere Cloud

Use the `ksbuilder login` subcommand to login to KubeSphere Cloud:

```
$ ksbuilder login
✔ Enter API token: ***

Login Succeeded
```

or:

```
$ ksbuilder login -t xxx

Login Succeeded
```

### Push and submit the extension

Use the `ksbuilder push` subcommand to submit the extension to KubeSphere Cloud. The `push` subcommand is similar to `publish` and the target can be either a directory or a packaged `.tgz` file:

```
$ ksbuilder push tower

$ ksbuilder push tower-1.0.0.tgz
```

> NOTE: We will upload static files such as icons and screenshots in the extension to the KubeSphere Cloud separately
and delete the static file directory in the original package to reduce the size of the entire chart.

### Check the extension status

After submitting the extension, it needs to be approved by an administrator before it can be listed on KubeSphere Marketplace. You can use the `ksbuilder get` or `ksbuilder list` subcommands to check the status of the extension:

```
$ ksbuilder list

ID                   NAME                STATUS              LATEST VERSION
469804312491467933   devops              ready               1.0.0
482307830796264605   kubeblocks          ready               0.6.3
```

```
$ ksbuilder get tower

Name:     tower
ID:       515518094316151974
Status:   draft

SNAPSHOT ID          VERSION   STATUS      UPDATE TIME
515518094316217510   1.0.0     submitted   2024-05-27 09:37:05
```

### Unpush a snapshot

Use the `ksbuilder unpush` subcommand to cancel the submission of a snapshot in KubeSphere Cloud.

When we submit an extension, its snapshot status is marked as `submitted`. If we want to make some new changes to the extension at this point, we can use this command to undo the submission. After undoing, the status of the extension will revert to `draft`, allowing it to be pushed again.

To undo the submission, we first need to use the `get` subcommand to obtain the corresponding snapshot ID:

```
$ ksbuilder get tower
Name:     tower
ID:       515518094316151974
Status:   draft

SNAPSHOT ID          VERSION   STATUS      UPDATE TIME
515518094316217510   1.0.0     submitted   2024-05-27 17:37:05 CST

$ ksbuilder unpush 515518094316217510
unpush snapshot 515518094316217510
Snapshot 515518094316217510 has been unsubmitted and reverted to draft state
```

Please refer to [KubeSphere extension development guide](https://dev-guide.kubesphere.io/extension-dev-guide/en/packaging-and-release/) for more details on extension packaging and releasing.
