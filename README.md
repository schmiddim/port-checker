# WIP

Check if ports are open on hosts.




Pass probes as flag
```
 ./port-checker -probes=127.0.0.1:80;tcp;1,127.0.0.1:443;tcp;1,127.0.0.1:8990;tcp;1
```

Pass probes in config file
```
./port-checker -prometheusServerPort=9102 -configFile=example.yaml
```
