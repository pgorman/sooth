Sooth
========================================

Sooth is a simple network monitoring tool that tracks how a collection of hosts respond to pings over time.

Sooth is quiet by default, until it detects trouble.

In interactive mode, enter `?` for help.

To show a summary for all hosts, enter `a`.

```
a
172.20.0.10               9/20      55% loss    9ms avg rtt   12ms mdev
www.example.net           0/20     100% loss    0ms avg rtt    0ms mdev
google.com               20/20       0% loss   18ms avg rtt    3ms mdev
172.20.0.20              20/20       0% loss  118ms avg rtt  147ms mdev
…
```

To show the latest problems, press ENTER.

```

172.20.0.20           Jitter                          Feb 17 22:09:09
    RTT    min 1 ms    avg 96 ms    max 349 ms    stddev 134 ms
    seq (ms)   0=1    1=87   2=1    3=1    4=1    5=349  6=1    7=307  8=224  9=1  
    Since Feb 17 22:05:55     100/100 pkts    0% loss    112ms avg rtt    147ms mdev
www.example.net       Packet Loss                     Feb 17 22:09:10
    100% loss    (0/10 replies received)
    Since Feb 17 22:05:55       0/100 pkts  100% loss      0ms avg rtt      0ms mdev
172.20.0.10           Packet Loss                     Feb 17 22:09:08
    30% loss    (7/10 replies received)
    RTT    min 1 ms    avg 10 ms    max 20 ms    stddev 7 ms
    seq (ms)   0=__   1=__   2=1    3=9    4=11   5=14   6=1    7=14   8=20   9=__ 
    Since Feb 17 22:05:55      74/100 pkts   26% loss      9ms avg rtt      9ms mdev

```

The list of hosts to monitor can be specified as command line arguments or in a file specified with the `-f` flag.
In the file, list one IP address or host name per line:

```
10.0.0.1
# A comment.
ns1.example.com
192.168.10.100
www.example.net
mail.example.com
```

For list of command line options:

```
$ sooth --help
```

With the default settings, for each host, Sooth:

- sends a sequence of ten pings (taking about ten seconds)
- sleeps fifty seconds (10 + 50 = checking each host every minute)
- keeps a history of sixty sets of pings (host health over the past hour)

Sooth pings concurrently, so all hosts get checked every minute.

On Linux, to allow unprivileged pings, run:

```
# sysctl -w net.ipv4.ping_group_range="0   2147483647"
```

…or run Sooth with `sudo`.

On Windows, you may need to throw the `-raw` flag.

Note Sooth uses Go-style flags, so `-raw` is the same as `--raw`, and `-q -v -f file.txt` works but `-qvf file.txt` does not.

Sooth depends on the package https://github.com/sparrc/go-ping.


Paul Gorman
February 2020


License
----------------------------------------

Sooth monitors ping responses from a group of network hosts.
Copyright (C) 2020 Paul Gorman

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
