## ppjudge
- Judge systems for popcon-sc

# How to run
```
docker run --env PP_PPROF=1 -v /tmp/pj:/tmp/pj -v /sys/fs/cgroup/:/sys/fs/cgroup -v /tmp/lang.json:/root/lang.json -v /var/run/docker.sock:/var/run/docker.sock -ti --privileged --rm ppjudge --server http://{{ppjc address}} -debug -auth test -lang path/to/lang.json -parallel 3
```
