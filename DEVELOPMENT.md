# Developing Boeing

What follows is a rundown on different ways to run and develop Boeing, its UI and its tools locally.

## Running Boeing

The easiest way to run Boeing locally is to run `make dev`. This will launch three processes: the API server, admin UI, and user UI. Opening `http://localhost:8080/admin/` will launch the admin UI. Changing the UI code will update the UI automatically. Changing any of the Go code requires restarting the server.

## Building and Running the Boeing Docker Image

Boeing is ultimately packaged into an image for distribution. You can build said image with `docker build -t my-boeing .`, and then run the image via `docker run -p 8080:8080 my-boeing`.

## Debugging Boeing

It is possible to run the server and/or UIs in and IDE for debugging purposes. These steps layout what is necessary for JetBrains IDEs, but an equivalent process can be used with VSCode-based editors.

### Server

To run the server in GoLand:
1. Create a new "Go Build" configuration.
2. In the "Program Arguments" section, enter `server --dev-mode`.

Then you're ready to run or debug this target.

### User UI

To run the User UI in GoLand or WebStorm:
1. Create a new "npm" build.
2. In the "package.json" dropdown, select the `package.json` file in the `ui/user` directory.
3. In the "Command" dropdown, select `run`.
4. In the "Scripts" dropdown, select `dev`.
5. In the "Environment" section, enter `VITE_API_IN_BROWSER=true`.

Then you're ready to run or debug this target.

## Developing Boeing Providers

Boeing has a set of providers. These are in the repo `github.com/boeing-ai-gateway/providers`. By default, Boeing will pull the providers' configuration from this repo. However, when developing tools in this repo, you can follow these steps to use a local copy.

1. Clone `github.com/boeing-ai-gateway/providers` to your local machine.
2. In the root directory of the tools repo on your local machine, run `make build-images`.
3. Run the Boeing server, either with `make dev` or in your IDE, with the `BOEING_SERVER_PROVIDER_REGISTRIES` environment variable set to `<local-tools-fork-root-directory>`; e.g. If you cloned the tools repo to the directory "above" the Boeing repo, you'd use `BOEING_SERVER_PROVIDER_REGISTRIES='../providers' make dev`.

Now, any time one of these tools is run, your local copy will be used.

> [!IMPORTANT]
> Any time you change a Go based tool in your local repo, you must run `make build` in the tools repo for the changes to take effect with Boeing.

> [!NOTE]
> Provider definitions and metadata are only synced to Boeing every hour. Therefore, if you make a change to the provider on your local machine, it may not reflect immediately in Boeing.

## Boeing Server Dev Mode

In the description above for running the server in an IDE, the `--dev-mode` flag is used. This flag is also used when running the server with `make dev`. This does a few things, the most helpful of which is to give you access to the database via `kubectl`. The kubeconfig is located at `tools/devmode-kubeconfig`.

For example, from the root directory of the boeing repo, you can list all agents in your setup with `kubectl --kubeconfig tools/devmode-kubeconfig get agents`.

## Local Jaeger

Boeing already supports standard OpenTelemetry exporters. For local tracing with Jaeger:

1. Start Jaeger:
```bash
make otel-jaeger-up
```
2. Point Boeing at Jaeger before running `make dev` or starting the server in your IDE:
```bash
export OTEL_TRACES_EXPORTER=otlp
export OTEL_METRICS_EXPORTER=none
export OTEL_LOGS_EXPORTER=none
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_SERVICE_NAME=boeing
export OTEL_TRACES_SAMPLER=always_on
```
3. Open Jaeger at `http://localhost:16686`.

Jaeger also exposes OTLP gRPC on `localhost:4317` and OTLP HTTP on `localhost:4318`, so Boeingbot can be pointed at the same local instance.

Useful commands:

```bash
make otel-jaeger-up
make otel-jaeger-logs
make otel-jaeger-down
```

## Resetting

There may be times when you want to completely wipe your setup and start fresh. The location of data and caches is dependent on your system. For Mac or Linux, you can run the respective command in the root of the boeing repo on your local machine.

On Mac:
```bash
rm -rf ~/Library/Application\ Support/boeing &&
rm -rf ~/Library/Caches/boeing &&
rm boeing.db
```

On Linux:
```bash
rm -rf ~/.local/share/boeing &&
rm -rf ~/.cache/boeing &&
rm boeing.db
```

## Serving the Documentation

The documentation for Boeing is in the main repo. You can serve the documentation from your local machine by running `make serve-docs` in the root of the boeing repo.

## Other Configuration

Boeing is configured via environment variables. You can see the relevant environment variables by building the binary (as above) and running `./bin/boeing server --help`. There is also documentation available. You can serve the documentation locally as above.

## Running Boeing Locally with Kubernetes (Boeingbot Agents)

Boeingbot agent containers run in Kubernetes and need to reach your local Boeing process. This requires [Telepresence](https://www.telepresence.io/) to bridge the network between your Mac and the cluster.

### Prerequisites

- [Rancher Desktop](https://rancherdesktop.io/) (or another local Kubernetes setup)
- [Telepresence](https://www.telepresence.io/docs/install/) v2.x
- Local images loaded into containerd (see below)

### 1. Load local images into containerd

Rancher Desktop uses containerd, not Docker's image store. Load any locally built images with:

```bash
docker save boeingbot:local | nerdctl --address /var/run/docker/containerd/containerd.sock load
docker save boeingbot-agent:local | nerdctl --address /var/run/docker/containerd/containerd.sock load
```

### 2. Configure the cluster namespaces

The `boeing-mcp` namespace is where MCP server pods run. It must exist with PSA set to `privileged` (required for Telepresence's network init container):

```bash
kubectl create namespace boeing-mcp --dry-run=client -o yaml | kubectl apply -f -
kubectl label namespace boeing-mcp \
  pod-security.kubernetes.io/enforce=privileged \
  pod-security.kubernetes.io/audit=restricted \
  pod-security.kubernetes.io/warn=restricted \
  --overwrite
```

The `default` namespace also needs PSA set to `privileged` for Telepresence:

```bash
kubectl label namespace default \
  pod-security.kubernetes.io/enforce=privileged \
  pod-security.kubernetes.io/audit=restricted \
  pod-security.kubernetes.io/warn=restricted \
  --overwrite
```

### 3. Set up Telepresence and intercept

Use the Makefile target to create/update the intercept target, reconnect Telepresence, restart the target deployment, and create the intercept in one step:

```bash
make telepresence-setup
```

This target runs:

```bash
kubectl create deployment boeing-upstream --image=alpine --dry-run=client -o yaml -- sleep infinity | kubectl apply -f -
kubectl create service clusterip boeing-upstream --tcp=8080:8080 --dry-run=client -o yaml | kubectl apply -f -
kubectl patch svc boeing-upstream --type='json' -p='[{"op":"replace","path":"/spec/ports/0/name","value":"http"}]'
kubectl apply -f tools/boeing-proxy.yaml
telepresence quit -s
telepresence connect
kubectl rollout restart deployment/boeing-upstream
telepresence intercept boeing-upstream -p 8080:8080
```

`boeing-upstream` is the Telepresence intercept target — traffic to it routes to your local port 8080. `tools/boeing-proxy.yaml` deploys an nginx pod as the `boeing` Service (port 80). Pods reach Boeing at `http://boeing.default.svc.cluster.local` → nginx → `boeing-upstream` (Telepresence) → your local process. nginx rewrites `http://localhost:8080` → `http://boeing.default.svc.cluster.local` in response bodies so that OAuth metadata URLs are correct for pods, while your browser continues to use `http://localhost:8080` directly.

Verify the intercept is `ACTIVE` with `telepresence list`.

### 4. Configure Boeing environment variables

```bash
export BOEING_SERVER_MCPRUNTIME_BACKEND='k8s'
export BOEING_SERVER_SERVICE_NAME=boeing
export BOEING_SERVER_SERVICE_NAMESPACE=default

# optional if using locally-built Boeingbot images
export BOEING_SERVER_BOEINGBOT_AGENT_IMAGE='boeingbot-agent:local'
export BOEING_SERVER_MCPREMOTE_SHIM_BASE_IMAGE='boeingbot:local'
```

### Troubleshooting

- **`ImagePullBackOff`**: Image isn't in containerd — re-run `nerdctl load`.
- **`timed out waiting for MCP server to be ready: <url>`**: The URL in the error shows what Boeing is trying to reach. If it's `*.svc.kubernetes` instead of `*.svc.cluster.local`, check `BOEING_SERVER_MCPCLUSTER_DOMAIN`.
- **Telepresence `NO_AGENT`**: Pod was created before intercept — run `kubectl rollout restart deployment/boeing-upstream`.
- **PSA violations on `tel-agent-init`**: Namespace enforce level must be `privileged` (step 2 above).
- **Stale intercept conflict** (`conflict with intercept ... on port 8080`): A previous intercept is stuck in the Traffic Manager. Reset it with:
  ```bash
  kubectl delete pod -n ambassador -l app=traffic-manager
  kubectl rollout restart deployment/boeing-upstream
  telepresence connect --namespace default
  telepresence intercept boeing-upstream -p 8080:8080
  ```
