Each Receiver is run on Docker.

Specifications of each receiver are written on the beginning of the source code.

The source code are built in docker environment.

Logs are stored in the ~/logs directory of ec2 instance.
the name of the ec2 is rabbitmq-test(t3a.medium)

---
Variable Format
```
rabbitID: c2-$runMode
rabbitPW: c2-$runMode-pw
rabbitmqAddr: domainof ec2
rabbitmqPort:5672
```
```
ExchangeName: $subs-$runMode-exchange
queueName: $subs-$runMode-queue
routingKey: $subs-$runMode-key
```
```
logPath: logs/$subs-$runMode-logs/
logFile: $subs-$runMode-<date(format 2006-01-02)>-logs.txt
logFilePath: logs/$subs-$runMode-logs/$subs-$runMode-<date(format 2006-01-02)>.txt
```


