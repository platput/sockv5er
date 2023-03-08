# sockv5er
a tool to create ssh tunnels on demand using the free tier ec2 instances from AWS from any given region

# Setup
- Create a new AWS account and set up the access id and secret key
- Set the below env variables in the system.

```shell
ACCESS_KEY_ID="" # AWS Access key ID
SECRET_KEY="" # AWS Access key
SOCKS_V5_PORT=1337 # A free port on your system.
```
- Execute `sockv5er` to start the socksv5 server
- Add a socksv5 proxy in your browser with the address 127.0.0.1 and port you have as the value for `SOCKS_V5_PORT`
- Press CTRL + C to exit.
- To clean up all the resources, execute `sockv5er` again and press `Y`

# Features
- Creates an EC2 instance in the free tier and starts an SSH tunnel which can be used as socksv5 proxy
- Terminates the created EC2 instances in 20 minutes after it starts up in case the app crashes
- Tracks the resources the app creates so that it can be deleted in the subsequent run

# TODO
- Postpone the shutdown of the EC2 instance as long as the ssh tunnel is active
- Create the ec2 security group in such a way that, only the current systems public ip is allowed in the ingres 
- Handle the exit from the SSH tunnel in a graceful way
- Add more test cases
- Make the readme.md a bit more elaborate
- Add better log messages and print statements

