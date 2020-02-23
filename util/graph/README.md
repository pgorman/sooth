Graph Sooth Results
========================================

Sooth detects packet loss, latency, and jitter.
With the `-w` (wide) flag, Sooth reports each problem on one line, making a log of those reports easy to process with a simple pipeline filter.


```
🐚 ~ $ sooth -w 172.20.0.10 google.com 172.20.0.20 > ~/tmp/sooth.log
```

This filter, `graph`, produces an ASCII graph that answers the question "during what hours do problems most frequently occur?"
In some environment, this can suggest whether network problems are due to high utilization/saturation.
