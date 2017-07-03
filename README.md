## Checks disks and finds large files

I got tired of writing convoluted shell scripts for same thing, it either returns the equivalent of `df -h` without arguments. Or if you give it paths, it will find large files within some threshold.


```
> ./diskchecker -t 10M flink-1.3.1

path: flink-1.3.1, Size: 155.75 MB
flink-1.3.1/lib/flink-dist_2.11-1.3.1.jar in flink-1.3.1/lib, Size: 68.75 MB => 65.94%
flink-1.3.1/lib/flink-shaded-hadoop2-uber-1.3.1.jar in flink-1.3.1/lib, Size: 34.94 MB => 33.51%
flink-1.3.1/opt/flink-ml_2.11-1.3.1.jar in flink-1.3.1/opt, Size: 25.94 MB => 60.25%
flink-1.3.1/opt/flink-table_2.11-1.3.1.jar in flink-1.3.1/opt, Size: 12.37 MB => 28.73%
```


This is returning all files that are over 10M and the corresponding directory


With no args, it will return the disk usage

```
./diskchecker
/ All: 7.74 GB Used: 6.51 GB Free: 1.23 GB
/opt All: 19.99 GB Used: 6.37 GB Free: 13.62 GB
```

You can filter further by supplying a regular expression as `-f`, say if you only want to match files ending in `.log`, this is what you would run

```
./diskchecker -t 10M -f 'jar$' /data
```
