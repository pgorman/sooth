Sooth
========================================================================

Sooth checks the availability of hosts.
Sooth depends on the system `ping` command, and that the output of `ping` is unix-like.

Sooth provides both console output and a JSON web API.

Sooth's automatic console alerting output looks like this:

```
Mar 26 16:48:29 fergus rtt min/avg/max/mdev = 0.212/1.147/9.278/2.710 ms
                ↳ fergus 540/540 0% loss, 0.55 ms avg, 0.04 ms mdev
Mar 26 16:49:32 bt 10 packets transmitted, 8 received, 20% packet loss, time 9010ms
                ↳ bt 546/550 1% loss, 28.70 ms avg, 7.51 ms mdev
Mar 26 16:51:58 xg-router 10 packets transmitted, 0 received, 100% packet loss, time 8998ms
                ↳ xg-router 523/550 2% loss, 47.55 ms avg, 7.17 ms mdev
                Last reply 3m ago.
Mar 26 16:55:01 hr 10 packets transmitted, 8 received, 20% packet loss, time 9001ms
                ↳ hr 564/580 3% loss, 37.73 ms avg, 5.58 ms mdev
```

Sooth's user-triggered console summary report looks like this:

```
s9             109/110      1% loss    32.11 ms avg,     9.34 ms mdev
s9-iad         109/110      1% loss    42.22 ms avg,     8.69 ms mdev
s9-router       98/100      2% loss    48.05 ms avg,     8.64 ms mdev
scc            110/110      0% loss    28.10 ms avg,     4.66 ms mdev
scc-iad        108/110      2% loss    39.32 ms avg,     8.03 ms mdev
scc-router     109/110      1% loss    48.73 ms avg,     7.08 ms mdev
storage        110/110      0% loss     0.41 ms avg,     0.16 ms mdev
xv             108/110      2% loss    51.29 ms avg,     8.20 ms mdev
```

Sooth tries to be a calm/quiet program, emitting output only to report problems.
Tune the configuration values if Sooth's definition of "problem" doesn't match yours.

Sooth has these command-line options:

- `-v` turns on verbose console output.
- `-c` specifies the configuration file.


Configuration
------------------------------------------------------------------------

Unless `-c` flag specifies otherwise, Sooth reads its configuration from `${XDG_CONFIG_HOME}/sooth.conf` (e.g., `~/.conf/sooth.conf`)
Sooth uses JSON for the configuration file format, like:

```
{
	"verbose": false,
	"web": {
		"ip": "127.0.0.1",
		"port": "9444"
	},
	"ping": {
		"checkInterval": 55,
		"packetCount": 5,
		"packetInterval": 1.0,
		"historyLength": 100,
		"jitterMiltiple": 2.0,
		"packetThreshold": 1,
		"lossReportRE": "^\\d+ packets transmitted, (\\d+) .+ (\\d+)% packet loss.*",
		"rttReportRE": "^r.+ (\\d+\\.\\d+)/(\\d+\\.\\d+)/(\\d+\\.\\d+)/(\\d+\\.\\d+) ms$"
	},
	"targets": [
		"10.0.0.1",
		"10.0.0.2",
		"example.com",
		"10.0.0.11"
	]
}
```

The "ping" section of the file demands the most explanation.
For each target/host that Sooth monitors, it starts a concurrent thread (goroutine) to ping the host with several packets, then sleep for a time.

- **checkInterval** sets the time (in seconds) each thread sleeps between sending a series of ping packets. The value must be an unquoted integer Number.
- **packetCount** sets the number of ICMP packets sent during each ping. The value must be an unquoted integer.
- **packetInterval** sets the delay (in milliseconds) between sending each packet. On many systems, only root may set this lower than 1.0.
- **historyLength** sets the number of check results to keep per target. The value must be an unquoted integer Number.
- **jitterMultiple** sets how large the round-trip time deviation must be as a multiple of the average round-trip time before Sooth prints an alert. E.g., with a jitterMultiple of 2.0, Sooth alerts on a ping response with an average RTT of 40 ms if the exceeds 80 ms.
- **packetThreshold** sets how many packets of a ping response must be lost before Sooth prints a warning. E.g., with a packetThreshold of 1, Sooth remains silent unless more than one packet is lost.
- **lossReportRE** defines the regular expression (with backslashes escaped to keep valid JSON) used to match the output line of the system `ping` command containing the summary of lost packets. The default regular expression should work on at least Linux and OpenBSD.
- **rttReportRE** defines the regular expression (with backslashes escaped to keep valid JSON) used to match the output line of the system `ping` command containing the round-trip time summary. The default regular expression should work on at least Linux and OpenBSD.


Console Interface
------------------------------------------------------------------------

Sooth prints alerts to the console.
Or, press ENTER to see a list of summary statistics for all targets.


Web API
------------------------------------------------------------------------

The web API provides JSON data at:

- http://127.0.0.1:9444/api/v1/conf
- http://127.0.0.1:9444/api/v1/history


License (2-clause BSD)
------------------------------------------------------------------------

Copyright 2018 Paul Gorman

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
