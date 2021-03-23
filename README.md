# mysql-agent
You can use it to install mysql(5.7.32) and manage mysql instance(add/delete/modify/query)  
Only test on Centos7 Linux System

# begin
## package
```sh
sh package.sh
```
## start
```sh
setsid ./mysql-agent
```

# interface  
Install MySQL
```sh
curl -X POST https://127.0.0.1:30033/v1/mysqlagent/mysql/install -H "Content-Type:application/json" -k -s
```

UnInstall MySQL
```sh
curl -X POST https://127.0.0.1:30033/v1/mysqlagent/mysql/uninstall -H "Content-Type:application/json" -k -s
```

Add MySQL Instance
```sh
curl -X POST https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/add -H "Content-Type:application/json" -d '{"port": 3403, "serverId": 3, "dataDir": "/data/3403", "pwd": "Xxxyy@1234"}' -k -s
```

Delete MySQL Instance
```sh
curl -X DELETE https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/delete?port=3403 -H "Content-Type:application/json" -k -s
```

Modify MySQL Instance
```sh
curl -X POST https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/modify -H "Content-Type:application/json" -d '{"port": 3403, "serverId": 999, "dataDir": "/data/3403"}' -k -s
```

Modify MySQL Instance Pwd
```sh
curl -X POST https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/modifyowd -H "Content-Type:application/json" -d '{"port": 3403, "oldPwd": xxxx, "newPwd": "yyy"}' -k -s
```

Query MySQL Instance
```sh
curl -X GET https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/list -H "Content-Type:application/json" -k -s
```

Start MySQL Instance
```sh
curl -X GET https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/start -H "Content-Type:application/json" -d '{"port": 3403}' -k -s
```

Stop MySQL Instance
```sh
curl -X GET https://127.0.0.1:30033/v1/mysqlagent/mysql/instance/stop -H "Content-Type:application/json" -d '{"port": 3403}' -k -s
```