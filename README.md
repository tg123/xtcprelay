# xtcprelay

xtcprelay is a framework which can take anything, literally, as a relayer of a tcp connection.

lets take mail based floppy data transmission as an example:

 ### Normal TCP Connection
```


.--.                      .--.
|__| .-------.            |__| .-------.
|=.| |.-----.|            |=.| |.-----.|
|--| ||  C  ||   <====>   |--| ||  S  ||
|  | |'-----'|            |  | |'-----'|
|__|~')_____('            |__|~')_____('
```

### xtcprelay with Mail based floppy driver

```
.--.                                                     
|__| .-------.                                           
|=.| |.-----.|         .-------------.           ______  
|--| ||  C  ||  <====> | Client Side | <====>   | |__| | 
|  | |'-----'|         |  xtcprelay  |          |  ()  | 
|__|~')_____('         |_____________|          |______| 
                                                    ^
                                                   | | 
                                                   | |
                                                    V
                                                 _________
                                               .`.        `.
                                              /   \ .======.\
                                              |   | |______||
      via  UPS/USPS/FedEx/....                |   |   _____ |
                                              |   |  /    / |
                                              |   | /____/  |
                                              | _ |         |
                                              |/ \|.-"```"-.|
                                              `` |||      |||

                                                    ^
                                                   | | 
                                                   | |
                                                   | |
.--.                                               | |     
|__| .-------.                                      V      
|=.| |.-----.|         .-------------.           ______  
|--| ||  S  ||  <====> | Server Side | <====>   | |__| | 
|  | |'-----'|         |  xtcprelay  |          |  ()  | 
|__|~')_____('         |_____________|          |______| 


```

This will be useful to bridge two networks with anything.


# Supported Drivers

## Azure Storage Queue
talk over azure storage queue

### Usage

 Bridge a tcp server listens on `:9000` as an example

 * Server Side
 
 ```
 ./xtcprelay -d azqueue --azqueue-account <YOUR ACC> --azqueue-key <YOUR KEY> serverside --azqueue-relayer-address test --azqueue-server-address 127.0.0.1:9000
 ```
 
 * Client Side
 
 ```
 ./xtcprelay -d azqueue --azqueue-account <YOUR ACC> --azqueue-key <YOUR KEY> clientside --azqueue-relayer-address test --azqueue-listen-address 0.0.0.0:9001
 ```
 
 Now on client side machine you will have a `9001` identical to remote `9000`. All traffic are relayed by your storage queue

##  Mail based floppy 

WIP

