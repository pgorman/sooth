Sooth
========================================================================

Sooth checks the availability of hosts.
Sooth depends on the system `ping` command, and that the output of `ping` is unix-like:

```
ping -c 3 10.0.0.2
PING 10.0.0.2 (10.0.0.2) 56(84) bytes of data.
64 bytes from 10.0.0.2: icmp_seq=1 ttl=64 time=0.367 ms
64 bytes from 10.0.0.2: icmp_seq=2 ttl=64 time=0.367 ms
64 bytes from 10.0.0.2: icmp_seq=3 ttl=64 time=0.379 ms

--- 10.0.0.2 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2049ms
rtt min/avg/max/mdev = 0.367/0.371/0.379/0.005 ms
```

Sooth has these command-line options:

- `-d` turns on debugging console output.
- `-c` specifies the configuration file.

Sooth tries to be a calm/quiet program, emitting output only to report problems.
Tune the configuration values if Sooth's definition of "problem" doesn't match yours.


Configuration
------------------------------------------------------------------------

Sooth uses JSON for its configuration file, like:

```
{
	"debug": false,
	"web": {
		"ip": "127.0.0.1",
		"port": "9444"
	},
	"ping": {
		"checkInterval": 55,
		"packetCount": 5,
		"packetInterval": "1.0",
		"historyLength": 100,
		"lossReportRE": "^\\d+ packets transmitted, (\\d+) .+ (\\d+)% packet loss.*",
		"rttReportRE": "^r.+ (\\d+\\.\\d+)/(\\d+\\.\\d+)/(\\d+\\.\\d+)/(\\d+\\.\\d+) ms$"
	},
	"targets": [
		{ "name": "gateway", "address": "10.0.0.1" },
		{ "name": "DNS server", "address": "10.0.0.2" },
		{ "name": "web", "address": "10.0.0.11" }
	]
}
```

The "ping" section of the file demands the most explanation.
For each target/host that Sooth monitors, it starts a concurrent thread (goroutine) to ping the host with several packets, then sleep for a time.

- **checkInterval** sets the time (in seconds) each thread sleeps between sending a series of ping packets. The value must be an unquoted integer Number.
- **packetCount** sets the number of ICMP packets sent during each ping. The value must be an unquoted integer.
- **packetInterval** sets the delay (in milliseconds) between sending each packet. On many systems, only root may set this lower than "1.0".
- **historyLength** sets the number of check results to keep per target. The value must be an unquoted integer Number.
- **lossReportRE** defines the regular expression (with backslashes escaped to keep valid JSON) used to match the output line of the system `ping` command containing the summary of lost packets. The default regular expression should work on at least Linux and OpenBSD.
- **rttReportRE** defines the regular expression (with backslashes escaped to keep valid JSON) used to match the output line of the system `ping` command containing the round-trip time summary. The default regular expression should work on at least Linux and OpenBSD.


License (2-clause BSD)
------------------------------------------------------------------------

Copyright 2018 Paul Gorman

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
