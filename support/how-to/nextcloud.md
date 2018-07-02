Guide to mount a nextcloud webdav access to docker container.

### 1. Installation of rclone 
Go to https://rclone.org/install/


### 2. Configure webdav remote
```
rclone config
  e/n/d/r/c/s/q> n
  name> nextcloud
  Storage> webdav
  url> https://your.nextcloud.com/remote.php/webdav/
  vendor> 1
  user> your_username
  y/g/n> y
  password: /* use an application password under securty of your account settings */
  password: /* repeat */
  y/e/d> y
  e/n/d/r/c/s/q> q

rclone listremotes
  nextcloud:
```
Full details : https://rclone.org/webdav/


### 3. Installation of plugin
```
docker plugin install sapk/plugin-rclone
  Plugin "sapk/plugin-rclone" is requesting the following privileges:
   - network: [host]
   - device: [/dev/fuse]
   - capabilities: [CAP_SYS_ADMIN]
  Do you grant the above permissions? [y/N] y   
```

### 4. Configure the volume
```
docker volume create --driver sapk/plugin-rclone --opt config="$(base64 ~/.config/rclone/rclone.conf)" --opt remote=nextcloud: --name nextcloud
```

### 4. Start the container and enjoy !
```
docker run -v nextcloud:/mnt --rm -ti ubuntu
```

