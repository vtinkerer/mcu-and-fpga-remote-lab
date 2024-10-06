#!/usr/bin/expect -f

# Set variables
set timeout -1
set source_dir "/home/vladyslav/phd/digitrans/backend/digitrans-lab-go"
set dest "pi@192.168.1.248:/home/pi"
set password "raspberry"

# Spawn rsync command
spawn rsync -avz $source_dir $dest

# Expect password prompt and send password
expect "password:" {
    send "$password\r"
}

# Wait for rsync to finish
expect eof