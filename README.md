# git-exfiltrate

## Info

To be run from a Windows machine


## Comments
When it starts listening, it creates a file called We Up. Then it reads instructions from a file called instructions. Either win command instructions or requests to exfil data.

It's also currently fileless. The repo is being hosted in ram and so for exfil requests, they are handled in 1024 byte chunks and uploaded 1 by 1.

I also append random chars to the end of file names in the event that 2 files have the same name of commands are run seperately
