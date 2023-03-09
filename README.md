# 🕳️ sockv5er
a tool to create ssh tunnels on demand using the free tier ec2 instances from AWS from any given region

## What does it do:
- Creates a security group with port 22 open to the world
- Creates a private key which will be used to connect to the ec2 instance
- Creates an ec2 instance which will shut down in 20 minutes if nothing is done.
- Connects to the ec2 instance and starts a socksv5 proxy using ssh tunnel
- Once you add the proxy settings for 127.0.01:1337 all your browser traffic will be encrypted and transferred through the tunnel between your system and the ec2 instance.
- Cleans up the ec2 instance after usage.
- Subsequent runs of this app will give you option to clean up any existing resources created by this app.

# ⚙️ Setup 
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

# 🎊 Features
- Creates an EC2 instance in the free tier and starts an SSH tunnel which can be used as socksv5 proxy
- Terminates the created EC2 instances in 20 minutes after it starts up in case the app crashes
- Tracks the resources the app creates so that it can be deleted in the subsequent run

# 📝 TODO
- Fix the existing test cases and add more coverage
- Postpone the shutdown of the EC2 instance as long as the ssh tunnel is active 
- Handle the exit from the SSH tunnel in a graceful way
- Make the readme.md a bit more elaborate
- Add better log messages and print statements

# ✏️ Contribute
All contributions are welcome. So raise away your PRs. Here's the [contributor guidelines](https://github.com/platput/sockv5er/blob/main/CONTRIBUTING.md).
