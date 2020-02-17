Sooth
========================================

Sooth is a simple network monitoring tool that tracks how a collection of hosts respond to pings over time.

Sooth is quiet by default, until it detects trouble.

To show a summary for all hosts, enter `a`.

```
a
172.20.0.10               9/20      55% loss    9ms avg rtt   12ms mdev
www.example.net           0/20     100% loss     0s avg rtt     0s mdev
google.com               20/20       0% loss   18ms avg rtt    3ms mdev
172.20.0.20              20/20       0% loss  118ms avg rtt  147ms mdev
…
```

To show the latest problems, press ENTER.

```

172.20.0.10           Packet Loss                     Feb 17 14:31:29
    30% loss (7/10 replies received)
    RTTs 0s min, 3ms avg, 14ms max, 5ms stddev
    seq:ms  0:0    1:__   2:__   3:0    4:__   5:0    6:0    7:2    8:13   9:1  
    History       29/50      42% loss    7ms avg rtt    8ms mdev
www.example.net  Packet Loss                     
    100% loss (0/10 replies received)
    History        0/50     100% loss     0s avg rtt     0s mdev
172.20.0.20           Latency, Jitter                 Feb 17 14:31:29
    RTTs 0s min, 166ms avg, 424ms max, 182ms stddev
    seq:ms  0:0    1:0    2:133  3:421  4:423  5:5    6:386  7:0    8:284  9:0  
    History       50/50       0% loss  134ms avg rtt  151ms mdev
```

The list of hosts to monitor can be specified as command line arguments or in a file specified with the `-f` flag.
In the file, list one IP address or host name per line:

```
10.0.0.1
ns1.example.com
192.168.10.100
www.example.net
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
