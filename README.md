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

Sooth uses JSON for its configuration file, like:

```
{
	"debug": false,
	"web": {
		"ip": "127.0.0.1",
		"port": "9444"
	},
	"ping": {
		"checkInterval": "60",
		"historyLength": "100",
		"packetCount": "5",
		"packetInterval": "0.3",
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

Sooth has these command-line options:

- `-d` turns on debugging console output.
- `-c` specifies the configuration file.


License (2-clause BSD)
------------------------------------------------------------------------

Copyright 2018 Paul Gorman

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
