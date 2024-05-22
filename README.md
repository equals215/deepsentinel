<h1 align="center">
  <br>
  <a href="http://www.amitmerchant.com/electron-markdownify"><img src="https://gist.githubusercontent.com/equals215/4cc46fe3225e4def80c1e915a5608c8d/raw/1b16d0817e88d8d5fa8c0730a8bfa66e072484c9/deepsentinel-crop.svg" alt="Markdownify" width="700"></a>
  <br>
</h1>


<!-- <p align="center">
  <a href="https://linktr.ee/equals215">
    <img src="https://img.shields.io/badge/$-donate-ff69b4.svg?maxAge=2592000&amp;style=flat">
  </a>
</p> -->
<p align="center">
If you ever worried "What would happen if my monitoring/alerting systems fail?" then search no further, you're at the right place
</p>

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#how-to-use">How To Use</a> •
  <a href="#credits">Credits</a> •
  <a href="#license">License</a>
</p>
<p align="center">
<img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/equals215/deepsentinel/.github%2Fworkflows%2Fgo-build-and-release.yml">
<a href="https://codecov.io/gh/equals215/deepsentinel" > 
 <img src="https://codecov.io/gh/equals215/deepsentinel/branch/master/graph/badge.svg?token=6JKK7IP4VO"/> 
</a>
<a href="https://goreportcard.com/report/github.com/equals215/deepsentinel">
  <img src="https://goreportcard.com/badge/github.com/equals215/deepsentinel" />
</a>
<a href="https://codeclimate.com/github/equals215/deepsentinel/maintainability">
  <img src="https://api.codeclimate.com/v1/badges/1058b0dd522c52babff0/maintainability" />
</a>
<img src="https://snyk.io/test/github/equals215/deepsentinel/badge.svg" />
</p>

## Quick description

DeepSentinel is a low-level, server-agent software that monitors crucial machine and service states in a scalable and concurrent way that can process incoming data in real-time, track the health status of various services on multiple machines, and trigger alerts based on the absence of expected signals within given time intervals.  

DeepSentinel is thought and coded to fit very niche use-cases where a reliability risk affects your monitoring and alerting systems ; don't expect it to replace your Datadog or Grafana cloud instances!  

## Key Features

* `server` runs fully in-ram with a low ressource footprint
    - Can be hosted on a high SLA serverless provider
    - Doesn't need disk access (but appreciate it to persist the auth-token)
* `agent` pushes simple JSON payloads via HTTP/S as an alive signal
* Both `server` and `agent` are monitoring themselves for any fatal error
* `agent` daemonize itself and runs no matter what
    - You only need to configure `server` address, machine name and auth token
    - No actions required on the server-side
* `agent` daemon is live configurable and live unregisterable

## How To Use
### Install Server
#### Via Docker
On any instance that has Docker installed you can run the following commands :

```bash
docker run \
    -v /your/host/path:/etc/deepsentinel \
    -p <host_port>:5000 \
    -e DEEPSENTINEL_ADDRESS=0.0.0.0 \
    -e DEEPSENTINEL_PORT=5000 \
    -e DEEPSENTINEL_PROBE_INACTIVITY_DELAY=10s \
    -e DEEPSENTINEL_DEGRADED_TO_FAILED=20 \
    -e DEEPSENTINEL_FAILED_TO_ALERT_LOW=30 \
    -e DEEPSENTINEL_ALERT_LOW_TO_ALERT_HIGH=50 \
    -e DEEPSENTINEL_LOGGING_LEVEL=info \
    -e DEEPSENTINEL_LOW_ALERT_PROVIDER=pagerduty \
    -e DEEPSENTINEL_HIGH_ALERT_PROVIDER=pagerduty \
    -e DEEPSENTINEL_PAGERDUTY_API_KEY=<pd_api_key> \
    -e DEEPSENTINEL_PAGERDUTY_INTEGRATION_KEY=<pd_integration_key> \
    -e DEEPSENTINEL_PAGERDUTY_INTEGRATION_URL=<pd_integration_url> \
    ghcr.io/equals215/deepsentinel-server:latest

```

Replace latest by any tag you want to run. Docker Image tag == Repo tag.  
Adapt environment variables based on the config you want to apply.  

**You will have to grab the auth token from either the logs or the config file bind mounted to your host**

#### Via binaries
1. Get the server binary that matches your system under the [Release section](https://github.com/equals215/deepsentinel/releases).  
- With `wget` :

```bash
wget https://github.com/equals215/deepsentinel/releases/download/v0.0.3-untested/deepsentinel-server-linux-amd64 -O deepsentinel-server && \
chmod +x deepsentinel-server
```

- With `curl` :
```bash
curl -o deepsentinel-server https://github.com/equals215/deepsentinel/releases/download/v0.0.3-untested/deepsentinel-server-linux-amd64 && \
chmod +x deepsentinel-server
```

2. Now you have to configure the server :  

Create the config file :
```bash
mkdir -p /etc/deepsentinel/
nano /etc/deepsentinel/server-config.json
```
```json
{
  "address": "0.0.0.0",
  "alertlow-to-alerthigh": 30,
  "degraded-to-failed": 10,
  "failed-to-alertlow": 20,
  "high-alert-provider": "pagerduty",
  "logging-level": "info",
  "low-alert-provider": "",
  "no-alert": false,
  "pagerduty": {
    "api-key": "...",
    "integration-key": "...",
    "integration-url": "..."
  },
  "port": "5000",
  "probe-inactivity-delay": "5s"
}
```

3. Now that you generated the configuration you can daemonize it if your system supports `systemd` or `launchd` :
```bash
./deepsentinel-server daemon install
```

4. The server should now be able to accept incoming connections. Remember to grab the auth token from either the logs or the config file at `/etc/deepsentinel/server-config.json` — you will need it to configure agents.

### Install Agent

As the agent is supposed to be run as close to the system as possible, it's not a good practice to run it inside a Docker container, hence why there is not Docker container for it 🤠  
1. Get the agent binary that matches your system under the [Release section](https://github.com/equals215/deepsentinel/releases).  

- With `wget` :
```bash
wget https://github.com/equals215/deepsentinel/releases/download/v0.0.3-untested/deepsentinel-agent-linux-amd64 -O deepsentinel-agent && \
chmod +x deepsentinel-agent
```

- With `curl` :
```bash
curl -o deepsentinel-agent https://github.com/equals215/deepsentinel/releases/download/v0.0.3-untested/deepsentinel-agent-linux-amd64 && \
chmod +x deepsentinel-agent
```

2. Now you have to configure the agent using the `install` command :  

The following command will ask you for `auth-token`, `server-address` and `machine-name` in order to set up properly. It will then generate it's config file located in `/etc/deepsentinel/agent-config.json` and daemonize itself if your system supports `systemd` or `launchd`

```bash
sudo ./deepsentinel-agent install
```

3. Your agent should now be sending alive signals to the server. Check the server's logs to ensure that everything is setup properly.  

## Credits
> TBD

## License

[GNU-GPLv3](https://github.com/equals215/deepsentinel/blob/8d8f70623c8725c2596ee5181d37ebdcf14ee81d/LICENSE)

---

> GitHub [@equals215](https://github.com/equals215) &nbsp;&middot;&nbsp;
> LinkedIn [Thomas Foubert](https://www.linkedin.com/in/thomas-f-devops/)
