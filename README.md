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
  <a href="#download">Download</a> •
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
> TBD

## Download
> TBD

## Credits
> TBD

## License

[GNU-GPLv3](https://github.com/equals215/deepsentinel/blob/8d8f70623c8725c2596ee5181d37ebdcf14ee81d/LICENSE)

---

> GitHub [@equals215](https://github.com/equals215) &nbsp;&middot;&nbsp;
> LinkedIn [Thomas Foubert](https://www.linkedin.com/in/thomas-f-devops/)
