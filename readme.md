# MQTT-SN-SIP Gateway
The project is a gateway prototype for encapsulation and decapsulation of the MQTT protocol (MQTT-SN) into SIP packets. To be able to transfer MQTT data over SIP.

## How to run
Before starting, you need to configure the configuration file - mqttsip.conf:
- asterisk_ip, asterisk_port: ip address of SIP server;
- fromId: sip id of client / broker
- toId: sip id of client / broker
- srcIp: ip address of client
- mode: mqtt client / mqtt broker
- brokerIp: ip address of mqtt broker


To run, you need to install Go with go modules support. Then enter the command
```go run main.go```


For the module to work, you need to prepare:
- SIP server (Asterisk);
- MQTT broker;
- MQTT client;

  It is recommended to run on a server with a running SIP switch (Asterisk).

### The project is under development. The main development is carried out on a closed Git repository, intermediate results are presented here.