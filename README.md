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

## Quick description

DeepSentinel is a low-level, server-agent software that monitors crucial machine and service states in a scalable and concurrent way that can process incoming data in real-time, track the health status of various services on multiple machines, and trigger alerts based on the absence of expected signals within given time intervals.  

DeepSentinel is thought and coded to fit very niche use-cases where a reliability risk affects your monitoring and alerting systems ; don't expect it to replace your Datadog or Grafana cloud instances!  

## Key Features

* `server` runs fully in-ram with a low ressource footprint
    - Can be hosted on a high SLA serverless provider
    - Doesn't need disk access
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
docker pull ghcr.io/equals215/deepsentinel-server:latest
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
With `wget` :
```bash
wget https://github.com/equals215/deepsentinel/releases/download/v0.0.2-untested/deepsentinel-server-linux-amd64 -o deepsentinel-server
chmod +x deepsentinel-server
```

With `curl` :
```bash
curl -o deepsentinel-server https://github.com/equals215/deepsentinel/releases/download/v0.0.2-untested/deepsentinel-server-linux-amd64
chmod +x deepsentinel-server
```

2. Now you have to configure the server, you can either do it :  
**using a config file (preferred method)**
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
  "low-alert-provider": "pagerduty",
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
**using the flags with the `run` command and `^C` the running program**
```go
./deepsentinel-server run --help
Run the API server

Usage:
  deepsentinel-server run [flags]

Flags:
      --address string                     Listening address
                                           Environment variable: DEEPSENTINEL_ADDRESS
                                           (default "0.0.0.0")
      --alertLow-to-alertHigh int          Number of alertLow event before alerting high
                                           Environment variable: DEEPSENTINEL_ALERT_LOW_TO_ALERT_HIGH
                                           (default 30)
      --degraded-to-failed int             Number of degraded event before considering a probe or service as failed
                                           Environment variable: DEEPSENTINEL_DEGRADED_TO_FAILED
                                           (default 10)
      --failed-to-alertLow int             Number of failed event before alerting low
                                           Environment variable: DEEPSENTINEL_FAILED_TO_ALERT_LOW
                                           (default 20)
      --high-alert-provider string         High alert provider name
                                           Environment variable: DEEPSENTINEL_HIGH_ALERT_PROVIDER

      --logging-level string               Logging level
                                           Environment variable: DEEPSENTINEL_LOGGING_LEVEL
                                           (default "info")
      --low-alert-provider string          Low alert provider name
                                           Environment variable: DEEPSENTINEL_LOW_ALERT_PROVIDER

      --no-alert                           Disable alerting
      --pagerduty.api-key string           PagerDuty API key
                                           Environment variable: DEEPSENTINEL_PAGERDUTY_API_KEY

      --pagerduty.integration-key string   PagerDuty integration key
                                           Environment variable: DEEPSENTINEL_PAGERDUTY_INTEGRATION_KEY

      --pagerduty.integration-url string   PagerDuty integration URL
                                           Environment variable: DEEPSENTINEL_PAGERDUTY_INTEGRATION_URL

      --port string                        Listening port
                                           Environment variable: DEEPSENTINEL_PORT
                                           (default "5000")
      --probe-inactivity-delay string      Delay before considering a probe inactive
                                           Environment variable: DEEPSENTINEL_PROBE_INACTIVITY_DELAY
                                           (default "2s")
```
3. Now that you or the server generated then configuration you can daemonize it if you system supports `systemd` or `launchd` :
```bash
./deepsentinel-server daemon install
```

4. The server should now be able to accept incoming connections. Remember to grab the auth token from either the logs or the config file at `/etc/deepsentinel/server-config.json` — you will need it to configure agents.

### Install Agent

As the agent is supposed to be run as close to the system as possible, it's not a good practice to run it inside a Docker container, hence why there is not Docker container for it 🤠  
1. Get the agent binary that matches your system under the [Release section](https://github.com/equals215/deepsentinel/releases).  
With `wget` :
```bash
wget https://github.com/equals215/deepsentinel/releases/download/v0.0.2-untested/deepsentinel-agent-linux-amd64 -o deepsentinel-agent
chmod +x deepsentinel-agent
```

With `curl` :
```bash
curl -o deepsentinel-agent https://github.com/equals215/deepsentinel/releases/download/v0.0.2-untested/deepsentinel-agent-linux-amd64
chmod +x deepsentinel-agent
```
2. Now you have to configure the agent, you can either do it :  
**using the `config` command (preferred method) :**
```go
./deepsentinel-agent config --help
Configure the running agent

Usage:
  deepsentinel config [flags]
  deepsentinel config [command]

Available Commands:
  auth-token     Set the authentication token
  machine-name   Set the machine name
  server-address Set the server address

Flags:
      --logging-level string   Logging level
                               Environment variable: DEEPSENTINEL_LOGGING_LEVEL
                               (default "info")
```
For example :
```bash
./deepsentinel-agent config auth-token totofoobar123456 && \
./deepsentinel-agent config machine-name buzz-lightyear && \
./deepsentinel-agent config server-address https://google.com:5000
```
**using a config file :**
```bash
mkdir -p /etc/deepsentinel/
nano /etc/deepsentinel/agent-config.json
```
```json
{
  "auth-token": "totofoobar123456",
  "logging-level": "info",
  "machine-name": "buzz-lightyear",
  "server-address": "https://google.com:5000"
}
```
3. Now that the agent is properly configured, it's time to daemonize it if you system supports `systemd` or `launchd` :
```bash
./deepsentinel-agent daemon install
```

4. Your agent should now be sending alive signals to the server. Check the server's logs to ensure that everything is setup properly.
#### Via fly.io
>TBD

## Credits
> TBD

## License

[GNU-GPLv3](https://github.com/equals215/deepsentinel/blob/8d8f70623c8725c2596ee5181d37ebdcf14ee81d/LICENSE)

---

> GitHub [@equals215](https://github.com/equals215) &nbsp;&middot;&nbsp;
> LinkedIn [Thomas Foubert](https://www.linkedin.com/in/thomas-f-devops/)
